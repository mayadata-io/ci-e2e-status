package database

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/golang/glog"
	"github.com/mayadata-io/ci-e2e-status/config"
)

// Db variable use in other package
var Db *sql.DB

const (
	dbhost = "DBHOST"
	dbport = "DBPORT"
	dbuser = "DBUSER"
	dbpass = "DBPASS"
	dbname = "DBNAME"
)

// InitDb will start DB connection
func InitDb(gitlab config.Config) {
	config, err := dbConfig()
	if err != nil {
		glog.Fatalln(err)
	}
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config[dbhost], config[dbport],
		config[dbuser], config[dbpass], config[dbname])

	Db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		glog.Fatalln(err)
	}
	err = Db.Ping()
	if err != nil {
		glog.Fatalln(err)
	}
	glog.Infoln("Successfully connected to Database!")
	// Create table in database if not present
	createTable(gitlab)
}

// createTable in database if not abvailable
func createTable(gitlab config.Config) {

	for _, project := range gitlab.Projects {
		for _, branch := range project.Branches {
			BranchName := strings.Replace(branch.Name, "-", "_", -1)
			pipelineQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(project VARCHAR, id INT PRIMARY KEY, sha VARCHAR, ref VARCHAR, status VARCHAR, web_url VARCHAR, openshift_pid VARCHAR, k8s_version VARCHAR, release_tag VARCHAR, created_at VARCHAR);", fmt.Sprintf("%s_%s", project.Name, BranchName))
			pipelineQueryExec, err := Db.Query(pipelineQuery)
			if err != nil {
				glog.Error(err)
			}
			defer pipelineQueryExec.Close()
			query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(pipelineid INT, id INT PRIMARY KEY,status VARCHAR, stage VARCHAR, name VARCHAR, ref VARCHAR, github_readme VARCHAR, created_at VARCHAR, started_at VARCHAR, finished_at VARCHAR, message VARCHAR, author_name VARCHAR);", fmt.Sprintf("%s_%s_jobs", project.Name, BranchName))
			value, err := Db.Query(query)
			if err != nil {
				glog.Error(err)
			}
			defer value.Close()
		}
	}
}

// dbConfig get config from environment variable
func dbConfig() (map[string]string, error) {
	conf := make(map[string]string)
	host, ok := os.LookupEnv(dbhost)
	if !ok {
		return nil, errors.New("DBHOST environment variable required")
	}
	port, ok := os.LookupEnv(dbport)
	if !ok {
		return nil, errors.New("DBPORT environment variable required")
	}
	user, ok := os.LookupEnv(dbuser)
	if !ok {
		return nil, errors.New("DBUSER environment variable required")
	}
	password, ok := os.LookupEnv(dbpass)
	if !ok {
		return nil, errors.New("DBPASS environment variable required")
	}
	name, ok := os.LookupEnv(dbname)
	if !ok {
		return nil, errors.New("DBNAME environment variable required")
	}
	conf[dbhost] = host
	conf[dbport] = port
	conf[dbuser] = user
	conf[dbpass] = password
	conf[dbname] = name
	return conf, nil
}
