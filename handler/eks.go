package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/golang/glog"
	"github.com/openebs/ci-e2e-dashboard-go-backend/database"
)

// Ekshandler return eks pipeline data to /eks path
func Ekshandler(w http.ResponseWriter, r *http.Request) {
	// Allow cross origin request
	(w).Header().Set("Access-Control-Allow-Origin", "*")
	datas := dashboard{}
	err := QueryEksData(&datas)
	if err != nil {
		http.Error(w, err.Error(), 500)
		glog.Error(err)
	}
	out, err := json.Marshal(datas)
	if err != nil {
		http.Error(w, err.Error(), 500)
		glog.Error(err)
	}
	w.Write(out)
}

// EksData from gitlab api for Eks and dump to database
func EksData(token string) {
	eksPipelineID, err := database.Db.Query(`SELECT id,eks_trigger_pid FROM buildpipeline ORDER BY id DESC FETCH FIRST 20 ROWS ONLY`)
	if err != nil {
		glog.Error("EKS pipeline quering data Error:", err)
		return
	}
	for eksPipelineID.Next() {
		var logURL string
		pipelinedata := TriggredID{}
		err = eksPipelineID.Scan(
			&pipelinedata.BuildPID,
			&pipelinedata.ID,
		)
		eksPipelineData, err := eksPipeline(token, pipelinedata.ID)
		if err != nil {
			glog.Error(err)
			return
		}
		pipelineJobsdata, err := eksPipelineJobs(eksPipelineData.ID, token)
		if err != nil {
			glog.Error(err)
			return
		}
		if pipelinedata.ID != 0 {
			jobStartedAt := pipelineJobsdata[0].StartedAt
			JobFinishedAt := pipelineJobsdata[len(pipelineJobsdata)-1].FinishedAt
			logURL = Kibanaloglink(eksPipelineData.Sha, eksPipelineData.ID, eksPipelineData.Status, jobStartedAt, JobFinishedAt)
		}
		sqlStatement := `
			INSERT INTO ekspipeline (build_pipelineid, id, sha, ref, status, web_url, kibana_url)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (build_pipelineid) DO UPDATE
			SET id = $2, sha = $3, ref = $4, status = $5, web_url = $6, kibana_url = $7
			RETURNING id`
		id := 0
		err = database.Db.QueryRow(sqlStatement,
			pipelinedata.BuildPID,
			eksPipelineData.ID,
			eksPipelineData.Sha,
			eksPipelineData.Ref,
			eksPipelineData.Status,
			eksPipelineData.WebURL,
			logURL,
		).Scan(&id)
		if err != nil {
			glog.Error(err)
		}
		glog.Infoln("New record ID for EKS Pipeline:", id)
		if pipelinedata.ID != 0 {
			for j := range pipelineJobsdata {
				sqlStatement := `
					INSERT INTO eksjobs (pipelineid, id, status, stage, name, ref, created_at, started_at, finished_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
					ON CONFLICT (id) DO UPDATE
					SET status = $3, stage = $4, name = $5, ref = $6, created_at = $7, started_at = $8, finished_at = $9
					RETURNING id`
				id := 0
				err = database.Db.QueryRow(sqlStatement,
					eksPipelineData.ID,
					pipelineJobsdata[j].ID,
					pipelineJobsdata[j].Status,
					pipelineJobsdata[j].Stage,
					pipelineJobsdata[j].Name,
					pipelineJobsdata[j].Ref,
					pipelineJobsdata[j].CreatedAt,
					pipelineJobsdata[j].StartedAt,
					pipelineJobsdata[j].FinishedAt,
				).Scan(&id)
				if err != nil {
					glog.Error(err)
				}
				glog.Infof("New record ID for EKS Jobs: %s", id)
			}
		}
	}
}

// eksPipeline will get data from gitlab api and store to DB
func eksPipeline(token string, pipelineID int) (*PlatformPipeline, error) {
	dummyJSON := []byte(`{"id":0,"sha":"00000000000000000000","ref":"none","status":"none","web_url":"none"}`)
	if pipelineID == 0 {
		var obj PlatformPipeline
		json.Unmarshal(dummyJSON, &obj)
		return &obj, nil
	}
	// Store eks pipeline data form gitlab api to eksObj
	url := BaseURL + "api/v4/projects/" + EKSID + "/pipelines/" + strconv.Itoa(pipelineID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Close = true
	// Set header for api request
	req.Header.Set("Connection", "close")
	req.Header.Add("PRIVATE-TOKEN", token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	// Unmarshal response data
	var obj PlatformPipeline
	json.Unmarshal(body, &obj)
	return &obj, nil
}

// eksPipelineJobs will get pipeline jobs details from gitlab jobs api
func eksPipelineJobs(pipelineID int, token string) (Jobs, error) {
	// Generate pipeline jobs api url using BaseURL, pipelineID and EKSID
	if pipelineID == 0 {
		return nil, nil
	}
	url := BaseURL + "api/v4/projects/" + EKSID + "/pipelines/" + strconv.Itoa(pipelineID) + "/jobs?per_page=50"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Close = true
	// Set header for api request
	req.Header.Set("Connection", "close")
	req.Header.Add("PRIVATE-TOKEN", token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	// Unmarshal response data
	var obj Jobs
	json.Unmarshal(body, &obj)
	return obj, nil
}

// QueryEksData fetch the pipeline data as well as jobs data form eks table of database
func QueryEksData(datas *dashboard) error {
	// Select all data from packetpipeline table of DB
	pipelinerows, err := database.Db.Query(`SELECT id,sha,ref,status,web_url,kibana_url FROM ekspipeline ORDER BY build_pipelineid DESC`)
	if err != nil {
		return err
	}
	// Close DB connection after r/w operation
	defer pipelinerows.Close()
	// Iterate on each rows of pipeline table data for perform more operation related to pipeline Data
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
			return err
		}
		// Query eksjobs data of respective pipeline using pipelineID from eksjobs table
		jobsquery := `SELECT * FROM eksjobs WHERE pipelineid = $1 ORDER BY id`
		jobsrows, err := database.Db.Query(jobsquery, pipelinedata.ID)
		if err != nil {
			return err
		}
		// Close DB connection after r/w operation
		defer jobsrows.Close()
		jobsdataarray := []Jobssummary{}
		// Iterate on each rows of table data for perform more operation related to pipelineJobsData
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
				return err
			}
			// Append each row data to an array(jobsDataArray)
			jobsdataarray = append(jobsdataarray, jobsdata)
			// Add jobs details of pipeline into jobs field of pipelineData
			pipelinedata.Jobs = jobsdataarray
		}
		// Append each pipeline data to datas of field Dashobard
		datas.Dashboard = append(datas.Dashboard, pipelinedata)
	}
	err = pipelinerows.Err()
	if err != nil {
		return err
	}
	return nil
}
