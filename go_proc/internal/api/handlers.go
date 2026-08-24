package api

/*
Written by AI
*/

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"img_proc/go_proc/internal/archive"
	"img_proc/go_proc/internal/invoker"
	"img_proc/go_proc/internal/pool"
	"img_proc/go_proc/internal/results"
)

// Settings built once at startup in cmd/ingest/main.go.
type Server struct {
	ImgRoutines int
	ResultsDir  string
	Invoker     *invoker.Client
}

// What Rails expects back.
type createResponse struct {
	JobID string `json:"job_id"`
}

// Zip uploads can be large, so cap what multipart keeps in RAM.
// Anything past this spills to disk on its own.
const maxMemory int64 = 32 << 20 // 32 MB

func (s *Server) CreateJob(w http.ResponseWriter, r *http.Request) {
	var err error = r.ParseMultipartForm(maxMemory)
	if err != nil {
		http.Error(w, "bad multipart form", http.StatusBadRequest)
		return
	}

	var operation string = r.FormValue("operation")
	if operation != "ocr" && operation != "compress" {
		http.Error(w, "unknown operation", http.StatusBadRequest)
		return
	}

	var upload multipart.File
	var header *multipart.FileHeader
	upload, header, err = r.FormFile("archive")
	if err != nil {
		http.Error(w, "no archive uploaded", http.StatusBadRequest)
		return
	}
	defer upload.Close()

	// Copy to a real file on disk. The zip reader needs to seek, which
	// it cannot do on a streamed request body.
	var zipPath string
	zipPath, err = saveTemp(upload)
	if err != nil {
		log.Printf("saving upload: %v", err)
		http.Error(w, "could not save upload", http.StatusInternalServerError)
		return
	}
	defer os.Remove(zipPath)

	var jobID string = newJobID()

	log.Printf("job %s: %s, operation=%s, concurrency=%d",
		jobID, header.Filename, operation, s.ImgRoutines)

	// Reader hands out one image at a time.
	var images <-chan archive.Image
	images, err = archive.Stream(zipPath)
	if err != nil {
		log.Printf("job %s: opening zip: %v", jobID, err)
		http.Error(w, "not a readable zip", http.StatusBadRequest)
		return
	}

	var started time.Time = time.Now()

	var resultCh <-chan invoker.Result = pool.Run(
		r.Context(),
		images,
		s.Invoker,
		operation,
		s.ImgRoutines,
	)

	// Blocks until the pool closes resultCh, which happens only after
	// every image has been handled. So this is where the job finishes.
	var total int
	var failed int

	total, failed, err = results.Write(s.ResultsDir, jobID, operation, resultCh)
	if err != nil {
		log.Printf("job %s: %v", jobID, err)
		http.Error(w, "could not write results", http.StatusInternalServerError)
		return
	}

	log.Printf("job %s: %d images, %d failed, %s",
		jobID, total, failed, time.Since(started).Round(time.Millisecond))

	w.Header().Set("Content-Type", "application/json")

	var body createResponse = createResponse{JobID: jobID}
	err = json.NewEncoder(w).Encode(body)
	if err != nil {
		log.Printf("writing response: %v", err)
	}
}

// saveTemp writes the uploaded zip somewhere on disk and returns its path.
func saveTemp(upload multipart.File) (string, error) {
	var tmp *os.File
	var err error

	tmp, err = os.CreateTemp("", "upload-*.zip")
	if err != nil {
		return "", err
	}
	defer tmp.Close()

	_, err = io.Copy(tmp, upload)
	if err != nil {
		os.Remove(tmp.Name())
		return "", err
	}

	return tmp.Name(), nil
}

// newJobID returns a random hex string used as the results filename.
func newJobID() string {
	var buf []byte = make([]byte, 8)
	rand.Read(buf)
	return hex.EncodeToString(buf)
}
