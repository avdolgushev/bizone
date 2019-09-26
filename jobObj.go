package main

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"syscall"
)

// Represents one Job object
type JobObj struct {
	// Values of the object
	// Pointers - to make default values as nil instead of 0
	Arg1, Arg2 *float64
	res        int
	err        bool

	lock int32
}

func (obj *JobObj) getRes() interface{} {
	obj.Lock()
	defer obj.Unlock()

	if obj.err {
		return "err\n"
	} else {
		return fmt.Sprintln(obj.res)
	}
}

func (obj *JobObj) doJob(arg interface{}) {
	obj.Lock()
	defer obj.Unlock()

	proc := arg.(*syscall.Proc)
	if obj.Arg1 == nil || obj.Arg2 == nil || *obj.Arg2 == 0 {
		obj.err = true
	} else {
		res, _, _ := proc.Call(uintptr(*obj.Arg1), uintptr(*obj.Arg2))
		obj.res = int(res)
	}
}

func (obj *JobObj) Lock() {
	for !atomic.CompareAndSwapInt32(&obj.lock, 0, 1) {
		runtime.Gosched()
	}
}

func (obj *JobObj) Unlock() {
	atomic.AddInt32(&obj.lock, -1)
}
