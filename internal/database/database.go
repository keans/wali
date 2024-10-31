package database

import (
	"database/sql"
	"log"
	"log/slog"
	"os"

	"github.com/keans/wali/internal/yaml"

	_ "github.com/mattn/go-sqlite3"
)

const (
	createTableSql string = `CREATE TABLE IF NOT EXISTS jobs (
		key VARCHAR(50) NOT NULL PRIMARY KEY,
		url VARCHAR(255),
		xpath VARCHAR(255),
		frequency INTEGER,
		page_hash VARCHAR(32),
		created DATETIME,
		last_executed DATETIME,
		last_change DATETIME,
		status INTEGER
	);`
	queryJobs        string = `SELECT * FROM jobs;`
	queryJobByKeySql string = `SELECT * FROM jobs WHERE key=?;`
	insertJobSql     string = `INSERT INTO jobs(key, url, xpath,frequency,
		page_hash, created, last_executed, last_change, status)
		values(?, ?, ?, ?, ?, ?, ?, ?, ?);`
	updateJobSql string = `UPDATE jobs SET url=?, xpath=?,
		frequency=?, page_hash=?, created=?, last_executed=?,
		last_change=?, status=? WHERE key=?;`
	deleteJobSql               string = `DELETE FROM jobs WHERE key=?;`
	updateJobStatusesToStopSql string = `UPDATE jobs SET status=0;`
)

type Database struct {
	Filename string
	db       *sql.DB
	log      *slog.Logger
}

// create new database instance of given filename
func NewDb(filename string) *Database {
	return &Database{
		Filename: filename,
		log:      slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
}

// open the database
func (db *Database) Open() error {
	db.log.Debug("opening database", "filename", db.Filename)

	var err error
	db.db, err = sql.Open("sqlite3", db.Filename)

	return err
}

// close the database
func (db *Database) Close() error {
	db.log.Debug("closing database", "filename", db.Filename)

	return db.db.Close()
}

func (db *Database) CreateTables() error {
	db.log.Debug("creating tables (if not existing)",
		"filename", db.Filename)

	_, err := db.db.Exec(createTableSql)
	return err
}

// insert a job into the database
func (db *Database) InsertJob(wj *Job) error {
	tx, err := db.db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare(insertJobSql)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(wj.Key, wj.Url, wj.Xpath, wj.Frequency, wj.PageHash,
		wj.Created, nil, nil, wj.Status)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

// update an existing job in the database
func (db *Database) UpdateJob(wj *Job) error {
	tx, err := db.db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare(updateJobSql)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(wj.Url, wj.Xpath, wj.Frequency, wj.PageHash,
		wj.Created, wj.LastExecuted, wj.LastChange, wj.Status, wj.Key)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

// delete a job from the database by given key
func (db *Database) DeleteJob(key string) error {
	_, err := db.db.Exec(deleteJobSql, key)

	return err
}

// resets all job statuses to Stopped (=0)
func (db *Database) ResetJobsStatuses() error {
	tx, err := db.db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare(updateJobStatusesToStopSql)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

// returns job from database by given key
func (db *Database) GetJobByKey(key string) (*Job, error) {
	stmt, err := db.db.Prepare(queryJobByKeySql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var j Job
	if err = stmt.QueryRow(key).Scan(&j.Key, &j.Url, &j.Xpath, &j.Frequency,
		&j.PageHash, &j.Created, &j.LastExecuted, &j.LastChange,
		&j.Status); err != nil {

		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &j, nil
}

// get a list of all jobs
func (db *Database) GetAllJobs() ([]*Job, error) {
	stmt, err := db.db.Prepare(queryJobs)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		if err == sql.ErrNoRows {
			return []*Job{}, nil
		}

		return nil, err
	}

	var jobs []*Job
	for rows.Next() {
		var j Job
		rows.Scan(&j.Key, &j.Url, &j.Xpath, &j.Frequency, &j.PageHash,
			&j.Created, &j.LastExecuted, &j.LastChange, &j.Status)

		jobs = append(jobs, &j)
	}

	return jobs, nil
}

// adds all jobs from the YAML file to the dabase
// (existing ones will be updated)
func (db *Database) AddFromYaml(jobs *yaml.WaliYaml) {
	for _, j := range jobs.WebJobs {
		// get job from database
		job, err := db.GetJobByKey(j.Key)
		if err != nil {
			panic(err)
		}

		if job == nil {
			// job not yet in database => add it
			db.log.Debug("inserting job", "job", j)
			if err := db.InsertJob(NewJobFromWebJob(&j)); err != nil {
				panic(err)
			}

		} else {
			// job already in the database => update it
			job.Url = j.Url
			job.Xpath = j.Xpath
			job.Frequency = j.FrequencyMs

			db.log.Debug("updating job", "job", job)
			if err := db.UpdateJob(job); err != nil {
				panic(err)
			}
		}
	}
}

// removes all jobs that in the database, but not in the YAML anymore
func (db *Database) ClearJobsNotInYaml(jobs *yaml.WaliYaml) {
	// get all jobs from database
	dbJobs, err := db.GetAllJobs()
	if err != nil {
		panic(err)
	}

	for _, dbJob := range dbJobs {
		// find db job key in YAML jobs
		isInYaml := false
		for _, j := range jobs.WebJobs {
			if dbJob.Key == j.Key {
				// database key is matching the key from YAML
				isInYaml = true
				break
			}
		}

		if !isInYaml {
			// db key is not in YAML => remove from db
			db.log.Debug("deleting job from database", "jobKey", dbJob.Key)
			if err := db.DeleteJob(dbJob.Key); err != nil {
				panic(err)
			}
		}
	}
}
