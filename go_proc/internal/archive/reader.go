package archive

//Opens the user's uploaded zip and streams entries one at a time into a channel

import (
	"archive/zip"
	"io"
	"path/filepath"
	"strings"
)

// One Image Struct
type Image struct {
	Name  string
	Bytes []byte
}

var imageExts map[string]bool = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
}

func Stream(zipPath string) (<-chan Image, error) {
	var reader *zip.ReadCloser
	var err error

	reader, err = zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}

	//variable out of type channel carrying Image types
	var out chan Image = make(chan Image)

	/*
		Channel is blocking and waits until another goroutine is waiting to receive.
		This blocks the entire program because we have not created pool to read from channel yet.

		We use a gorountine here to allow go func() to keep running.
		So when we call Stream it can return the channel and gofunc() will be blocked in the background.

	*/
	go func() {
		defer reader.Close()
		defer close(out)

		var readErr error
		var i int

		for i = 0; i < len(reader.File); i++ {
			var entry *zip.File = reader.File[i]

			if !isImage(entry) {
				continue
			}

			var data []byte
			data, readErr = readEntry(entry)
			if readErr != nil {
				continue
			}

			/*
				Construct he Image and send it into channel.
			*/
			out <- Image{Name: entry.Name, Bytes: data}
		}
	}()

	return out, nil
}

func isImage(entry *zip.File) bool {
	if entry.FileInfo().IsDir() {
		return false
	}

	var name string = entry.Name

	if strings.HasPrefix(name, "__MACOSX/") {
		return false
	}
	if strings.HasPrefix(filepath.Base(name), "._") {
		return false
	}

	var ext string = strings.ToLower(filepath.Ext(name))
	return imageExts[ext]
}

func readEntry(entry *zip.File) ([]byte, error) {
	var rc io.ReadCloser
	var err error

	rc, err = entry.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	return io.ReadAll(rc)
}
