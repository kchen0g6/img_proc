package results

/*
Written by AI
*/

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"img_proc/go_proc/internal/invoker"
)

type File struct {
	JobID     string           `json:"job_id"`
	Operation string           `json:"operation"`
	Total     int              `json:"total"`
	Failed    int              `json:"failed"`
	Images    []invoker.Result `json:"images"`
}

func Write(
	dir string,
	jobID string,
	operation string,
	incoming <-chan invoker.Result,
) (int, int, error) {

	var collected []invoker.Result = make([]invoker.Result, 0, 64)
	var failed int = 0

	var result invoker.Result
	for result = range incoming {
		if result.Err != "" {
			failed++
		}
		collected = append(collected, result)
	}

	var output File = File{
		JobID:     jobID,
		Operation: operation,
		Total:     len(collected),
		Failed:    failed,
		Images:    collected,
	}

	var encoded []byte
	var err error

	encoded, err = json.MarshalIndent(output, "", "  ")
	if err != nil {
		return len(collected), failed, fmt.Errorf("encoding results: %w", err)
	}

	var path string = filepath.Join(dir, jobID+".json")

	err = os.WriteFile(path, encoded, 0o644)
	if err != nil {
		return len(collected), failed, fmt.Errorf("writing %s: %w", path, err)
	}

	return len(collected), failed, nil
}
