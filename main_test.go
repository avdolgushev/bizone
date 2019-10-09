package main

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func checkErr(err error, t *testing.T) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestSimple(t *testing.T) {
	_, err := processJobsFromFile("testFiles/jobsSimple.json")
	checkErr(err, t)
	data, err := ioutil.ReadFile("out.txt")
	checkErr(err, t)

	res1, res2 := "2\n8\n", "8\n2\n"
	if bytes.Compare([]byte(res1), data) != 0 && bytes.Compare([]byte(res2), data) != 0 {
		t.Fatalf("Expected: %q or %q, Found: %q", res1, res2, string(data))
	}
}

func TestCorrupted(t *testing.T) {
	_, err := processJobsFromFile("testFiles/jobsCorrupted.json")
	checkErr(err, t)
	data, err := ioutil.ReadFile("out.txt")
	checkErr(err, t)

	expected := bytes.Repeat([]byte("err\n"), 10)
	if bytes.Compare(expected, data) != 0 {
		t.Fatalf("Expected: %q, Found: %q", string(expected), string(data))
	}
}

func TestHuge(t *testing.T) {
	_, err := processJobsFromFile("testFiles/jobsHuge.json")
	checkErr(err, t)
	data, err := ioutil.ReadFile("out.txt")
	checkErr(err, t)

	if bytes.Count(data, []byte("2\n")) != 25000 || bytes.Count(data, []byte("8\n")) != 25000 || bytes.Count(data, []byte("\n")) != 50000 {
		t.Fatal("Expected: 25000 of \"2\\n\" and \"8\\n\" and 50000 lines.")
	}
}
