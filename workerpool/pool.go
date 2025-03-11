package workerpool

import (
	"sync"
)

type Pool[J any, R any] struct {
	Jobs    chan J
	Results chan R
	Workers []*Worker[J, R]
	wg      sync.WaitGroup
}

type Worker[J any, R any] struct{}

func NewPool[J any, R any](workerCount uint) *Pool[J, R] {
	pool := &Pool[J, R]{
		Jobs:    make(chan J, workerCount),
		Results: make(chan R, workerCount),
		Workers: make([]*Worker[J, R], workerCount),
		wg:      sync.WaitGroup{},
	}

	for i := 0; uint(i) < workerCount; i++ {
		pool.Workers[i] = &Worker[J, R]{}
	}

	return pool
}

func (pool *Pool[J, R]) Start(workerFunc func(J) R) {
	pool.wg.Add(len(pool.Workers))

	for _, worker := range pool.Workers {
		go func(w *Worker[J, R]) {
			defer pool.wg.Done()
			for job := range pool.Jobs {
				result := workerFunc(job)
				pool.Results <- result
			}
		}(worker)
	}
}

func (pool *Pool[J, R]) Submit(job J) {
	pool.Jobs <- job
}

func (pool *Pool[J, R]) GetResults() <-chan R {
	return pool.Results
}

func (pool *Pool[J, R]) Shutdown() {
	close(pool.Jobs)
	pool.wg.Wait()
	close(pool.Results)
}
