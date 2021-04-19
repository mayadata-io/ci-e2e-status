package database

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/golang/glog"
)

var Platform = [...]string{"nativek8s"}
var Branch = [...]string{"openebs_localpv", "openebs_jiva", "openebs_cstor_csi", "openebs_cstor"}
var NativeBranch = [...]string{"release_branch", "lvm_localpv"}

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
func InitDb() {
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
	createTable()
}

// createTable in database if not abvailable
func createTable() {

	for i := range Platform {
		if Platform[i] == "nativek8s" {
			for j := range NativeBranch {
				query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(project VARCHAR, id INT PRIMARY KEY, sha VARCHAR, ref VARCHAR, status VARCHAR, web_url VARCHAR, openshift_pid VARCHAR, kibana_url VARCHAR, release_tag VARCHAR);", fmt.Sprintf(Platform[i]+"_"+NativeBranch[j]))
				value, err := Db.Query(query)
				if err != nil {
					glog.Error(err)
				}
				defer value.Close()
			}
		} else {
			for j := range Branch {
				query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(project VARCHAR, id INT PRIMARY KEY, sha VARCHAR, ref VARCHAR, status VARCHAR, web_url VARCHAR, openshift_pid VARCHAR, kibana_url VARCHAR, release_tag VARCHAR);", fmt.Sprintf(Platform[i]+"_"+Branch[j]))
				value, err := Db.Query(query)
				if err != nil {
					glog.Error(err)
				}
				defer value.Close()
			}
		}
	}
	// Create pipeline jobs table in database
	for i := range Platform {
		if Platform[i] == "nativek8s" {
			for j := range NativeBranch {
				query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(pipelineid INT, id INT PRIMARY KEY,status VARCHAR, stage VARCHAR, name VARCHAR, ref VARCHAR, github_readme VARCHAR, created_at VARCHAR, started_at VARCHAR, finished_at VARCHAR, message VARCHAR, author_name VARCHAR);", fmt.Sprintf(Platform[i]+"_"+NativeBranch[j]+"_jobs"))
				value, err := Db.Query(query)
				if err != nil {
					glog.Error(err)
				}
				defer value.Close()
			}
		} else {
			for j := range Branch {
				query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(pipelineid INT, id INT PRIMARY KEY,status VARCHAR, stage VARCHAR, name VARCHAR, ref VARCHAR, github_readme VARCHAR, created_at VARCHAR, started_at VARCHAR, finished_at VARCHAR, message VARCHAR, author_name VARCHAR);", fmt.Sprintf(Platform[i]+"_"+Branch[j]+"_jobs"))
				value, err := Db.Query(query)
				if err != nil {
					glog.Error(err)
				}
				defer value.Close()
			}
		}
	}

	// zfs_localpv_query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(project VARCHAR, id INT PRIMARY KEY, sha VARCHAR, ref VARCHAR, status VARCHAR, web_url VARCHAR, openshift_pid VARCHAR, kibana_url VARCHAR, release_tag VARCHAR);", fmt.Sprintf("%s_%s", "zfs_localpv", "release_branch"))
	// zfs_localpv, err := Db.Query(zfs_localpv_query)
	// if err != nil {
	// 	glog.Error(err)
	// }
	// defer zfs_localpv.Close()
	// lvmlocalpv_query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(project VARCHAR, id INT PRIMARY KEY, sha VARCHAR, ref VARCHAR, status VARCHAR, web_url VARCHAR, openshift_pid VARCHAR, kibana_url VARCHAR, release_tag VARCHAR);", fmt.Sprintf("%s_%s", "lvmlocalpv", "lvml-oaclpv"))
	// lvmlocalpv, err := Db.Query(lvmlocalpv_query)
	// if err != nil {
	// 	glog.Error(err)
	// }
	// defer lvmlocalpv.Close()
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
