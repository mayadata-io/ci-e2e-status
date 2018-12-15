package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/openebs/ci-e2e-dashboard-go-backend/database"
)

// Awshandler return aws pipeline data to api
func Awshandler(w http.ResponseWriter, r *http.Request) {
	(w).Header().Set("Access-Control-Allow-Origin", "*")
	datas := dashboard{}
	err := queryAwsData(&datas)
	if err != nil {
		http.Error(w, err.Error(), 500)
		fmt.Println(err)
		return
	}
	out, err := json.Marshal(datas)
	if err != nil {
		http.Error(w, err.Error(), 500)
		fmt.Println(err)
		return
	}
	w.Write(out)
}

// awsPipelineJobs will get pipeline jobs details from gitlab api
func awsPipelineJobs(id int, token string) Jobs {
	url := BaseURL + "api/v4/projects/" + PlatformID["aws"] + "/pipelines/" + strconv.Itoa(id) + "/jobs?per_page=50"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	http.DefaultClient.Timeout = time.Minute * 10
	req.Close = true
	req.Header.Set("Connection", "close")
	req.Header.Add("PRIVATE-TOKEN", token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	var obj Jobs
	json.Unmarshal(body, &obj)
	return obj
}

// awsPipeline get pipeline data from gitlab
func awsPipeline(token string) Pipeline {
	url := BaseURL + "api/v4/projects/" + PlatformID["aws"] + "/pipelines?ref=master"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	http.DefaultClient.Timeout = time.Minute * 10
	req.Close = true
	req.Header.Set("Connection", "close")
	req.Header.Add("PRIVATE-TOKEN", token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	var obj Pipeline
	json.Unmarshal(body, &obj)
	return obj
}

// AwsData from gitlab api for aws and dump to database
func AwsData(token string) {
	awsObj := awsPipeline(token)
	for i := range awsObj {
		jobsdata := awsPipelineJobs(awsObj[i].ID, token)
		jobStartedAt := jobsdata[0].StartedAt
		JobFinishedAt := jobsdata[len(jobsdata)-1].FinishedAt
		logURL := Kibanaloglink(awsObj[i].Sha, awsObj[i].ID, awsObj[i].Status, jobStartedAt, JobFinishedAt)

		// Add Aws pipelines data to Database
		sqlStatement := `
			INSERT INTO awspipeline (id, sha, ref, status, web_url, kibana_url)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id) DO UPDATE
			SET status = $4, kibana_url = $6
			RETURNING id`
		id := 0
		err := database.Db.QueryRow(sqlStatement, awsObj[i].ID, awsObj[i].Sha, awsObj[i].Ref, awsObj[i].Status, awsObj[i].WebURL, logURL).Scan(&id)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("New record ID for AWS Pipeline:", id)

		// Add Aws jobs data to Database
		for j := range jobsdata {
			sqlStatement := `
				INSERT INTO awsjobs (pipelineid, id, status, stage, name, ref, created_at, started_at, finished_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				ON CONFLICT (id) DO UPDATE
				SET status = $3, stage = $4, name = $5, ref = $6, created_at = $7, started_at = $8, finished_at = $9
				RETURNING id`
			id := 0
			err = database.Db.QueryRow(sqlStatement,
				awsObj[i].ID,
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
			fmt.Println("New record ID for AWS Jobs: ", id)
		}
	}
}

// queryAwsData fetch the pipeline data as well as jobs data for aws platform
func queryAwsData(datas *dashboard) error {
	pipelinerows, err := database.Db.Query(`SELECT * FROM awspipeline ORDER BY id DESC`)
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

		jobsquery := `SELECT * FROM awsjobs WHERE pipelineid = $1 ORDER BY id`
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
