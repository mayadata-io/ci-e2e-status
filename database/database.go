package database

import (
	"database/sql"
	"fmt"
	"os"
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
func InitDb() {
	config := dbConfig()
	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config[dbhost], config[dbport],
		config[dbuser], config[dbpass], config[dbname])

	Db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	err = Db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully connected to Database!")
	createTable()
}

func createTable() {
	platform := map[string][]string{
		"pipeline":     []string{"gkepipeline", "akspipeline", "ekspipeline", "packetpipeline", "gcppipeline", "awspipeline", "buildpipeline"},
		"pipelineJobs": []string{"gkejobs", "aksjobs", "eksjobs", "packetjobs", "gcpjobs", "awsjobs", "buildjobs"},
	}
	for i := range platform["pipeline"] {
		query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(id INT PRIMARY KEY, sha VARCHAR, ref VARCHAR, status VARCHAR, web_url VARCHAR, kibana_url VARCHAR);", platform["pipeline"][i])
		_, err := Db.Query(query)
		if err != nil {
			fmt.Println(err)
		}

		query = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(pipelineid INT, id INT PRIMARY KEY,status VARCHAR, stage VARCHAR, name VARCHAR, ref VARCHAR, created_at VARCHAR, started_at VARCHAR, finished_at VARCHAR);", platform["pipelineJobs"][i])
		_, err = Db.Query(query)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func dbConfig() map[string]string {
	conf := make(map[string]string)
	host, ok := os.LookupEnv(dbhost)
	if !ok {
		panic("DBHOST environment variable required but not set")
	}
	port, ok := os.LookupEnv(dbport)
	if !ok {
		panic("DBPORT environment variable required but not set")
	}
	user, ok := os.LookupEnv(dbuser)
	if !ok {
		panic("DBUSER environment variable required but not set")
	}
	password, ok := os.LookupEnv(dbpass)
	if !ok {
		panic("DBPASS environment variable required but not set")
	}
	name, ok := os.LookupEnv(dbname)
	if !ok {
		panic("DBNAME environment variable required but not set")
	}
	conf[dbhost] = host
	conf[dbport] = port
	conf[dbuser] = user
	conf[dbpass] = password
	conf[dbname] = name
	return conf
}
