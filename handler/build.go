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

// Buildhandler return packet pipeline data to api
func Buildhandler(w http.ResponseWriter, r *http.Request) {
	datas := dashboard{}
	err := QueryBuildData(&datas)
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
	go BuildData(token)
}

// jivaPipelineJobs will get pipeline jobs details from gitlab api
func jivaPipelineJobs(id int, token string) Jobs {
	url := BaseURL + "api/v4/projects/" + PlatformID["jiva"] + "/pipelines/" + strconv.Itoa(id) + "/jobs?per_page=50"
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

// mayaPipelineJobs will get pipeline jobs details from gitlab api
func mayaPipelineJobs(id int, token string) Jobs {
	url := BaseURL + "api/v4/projects/" + PlatformID["maya"] + "/pipelines/" + strconv.Itoa(id) + "/jobs?per_page=50"
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

// BuildData from gitlab api for packet and dump to database
func BuildData(token string) {
	// Fetch jiva data from gitlab
	jivaURL := BaseURL + "api/v4/projects/" + PlatformID["jiva"] + "/pipelines?ref=master"
	req, _ := http.NewRequest("GET", jivaURL, nil)
	req.Header.Add("PRIVATE-TOKEN", token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	jivaData, _ := ioutil.ReadAll(res.Body)
	var jivaObj Pipeline
	json.Unmarshal(jivaData, &jivaObj)

	for i := range jivaObj {
		jivaJobsData := jivaPipelineJobs(jivaObj[i].ID, token)
		if err != nil {
			fmt.Println(err)
		}
		// Push jiva pipelines data to Database
		sqlStatement := `
			INSERT INTO buildpipeline (id, sha, ref, status, web_url, kibana_url)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id) DO UPDATE
			SET status = $4
			RETURNING id`
		id := 0
		err = database.Db.QueryRow(sqlStatement, jivaObj[i].ID, jivaObj[i].Sha, jivaObj[i].Ref, jivaObj[i].Status, jivaObj[i].WebURL, "/").Scan(&id)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("New record ID for jiva build Pipeline:", id)

		// Push Aks jobs data to Database
		for j := range jivaJobsData {
			sqlStatement := `
				INSERT INTO buildjobs (pipelineid, id, status, stage, name, ref, created_at, started_at, finished_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				ON CONFLICT (id) DO UPDATE
				SET status = $3, stage = $4, name = $5, ref = $6, created_at = $7, started_at = $8, finished_at = $9
				RETURNING id`
			id := 0
			err = database.Db.QueryRow(sqlStatement,
				jivaObj[i].ID,
				jivaJobsData[j].ID,
				jivaJobsData[j].Status,
				jivaJobsData[j].Stage,
				jivaJobsData[j].Name,
				jivaJobsData[j].Ref,
				jivaJobsData[j].CreatedAt,
				jivaJobsData[j].StartedAt,
				jivaJobsData[j].FinishedAt,
			).Scan(&id)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("New record ID for jiva pipeline Jobs: ", id)
		}
	}

	// // Fetch maya data from gitlab
	mayaURL := BaseURL + "api/v4/projects/" + PlatformID["maya"] + "/pipelines?ref=master"
	req, _ = http.NewRequest("GET", mayaURL, nil)
	req.Header.Add("PRIVATE-TOKEN", token)
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	mayaData, _ := ioutil.ReadAll(res.Body)
	var mayaObj Pipeline
	json.Unmarshal(mayaData, &mayaObj)

	for i := range mayaObj {
		mayaJobsData := mayaPipelineJobs(mayaObj[i].ID, token)
		if err != nil {
			fmt.Println(err)
		}
		// Push jiva pipelines data to Database
		sqlStatement := `
			INSERT INTO buildpipeline (id, sha, ref, status, web_url, kibana_url)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id) DO UPDATE
			SET status = $4
			RETURNING id`
		id := 0
		err = database.Db.QueryRow(sqlStatement, mayaObj[i].ID, mayaObj[i].Sha, mayaObj[i].Ref, mayaObj[i].Status, mayaObj[i].WebURL, "/").Scan(&id)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("New record ID for jiva build Pipeline:", id)

		// Push Aks jobs data to Database
		for j := range mayaJobsData {
			sqlStatement := `
				INSERT INTO buildjobs (pipelineid, id, status, stage, name, ref, created_at, started_at, finished_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				ON CONFLICT (id) DO UPDATE
				SET status = $3, stage = $4, name = $5, ref = $6, created_at = $7, started_at = $8, finished_at = $9
				RETURNING id`
			id := 0
			err = database.Db.QueryRow(sqlStatement,
				mayaObj[i].ID,
				mayaJobsData[j].ID,
				mayaJobsData[j].Status,
				mayaJobsData[j].Stage,
				mayaJobsData[j].Name,
				mayaJobsData[j].Ref,
				mayaJobsData[j].CreatedAt,
				mayaJobsData[j].StartedAt,
				mayaJobsData[j].FinishedAt,
			).Scan(&id)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("New record ID for maya pipeline Jobs: ", id)
		}
	}
}

// QueryBuildData first fetches the dashboard data from the db
func QueryBuildData(datas *dashboard) error {
	pipelinerows, err := database.Db.Query(`SELECT * FROM buildpipeline ORDER BY id DESC`)
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

		jobsquery := `SELECT * FROM buildjobs WHERE pipelineid = $1 ORDER BY id`
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
