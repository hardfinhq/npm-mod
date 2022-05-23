// Copyright 2022 Hardfin, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package concurrency

import (
	"sync"
)

// H/T: https://brandur.org/go-worker-pool

// Task encapsulates a work item that should go in a worker pool.
//
// If a task is run more than once, the `Error` will be overwritten.
type Task struct {
	// Error holds an error that occurred during a task. Its
	// result is only meaningful after Run has been called
	// for the pool that holds it.
	Error error

	f func(i int) error
}

// NewTask creates a new task based on a given work function.
func NewTask(f func(i int) error) *Task {
	return &Task{f: f}
}

// Run runs a Task, stores any error and marks the passed in work group as done.
//
// This modifies the current task in a way that is not concurrency safe; the
// expectation is that a given task will be handled by / run in a single
// goroutine.
func (t *Task) Run(wg *sync.WaitGroup, i int) {
	defer wg.Done()
	t.Error = t.f(i)
}

// Pool is a worker group that runs a number of tasks at a configured
// concurrency.
type Pool struct {
	Tasks []*Task

	concurrency int
	tasksChan   chan *Task
	wg          sync.WaitGroup
}

// NewPool initializes a new pool with the given tasks and at the given
// concurrency.
func NewPool(tasks []*Task, concurrency int) *Pool {
	return &Pool{
		Tasks:       tasks,
		concurrency: concurrency,
		tasksChan:   make(chan *Task),
	}
}

// Run runs all work within the pool and blocks until it's finished. Upon
// completion, the task errors will be collected into a multi-error.
func (p *Pool) Run() error {
	for i := 0; i < p.concurrency; i++ {
		go p.work(i)
	}

	p.wg.Add(len(p.Tasks))
	for _, task := range p.Tasks {
		p.tasksChan <- task
	}

	// Close channel just so as not to orphan resources; note since
	// `p.tasksChan` is an unbuffered channel, all of the `p.tasksChan <- task`
	// sends will block until a `doWork()` goroutine can pull them off. (I.e. we
	// won't be closing the channel before it is done being processed.) The only
	// concern is a **send** to a closed channel since a **receive** from a
	// closed channel is still valid.
	close(p.tasksChan)

	p.wg.Wait()

	return p.errors()
}

// work runs the work loop; expected to be run in a single goroutine in
// the worker pool.
func (p *Pool) work(i int) {
	for task := range p.tasksChan {
		task.Run(&p.wg, i)
	}
}

// errors collects the errors for all tasks in the pool.
func (p *Pool) errors() error {
	encountered := make([]error, len(p.Tasks))
	for i, task := range p.Tasks {
		encountered[i] = task.Error
	}
	return maybeMultiError(encountered...)
}
