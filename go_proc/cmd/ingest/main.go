package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"img_proc/go_proc/internal/api"
	"img_proc/go_proc/internal/invoker"
)

const resultsDir string = "../results"
const functionName string = "img-proc-worker"

func main() {
	// Allow flags to be set on command
	var concurrentImgThreads *int = flag.Int("concurrency", 8, "images in flight at once")
	var port *string = flag.String("port", "8080", "port Rails talks to")
	flag.Parse()

	var ctx context.Context = context.Background()

	var lambdaClient *invoker.Client
	var err error

	lambdaClient, err = invoker.New(ctx, functionName)
	if err != nil {
		log.Fatalf("aws setup failed: %v", err)
	}

	var server *api.Server = &api.Server{
		ImgRoutines: *concurrentImgThreads,
		ResultsDir:  resultsDir,
		Invoker:     lambdaClient,
	}

	http.HandleFunc("POST /jobs", server.CreateJob)

	log.Printf("Ingestion server listening on port:%s (concurrency=%d, results=%s)",
		*port, *concurrentImgThreads, resultsDir)

	err = http.ListenAndServe(":"+*port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
