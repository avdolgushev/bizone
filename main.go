package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

// Represents one Job object
type JobObj struct {
	// Values of the object
	// Pointers - to make default values as nil instead of 0
	Arg1, Arg2 *int
	rawdata    []byte
	res        int
	err        bool
	*sync.WaitGroup
}

// Gets specified function from math.dll and nil error if success
func getProc(procname string) (*syscall.Proc, error) {
	wd, _ := os.Getwd()
	dllpath := ""

	if t := uintptr(0); unsafe.Sizeof(t) == 8 {
		dllpath = filepath.Join(wd, "math_x64.dll")
	} else {
		dllpath = filepath.Join(wd, "math_x32.dll")
	}

	dll, err := syscall.LoadDLL(dllpath)
	if err != nil {
		return nil, err
	} else {
		log.Println("Loaded dll: ", dllpath)
	}

	proc, err := dll.FindProc(procname)
	if err != nil {
		return nil, err
	} else {
		log.Println("Proc found: ", procname)
	}

	return proc, nil
}

// Performs the passed job
func calcOne(job *JobObj, proc *syscall.Proc) {
	err := json.Unmarshal(bytes.TrimRight(job.rawdata, ","), &job)
	if err != nil || job.Arg1 == nil || job.Arg2 == nil || *job.Arg2 == 0 {
		job.err = true
	} else {
		res, _, _ := proc.Call(uintptr(*job.Arg1), uintptr(*job.Arg2))
		job.res = int(res)
	}

	job.Done()
}

// Performs all jobs from file and saves it to out.txt
func processJobs(path string) (int, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, err
	}

	b = bytes.Trim(b, "[] \t\r\n")
	bs := bytes.SplitAfter(b, []byte("},"))

	proc, err := getProc("Div")
	if err != nil {
		return 0, err
	}

	jobs := make([]JobObj, len(bs))
	var wg sync.WaitGroup
	//wg.Add(len(bs))

	for i := range bs {
		wg.Add(1)
		jobs[i] = JobObj{rawdata: bs[i], WaitGroup: &wg}
		go calcOne(&jobs[i], proc)
	}

	outpath := "out.txt"
	fout, err := os.Create(outpath)
	if err != nil {
		return 0, err
	}
	defer fout.Close()

	wg.Wait()
	for _, v := range jobs {
		if v.err {
			_, err = fout.WriteString("err\n")
		} else {
			_, err = fout.WriteString(fmt.Sprintln(v.res))
		}
		if err != nil {
			log.Println(err)
		}
	}
	log.Printf("Calculated %d jobs, saved to %s\n", len(jobs), outpath)

	return len(jobs), nil
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: program.exe \"path_to_jobs_file\"")
	}

	start := time.Now()
	_, err := processJobs(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	elapsed := time.Since(start)
	log.Printf("Took time: %v", elapsed)
}
