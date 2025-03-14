/*
   gobooru-downloader
   Copyright (C) 2025 Kasyanov Nikolay Alexeevich (Unbewohnte)

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

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
