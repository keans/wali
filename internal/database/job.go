package database

import (
	"crypto/sha256"
	"fmt"
	"time"
	"wali/internal/utils"
	"wali/internal/yaml"
)

const (
	Stopped = iota
	Enqueued
	Running
)

var StateName = map[int]string{
	Stopped:  "stopped",
	Enqueued: "enqueued",
	Running:  "error",
}

type Job struct {
	Key          string
	Url          string
	Xpath        string
	Frequency    int64
	PageHash     string
	Created      interface{}
	LastExecuted interface{}
	LastChange   interface{}
	Status       int8
}

func NewJob(key string, url string, xpath string, frequency int64) *Job {
	return &Job{
		Key:       key,
		Url:       url,
		Xpath:     xpath,
		Frequency: frequency,
		Created:   time.Now(),
		Status:    Stopped,
	}
}

func NewJobFromWebJob(j *yaml.Job) *Job {
	return NewJob(j.Key, j.Url, j.Xpath, j.FrequencyMs)
}

func (j *Job) IsExceeded() bool {
	if j.Status != Stopped {
		// only if status is stopped in can exceed, otherwise
		// it is already waiting for execution or is running
		return false
	}

	if j.LastExecuted == nil {
		// job has never been executed before => mark as exceeded
		return true
	}

	// check if last execution is older than frequency
	return time.Now().Compare(
		(j.LastExecuted).(time.Time).Add(
			time.Duration(j.Frequency)*time.Millisecond)) == 1
}

func (j *Job) Execute(db *Database) bool {
	fmt.Printf("Loading URL: %v\n", j.Url)

	// set status to running
	j.Status = Running
	db.UpdateJob(j)

	body, err := utils.Get(j.Url)
	if err != nil {
		panic(err)
	}

	// compute hash of body
	h := sha256.New()
	h.Write(body)
	hexdigest := fmt.Sprintf("%x", h.Sum(nil))
	fmt.Println(hexdigest)

	// set status to stopped
	j.Status = Stopped
	j.LastExecuted = time.Now()
	j.PageHash = hexdigest
	db.UpdateJob(j)

	fmt.Printf("DONE Loading URL: %v\n", j.Url)

	return true
}
