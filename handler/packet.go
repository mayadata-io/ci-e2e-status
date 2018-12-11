package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/openebs/ci-e2e-dashboard-go-backend/database"
)

// Packethandler return packet pipeline data to api
func Packethandler(w http.ResponseWriter, r *http.Request) {
	datas := dashboard{}
	err := QueryPacketData(&datas)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	out, err := json.Marshal(datas)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(out)
	token, ok := os.LookupEnv(token)
	if !ok {
		panic("TOKEN environment variable required but not set")
	}
	go PacketData(token)
}

// packetPipelineJobs will get pipeline jobs details from gitlab api
func packetPipelineJobs(id int, token string) Jobs {
	url := BaseURL + "api/v4/projects/" + PlatformID["packet"] + "/pipelines/" + strconv.Itoa(id) + "/jobs?per_page=50"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("PRIVATE-TOKEN", token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	var obj Jobs
	json.Unmarshal(body, &obj)
	return obj
}

// PacketData from gitlab api for packet and dump to database
func PacketData(token string) {
	url := BaseURL + "api/v4/projects/" + PlatformID["packet"] + "/pipelines?ref=master"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("PRIVATE-TOKEN", token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	var obj Pipeline
	json.Unmarshal(body, &obj)

	for i := range obj {
		jobsdata := packetPipelineJobs(obj[i].ID, token)
		if err != nil {
			fmt.Println(err)
		}
		jobStartedAt := jobsdata[0].StartedAt
		JobFinishedAt := jobsdata[len(jobsdata)-1].FinishedAt
		logURL := Kibanaloglink(obj[i].Sha, obj[i].ID, obj[i].Status, jobStartedAt, JobFinishedAt)

		// Push Packet pipelines data to Database
		sqlStatement := `
			INSERT INTO packetpipeline (id, sha, ref, status, web_url, kibana_url)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id) DO UPDATE
			SET status = $4, kibana_url = $6
			RETURNING id`
		id := 0
		err = database.Db.QueryRow(sqlStatement, obj[i].ID, obj[i].Sha, obj[i].Ref, obj[i].Status, obj[i].WebURL, logURL).Scan(&id)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("New record ID for Packet Pipeline:", id)

		// Push Packet jobs data to Database
		for j := range jobsdata {
			sqlStatement := `
				INSERT INTO packetjobs (pipelineid, id, status, stage, name, ref, created_at, started_at, finished_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				ON CONFLICT (id) DO UPDATE
				SET status = $3, stage = $4, name = $5, ref = $6, created_at = $7, started_at = $8, finished_at = $9
				RETURNING id`
			id := 0
			err = database.Db.QueryRow(sqlStatement,
				obj[i].ID,
				jobsdata[j].ID,
				jobsdata[j].Status,
				jobsdata[j].Stage,
				jobsdata[j].Name,
				jobsdata[j].Ref,
				jobsdata[j].CreatedAt,
				jobsdata[j].StartedAt,
				jobsdata[j].FinishedAt,
			).Scan(&id)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("New record ID for Packet Jobs: ", id)
		}
	}
}

// QueryPacketData fetch the pipeline data as well as jobs data for packet platform
func QueryPacketData(datas *dashboard) error {
	pipelinerows, err := database.Db.Query(`SELECT * FROM packetpipeline ORDER BY id DESC`)
	if err != nil {
		fmt.Println(err)
	}
	defer pipelinerows.Close()
	for pipelinerows.Next() {
		pipelinedata := pipelineSummary{}
		err = pipelinerows.Scan(
			&pipelinedata.ID,
			&pipelinedata.Sha,
			&pipelinedata.Ref,
			&pipelinedata.Status,
			&pipelinedata.WebURL,
			&pipelinedata.LogURL,
		)
		if err != nil {
			fmt.Println(err)
		}

		jobsquery := `SELECT * FROM packetjobs WHERE pipelineid = $1 ORDER BY id`
		jobsrows, err := database.Db.Query(jobsquery, pipelinedata.ID)
		if err != nil {
			fmt.Println(err)
		}
		defer jobsrows.Close()
		jobsdataarray := []Jobssummary{}
		for jobsrows.Next() {
			jobsdata := Jobssummary{}
			err = jobsrows.Scan(
				&jobsdata.PipelineID,
				&jobsdata.ID,
				&jobsdata.Status,
				&jobsdata.Stage,
				&jobsdata.Name,
				&jobsdata.Ref,
				&jobsdata.CreatedAt,
				&jobsdata.StartedAt,
				&jobsdata.FinishedAt,
			)
			if err != nil {
				fmt.Println(err)
			}
			jobsdataarray = append(jobsdataarray, jobsdata)
			pipelinedata.Jobs = jobsdataarray
		}
		datas.Dashboard = append(datas.Dashboard, pipelinedata)
	}
	err = pipelinerows.Err()
	if err != nil {
		fmt.Println(err)
	}
	return nil
}
