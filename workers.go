package main

import (
	"runtime"
	"sync"
	"sync/atomic"
)

type Ijob interface {
	doJob(arg interface{})
	getRes() interface{}
}

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
	atomic.AddInt32(&obj.lock, -1)
}

func (obj *Workers) CreateNewWorker() {
	obj.Lock()
	defer obj.Unlock()

	if obj.currentWorkers < obj.maxWorkers {
		obj.wg.Add(1)
		go obj.processJobs()
		obj.currentWorkers++
	}
}

func (obj *Workers) processJobs() {
	for job := range obj.In {
		job.doJob(obj.arg)
		obj.Out <- job
	}

	obj.Lock()
	defer obj.Unlock()

	obj.wg.Done()
	obj.currentWorkers--
}

func (obj *Workers) CreateCloser() {
	if obj.once == false {
		obj.once = true

		go func() {
			obj.wg.Wait()
			close(obj.Out)
		}()
	}
}
