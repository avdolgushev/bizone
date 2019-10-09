package workers

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
	lock int32

	CurrentWorkers, MaxWorkers int32
	In, Out                    chan Ijob

	Arg interface{}

	wg   sync.WaitGroup
	once bool
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

	if obj.CurrentWorkers < obj.MaxWorkers {
		obj.wg.Add(1)
		go obj.processJobs()
		obj.CurrentWorkers++
	}
}

// Processes input's jobs and sends them to out
func (obj *Workers) processJobs() {
	for job := range obj.In {
		job.DoJob(obj.Arg)
		obj.Out <- job
	}

	obj.Lock()
	defer obj.Unlock()

	obj.wg.Done()
	obj.CurrentWorkers--
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
