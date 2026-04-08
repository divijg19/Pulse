package engine

import (
	"context"
	"sync"

	"github.com/divijg19/Pulse/internal/model"
	"github.com/divijg19/Pulse/internal/stream"
)

func ExecuteConcurrent(ctx context.Context, url string, method string, concurrency int, hub *stream.Hub) {
	var wg sync.WaitGroup

	for range concurrency {
		select {
		case <-ctx.Done():
			wg.Wait()
			return
		default:
		}

		wg.Add(1)

		// Start a new goroutine for EACH request
		go func() {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			default:
			}

			// 1. Make the HTTP call
			res := ExecuteSingle(ctx, url, method)
			if ctx.Err() != nil {
				return
			}

			// 2. Instantly stream the result to the browser!
			hub.Broadcast(model.Event{
				Type: "result",
				Data: res,
			})
		}()
	}

	// Block the function from exiting until all goroutines call wg.Done()
	wg.Wait()
}
