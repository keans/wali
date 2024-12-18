package database

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"time"

	"github.com/k3a/html2text"

	"github.com/keans/wali/internal/utils"
	"github.com/keans/wali/internal/yaml"
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
func (j *Job) Execute(db *Database, smtp *utils.Smtp, log *slog.Logger) bool {
	log.Info("loading URL", "url", j.Url)

	// set status to running
	j.Status = Running
	db.UpdateJob(j)

	body, err := utils.Get(j.Url, j.Xpath)
	if err != nil {
		log.Error("Could not get URL", "url", j.Url, "err", err)
		return false
	}

	// compute hash of body
	h := sha256.New()
	h.Write(body)
	hexdigest := fmt.Sprintf("%x", h.Sum(nil))

	if j.PageHash != hexdigest {
		// there was a change
		log.Info("change of page detected", "key", j.Key, "url", j.Url,
			"digest", hexdigest)

		j.PageHash = hexdigest

		// convert HTML to text for mail
		plain := html2text.HTML2Text(string(body))

		// prepare message and send it
		subject := fmt.Sprintf("[Wali] Change of %s detected", j.Key)
		msg := fmt.Sprintf("A change of %s has been detected on %s:\n\n%s",
			j.Url, time.Now().Format(time.ANSIC), plain)

		log.Info("sending mail", "subject", subject)

		smtp.SendMail(subject, msg)
	}

	// set status to stopped
	j.Status = Stopped
	j.LastExecuted = time.Now()
	j.PageHash = hexdigest
	db.UpdateJob(j)

	log.Info("loading URL completed", "url", j.Url)

	return true
}
