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
	_, err := processJobsFromFile("jobsSimple.json")
	checkErr(err, t)
	data, err := ioutil.ReadFile("Out.txt")
	checkErr(err, t)

	expected := []byte("2\n8\n")
	if bytes.Compare(expected, data) != 0 {
		t.Fatalf("Expected: %q, Found: %q", string(expected), string(data))
	}
}

func TestCorrupted(t *testing.T) {
	_, err := processJobsFromFile("jobsCorrupted.json")
	checkErr(err, t)
	data, err := ioutil.ReadFile("Out.txt")
	checkErr(err, t)

	expected := []byte("err\nerr\nerr\nerr\n8\n")
	if bytes.Compare(expected, data) != 0 {
		t.Fatalf("Expected: %q, Found: %q", string(expected), string(data))
	}
}

func TestHuge(t *testing.T) {
	_, err := processJobsFromFile("jobsHuge.json")
	checkErr(err, t)
}
