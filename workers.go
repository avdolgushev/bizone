package main

import (
	"runtime"
	"sync"
	"sync/atomic"
)

type Ijob interface {
	DoJob(arg interface{})
}

// Structure to run several Ijob workers
type Workers struct {
	currentWorkers, maxWorkers int32
	In, Out                    chan Ijob

	arg interface{}

	wg   sync.WaitGroup
	once bool

	lock int32
}

func (obj *Workers) Lock() {
	for !atomic.CompareAndSwapInt32(&obj.lock, 0, 1) {
		runtime.Gosched()
	}
}

func (obj *Workers) Unlock() {
	atomic.StoreInt32(&obj.lock, 0)
}

// Creates new worker if it fits the limit
func (obj *Workers) CreateNewWorker() {
	obj.Lock()
	defer obj.Unlock()

	if obj.currentWorkers < obj.maxWorkers {
		obj.wg.Add(1)
		go obj.processJobs()
		obj.currentWorkers++
	}
}

// Processes input's jobs and sends them to out
func (obj *Workers) processJobs() {
	for job := range obj.In {
		job.DoJob(obj.arg)
		obj.Out <- job
	}

	obj.Lock()
	defer obj.Unlock()

	obj.wg.Done()
	obj.currentWorkers--
}

// run a goroutine that will close out channel when all workers finish
func (obj *Workers) CreateCloser() {
	obj.Lock()
	defer obj.Unlock()

	if obj.once == false {
		obj.once = true

		go func() {
			obj.wg.Wait()
			close(obj.Out)
		}()
	}
}
