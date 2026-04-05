package engine

import (
	"sync"

	"github.com/divijg19/Pulse/internal/model"
)

func ExecuteConcurrent(url string, method string, concurrency int) []model.Result {
	var wg sync.WaitGroup
	resultsCh := make(chan model.Result, concurrency)
	var results []model.Result

	for range concurrency {
		wg.Go(func() {
			resultsCh <- ExecuteSingle(url, method)
		})
	}

	wg.Wait()
	close(resultsCh)

	for res := range resultsCh {
		results = append(results, res)
	}

	return results
}
