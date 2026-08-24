package main

/*
Written by AI

Build:
    GOOS=linux GOARCH=arm64 go build -o dist/bootstrap ./cmd/worker
*/

import (
	"bytes"
	"context"
	"encoding/base64"
	"image"
	"image/jpeg"
	"log"
	"strings"

	_ "image/png"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/textract"
	"github.com/aws/aws-sdk-go-v2/service/textract/types"
)

type request struct {
	Name      string `json:"name"`
	Operation string `json:"operation"`
	ImageB64  string `json:"image_b64"`
}

type response struct {
	Name     string `json:"name"`
	Text     string `json:"text,omitempty"`
	BytesIn  int    `json:"bytes_in,omitempty"`
	BytesOut int    `json:"bytes_out,omitempty"`
	Err      string `json:"error,omitempty"`
}

const jpegQuality int = 60

var textractClient *textract.Client

func init() {
	var cfg, err = config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Printf("aws config: %v", err)
		return
	}
	textractClient = textract.NewFromConfig(cfg)
}

func handle(ctx context.Context, req request) (response, error) {
	var raw []byte
	var err error

	raw, err = base64.StdEncoding.DecodeString(req.ImageB64)
	if err != nil {
		return response{Name: req.Name, Err: "bad base64"}, nil
	}

	switch req.Operation {
	case "compress":
		return compress(req.Name, raw), nil
	case "ocr":
		return ocr(ctx, req.Name, raw), nil
	default:
		return response{Name: req.Name, Err: "unknown operation"}, nil
	}
}

/*
ocr sends the image to Textract and joins the detected lines into one
string. Textract returns a flat list of blocks; only LINE blocks are
wanted, since WORD blocks would duplicate the same text.
*/
func ocr(ctx context.Context, name string, raw []byte) response {
	if textractClient == nil {
		return response{Name: name, Err: "textract client not initialised"}
	}

	var out *textract.DetectDocumentTextOutput
	var err error

	out, err = textractClient.DetectDocumentText(ctx, &textract.DetectDocumentTextInput{
		Document: &types.Document{Bytes: raw},
	})
	if err != nil {
		return response{Name: name, Err: "textract: " + err.Error()}
	}

	var lines []string = make([]string, 0, len(out.Blocks))

	var i int
	for i = 0; i < len(out.Blocks); i++ {
		var block types.Block = out.Blocks[i]
		if block.BlockType == types.BlockTypeLine && block.Text != nil {
			lines = append(lines, *block.Text)
		}
	}

	return response{
		Name: name,
		Text: strings.Join(lines, "\n"),
	}
}

func compress(name string, raw []byte) response {
	var img image.Image
	var err error

	img, _, err = image.Decode(bytes.NewReader(raw))
	if err != nil {
		return response{Name: name, Err: "could not decode image"}
	}

	var buf bytes.Buffer
	var opts *jpeg.Options = &jpeg.Options{Quality: jpegQuality}

	err = jpeg.Encode(&buf, img, opts)
	if err != nil {
		return response{Name: name, Err: "could not encode jpeg"}
	}

	return response{
		Name:     name,
		BytesIn:  len(raw),
		BytesOut: buf.Len(),
	}
}

func main() {
	lambda.Start(handle)
}
