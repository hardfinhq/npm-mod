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

package concurrency_test

import (
	"errors"
	"fmt"
	"sort"
	"testing"

	testifyassert "github.com/stretchr/testify/assert"

	"github.com/hardfinhq/npm-mod/pkg/concurrency"
)

func TestPool_Run_WithoutError(t *testing.T) {
	t.Parallel()
	assert := testifyassert.New(t)

	funcs := []*taskFunc{
		{Input: 2},
		{Input: 5},
		{Input: 16},
		{Input: -4},
		{Input: 202},
		{Input: 205},
		{Input: 216},
		{Input: -204},
	}
	tasks := make([]*concurrency.Task, len(funcs))
	for i, f := range funcs {
		tasks[i] = concurrency.NewTask(f.Do)
	}

	pool := concurrency.NewPool(tasks, 3)
	err := pool.Run()
	assert.Nil(err)

	expected := []*taskFunc{
		{Input: 2, Output: 4},
		{Input: 5, Output: 10},
		{Input: 16, Output: 32},
		{Input: -4, Output: -8},
		{Input: 202, Output: 404},
		{Input: 205, Output: 410},
		{Input: 216, Output: 432},
		{Input: -204, Output: -408},
	}
	assert.Equal(expected, funcs)
}

func TestPool_Run_WorkerAssignment(t *testing.T) {
	t.Parallel()
	assert := testifyassert.New(t)

	n := 32
	poolSize := 5
	assert.Less(poolSize, n)

	channels := make([]chan struct{}, n)
	for i := 0; i < n; i++ {
		channels[i] = make(chan struct{})
	}
	funcs := make([]*serializedTaskFunc, n)
	for i := 0; i < n; i++ {
		// All funcs except the first few will have `poolSize - 1` parents.
		parentCount := minInt(i, poolSize-1)
		// All funcs except the last few will have `poolSize - 1` children.
		childrenStart := minInt(i+1, n)
		childrenEnd := minInt(i+poolSize, n)
		children := channels[childrenStart:childrenEnd]
		funcs[i] = &serializedTaskFunc{
			ParentCount: parentCount,
			Owned:       channels[i],
			Children:    children,
		}
	}
	// Put the functions into tasks
	tasks := make([]*concurrency.Task, len(funcs))
	for i, f := range funcs {
		tasks[i] = concurrency.NewTask(f.Do)
	}

	pool := concurrency.NewPool(tasks, poolSize)
	err := pool.Run()
	assert.Nil(err)

	// We can't control the order in which the first `poolSize` get processed,
	// but everything after that is serialized.
	workerByIndex := make([]int, poolSize)
	for i := 0; i < poolSize; i++ {
		workerByIndex[i] = funcs[i].Worker
	}

	for i := 0; i < n; i++ {
		j := i % poolSize
		assert.Equal(workerByIndex[j], funcs[i].Worker, "Worker %d", i)
	}

	// Sort `workerByIndex` and ensure it contains the expected values.
	sort.Ints(workerByIndex)
	expected := make([]int, poolSize)
	for i := 0; i < poolSize; i++ {
		expected[i] = i
	}
	assert.Equal(expected, workerByIndex)
}

func TestPool_Run_OneError(t *testing.T) {
	t.Parallel()
	assert := testifyassert.New(t)

	known := errors.New("WRENCH Do() 55")
	funcs := []*taskFunc{
		{Input: 12},
		{Input: 55, Error: known},
	}
	tasks := make([]*concurrency.Task, len(funcs))
	for i, f := range funcs {
		tasks[i] = concurrency.NewTask(f.Do)
	}

	pool := concurrency.NewPool(tasks, 1)
	err := pool.Run()
	assert.Equal(known, err)

	expected := []*taskFunc{
		{Input: 12, Output: 24},
		{Input: 55, Output: 110, Error: known},
	}
	assert.Equal(expected, funcs)
}

func TestPool_Run_MultiError(t *testing.T) {
	t.Parallel()
	assert := testifyassert.New(t)

	err1 := errors.New("WRENCH Do() 1337")
	err2 := errors.New("WRENCH Do() 42")
	funcs := []*taskFunc{
		{Input: 1337, Error: err1},
		{Input: 42, Error: err2},
	}
	tasks := make([]*concurrency.Task, len(funcs))
	for i, f := range funcs {
		tasks[i] = concurrency.NewTask(f.Do)
	}

	pool := concurrency.NewPool(tasks, 1)
	err := pool.Run()
	assert.NotNil(err)
	assert.Equal("2 errors occurred:\n\t* WRENCH Do() 1337\n\t* WRENCH Do() 42\n\n", fmt.Sprintf("%v", err))

	expected := []*taskFunc{
		{Input: 1337, Output: 2674, Error: err1},
		{Input: 42, Output: 84, Error: err2},
	}
	assert.Equal(expected, funcs)
}

type taskFunc struct {
	Input  int
	Output int
	Error  error
}

func (tf *taskFunc) Do(_ int) error {
	tf.Output = 2 * tf.Input
	return tf.Error
}

// serializedTaskFunc is a way to serialize work across `N` workers. We require
// all functions to interact with "all parents" and "all children" so that
// all tasks block the workers until the "next" worker slot is filled.
//
// The goal is to control the order in which tasks exist / tasks are scheduled
// for work on a worker. This way we can write deterministic tests for worker
// assignment.
type serializedTaskFunc struct {
	ParentCount int
	Owned       chan struct{}
	Children    []chan struct{}

	Worker int
}

func (stf *serializedTaskFunc) Do(i int) error {
	// Consume one value off the `Owned` channel for every parent
	for j := 0; j < stf.ParentCount; j++ {
		<-stf.Owned
	}
	stf.Worker = i
	// Send a value out to the channel owned by every child
	for _, c := range stf.Children {
		c <- struct{}{}
	}
	return nil
}

func minInt(i, j int) int {
	if i < j {
		return i
	}
	return j
}
