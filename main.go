package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

func checkErrFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Gets specified function from math.dll and nil error if success
func getProc(procname string) (*syscall.Proc, error) {
	wd, _ := os.Getwd()
	dllpath := ""

	if runtime.GOOS == "windows" {
		if runtime.GOARCH == "386" {
			dllpath = filepath.Join(wd, "math_x32.dll")
		} else {
			dllpath = filepath.Join(wd, "math_x64.dll")
		}
	} else {
		return nil, errors.New("unsupported OS")
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

// Processes all jobs from file and saves results to out.txt
func processJobs(path string) (int, error) {
	proc, err := getProc("Div")
	checkErrFatal(err)

	workers := Workers{
		maxWorkers: 10000,
		arg:        proc,
		In:         make(chan Ijob, 1000),
		Out:        make(chan Ijob, 1000),
	}

	fin, err := os.Open(path)
	checkErrFatal(err)

	go processJobsFromReader(fin, &workers)

	count, err := processOutput(&workers, "out.txt")
	if err != nil {
		return 0, err
	}

	log.Printf("Calculated %d jobs, saved to %s\n", count, "out.txt")
	return count, nil
}

func processOutput(workers *Workers, outpath string) (count int, err error) {
	fout, err := os.Create(outpath)
	checkErrFatal(err)
	wr := bufio.NewWriter(fout)
	defer fout.Close()

	count = 0
	for vi := range workers.Out {
		v := vi.(*JobObj).getRes().(string)
		_, err = wr.WriteString(v)
		if err != nil {
			log.Println(err)
		}
		count += 1
	}
	wr.Flush()
	return
}

func processJobsFromReader(reader io.Reader, workers *Workers) {
	dec := json.NewDecoder(reader)

	// read opening [
	token, err := dec.Token()
	checkErrFatal(err)

	if tokend, ok := token.(json.Delim); ok == false || tokend.String() != "[" {
		log.Fatal("First token != '['")
	}

	for dec.More() {
		var parsed JobObj
		err = dec.Decode(&parsed)
		checkErrFatal(err)

		workers.CreateNewWorker()
		workers.In <- &parsed
	}

	// read closing ]
	token, err = dec.Token()
	checkErrFatal(err)

	workers.CreateCloser()
	close(workers.In)
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
