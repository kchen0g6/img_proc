package invoker

/*
Written by AI
*/

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awslambda "github.com/aws/aws-sdk-go-v2/service/lambda"

	"img_proc/go_proc/internal/archive"
)

// What the worker on AWS expects to receive.
type request struct {
	Name      string `json:"name"`
	Operation string `json:"operation"`
	ImageB64  string `json:"image_b64"`
}

/*
What the worker sends back, and what ends up in results.json.

Text is filled for ocr, the byte counts for compress. Err carries a
per-image failure without killing the whole job.
*/
type Result struct {
	Name     string `json:"name"`
	Text     string `json:"text,omitempty"`
	BytesIn  int    `json:"bytes_in,omitempty"`
	BytesOut int    `json:"bytes_out,omitempty"`
	Err      string `json:"error,omitempty"`
}

// Holds the AWS client, built once and shared by every goroutine.
// The SDK client is safe for concurrent use.
type Client struct {
	lambda       *awslambda.Client
	functionName string
}

// Synchronous invoke caps the request payload at 6 MB, and base64 inflates
// bytes by about a third. Reject anything that would not fit.
const maxImageBytes int = 4 << 20 // 4 MB

/*
New builds a Lambda client from the usual AWS credential sources: the same
environment variables, profile, and config files the aws CLI reads.
*/
func New(ctx context.Context, functionName string) (*Client, error) {
	var cfg aws.Config
	var err error

	cfg, err = config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading aws config: %w", err)
	}

	var client *Client = &Client{
		lambda:       awslambda.NewFromConfig(cfg),
		functionName: functionName,
	}

	return client, nil
}

/*
Invoke sends one image and returns what the worker produced.

A returned error means the call itself failed. A Result carrying Err means
the call worked but that image could not be processed.
*/
func (c *Client) Invoke(ctx context.Context, img archive.Image, operation string) (Result, error) {
	if len(img.Bytes) > maxImageBytes {
		return Result{
			Name: img.Name,
			Err:  "image too large for a synchronous invoke",
		}, nil
	}

	var payload []byte
	var err error

	payload, err = json.Marshal(request{
		Name:      img.Name,
		Operation: operation,
		ImageB64:  base64.StdEncoding.EncodeToString(img.Bytes),
	})
	if err != nil {
		return Result{}, fmt.Errorf("encoding request: %w", err)
	}

	var out *awslambda.InvokeOutput
	out, err = c.lambda.Invoke(ctx, &awslambda.InvokeInput{
		FunctionName: aws.String(c.functionName),
		Payload:      payload,
	})
	if err != nil {
		return Result{}, fmt.Errorf("invoking %s: %w", c.functionName, err)
	}

	// The worker panicked or timed out. Payload holds the stack trace.
	if out.FunctionError != nil {
		return Result{
			Name: img.Name,
			Err:  fmt.Sprintf("worker failed: %s", string(out.Payload)),
		}, nil
	}

	var result Result
	err = json.Unmarshal(out.Payload, &result)
	if err != nil {
		return Result{}, fmt.Errorf("decoding response: %w", err)
	}

	if result.Name == "" {
		result.Name = img.Name
	}

	return result, nil
}

// ErrNoFunctionName is returned when the caller forgot to set one.
var ErrNoFunctionName error = errors.New("lambda function name is empty")
