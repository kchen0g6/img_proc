package pool

import (
	"context"
	"log"

	"golang.org/x/sync/errgroup"

	"img_proc/go_proc/internal/archive"
	"img_proc/go_proc/internal/invoker"
)

type Processor interface {
	Invoke(ctx context.Context, img archive.Image, operation string) (invoker.Result, error)
}

func Run(ctx context.Context, images <-chan archive.Image, proc Processor, operation string, limit int) <-chan invoker.Result {
	var results chan invoker.Result = make(chan invoker.Result)

	go func() {
		defer close(results)

		var group *errgroup.Group
		var groupCtx context.Context

		group, groupCtx = errgroup.WithContext(ctx)

		//Cap goroutine at limit
		group.SetLimit(limit)

		/*
			For each image we will send it to lambda and wait for a reply.
			Lambda will send back a json object.

			Group allows you to track goroutines current running.
			Then the goroutine will push it to results channel.
		*/
		var img archive.Image
		for img = range images {
			var current archive.Image = img

			group.Go(func() error {
				var result invoker.Result
				var err error

				result, err = proc.Invoke(groupCtx, current, operation)
				if err != nil {
					log.Printf("invoke %s: %v", current.Name, err)

					result = invoker.Result{
						Name: current.Name,
						Err:  err.Error(),
					}
				}

				results <- result
				return nil
			})
		}

		// Blocks until every launched goroutine has finished.
		var err error = group.Wait()
		if err != nil {
			log.Printf("pool: %v", err)
		}
	}()

	return results
}
