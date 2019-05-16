package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/golang/glog"
	"github.com/mayadata-io/ci-e2e-status/database"
)

// OpenshiftHandlerMaster return openshift pipeline data to /openshift path
func OpenshiftHandlerMaster(w http.ResponseWriter, r *http.Request) {
	// Allow cross origin request
	(w).Header().Set("Access-Control-Allow-Origin", "*")
	datas := dashboard{}
	err := QueryOpenshiftData(&datas, "Openshift_pipeline", "Openshift_jobs")
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

// OpenshiftData from gitlab api for Openshift and dump to database
func OpenshiftData(token, triggredIDColumnName, pipelineTableName, jobTableName string) {
	query := fmt.Sprintf("SELECT id,%s FROM build_pipeline ORDER BY id DESC FETCH FIRST 20 ROWS ONLY;", triggredIDColumnName)
	openshiftPipelineID, err := database.Db.Query(query)
	if err != nil {
		glog.Error("OPENSHIFT pipeline quering data Error:", err)
		return
	}
	for openshiftPipelineID.Next() {
		var logURL string
		pipelineData := TriggredID{}
		err = openshiftPipelineID.Scan(
			&pipelineData.BuildPID,
			&pipelineData.ID,
		)
		defer openshiftPipelineID.Close()
		openshiftPipelineData, err := openshiftPipeline(token, pipelineData.ID)
		if err != nil {
			glog.Error(err)
			return
		}
		pipelineJobsData, err := openshiftPipelineJobs(openshiftPipelineData.ID, token)
		if err != nil {
			glog.Error(err)
			return
		}
		if pipelineData.ID != 0 && len(pipelineJobsData) != 0 {
			jobStartedAt := pipelineJobsData[0].StartedAt
			JobFinishedAt := pipelineJobsData[len(pipelineJobsData)-1].FinishedAt
			logURL = Kibanaloglink(openshiftPipelineData.Sha, openshiftPipelineData.ID, openshiftPipelineData.Status, jobStartedAt, JobFinishedAt)
		}
		sqlStatement := fmt.Sprintf("INSERT INTO %s (build_pipelineid, id, sha, ref, status, web_url, kibana_url) VALUES ($1, $2, $3, $4, $5, $6, $7)"+
			"ON CONFLICT (build_pipelineid) DO UPDATE SET id = $2, sha = $3, ref = $4, status = $5, web_url = $6, kibana_url = $7 RETURNING id;", pipelineTableName)
		id := 0
		err = database.Db.QueryRow(sqlStatement,
			pipelineData.BuildPID,
			openshiftPipelineData.ID,
			openshiftPipelineData.Sha,
			openshiftPipelineData.Ref,
			openshiftPipelineData.Status,
			openshiftPipelineData.WebURL,
			logURL,
		).Scan(&id)
		if err != nil {
			glog.Error(err)
		}
		glog.Infoln("New record ID for OPENSHIFT Pipeline:", id)
		if pipelineData.ID != 0 {
			for j := range pipelineJobsData {
				sqlStatement := fmt.Sprintf("INSERT INTO %s (pipelineid, id, status, stage, name, ref, created_at, started_at, finished_at, message, author_name) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)"+
					"ON CONFLICT (id) DO UPDATE SET status = $3, stage = $4, name = $5, ref = $6, created_at = $7, started_at = $8, finished_at = $9 RETURNING id;", jobTableName)
				id := 0
				err = database.Db.QueryRow(sqlStatement,
					openshiftPipelineData.ID,
					pipelineJobsData[j].ID,
					pipelineJobsData[j].Status,
					pipelineJobsData[j].Stage,
					pipelineJobsData[j].Name,
					pipelineJobsData[j].Ref,
					pipelineJobsData[j].CreatedAt,
					pipelineJobsData[j].StartedAt,
					pipelineJobsData[j].FinishedAt,
					pipelineJobsData[j].Commit.Message,
					pipelineJobsData[j].Commit.AuthorName,
				).Scan(&id)
				if err != nil {
					glog.Error(err)
				}
				glog.Infoln("New record ID for OPENSHIFT Jobs:", id)
			}
		}
	}
}

// openshiftPipeline will get data from gitlab api and store to DB
func openshiftPipeline(token string, pipelineID int) (*PlatformPipeline, error) {
	dummyJSON := []byte(`{"id":0,"sha":"00000000000000000000","ref":"none","status":"none","web_url":"none"}`)
	if pipelineID == 0 {
		var obj PlatformPipeline
		json.Unmarshal(dummyJSON, &obj)
		return &obj, nil
	}
	// Store openshift pipeline data form gitlab api to openshiftObj
	url := BaseURL + "api/v4/projects/" + OPENSHIFTID + "/pipelines/" + strconv.Itoa(pipelineID)
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

// openshiftPipelineJobs will get pipeline jobs details from gitlab jobs api
func openshiftPipelineJobs(pipelineID int, token string) (Jobs, error) {
	// Generate pipeline jobs api url using BaseURL, pipelineID and OPENSHIFTID
	if pipelineID == 0 {
		return nil, nil
	}
	url := BaseURL + "api/v4/projects/" + OPENSHIFTID + "/pipelines/" + strconv.Itoa(pipelineID) + "/jobs?per_page=50"
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

// QueryOpenshiftData fetch the pipeline data as well as jobs data form Openshift table of database
func QueryOpenshiftData(datas *dashboard, pipelineTableName string, jobTableName string) error {
	// Select all data from Openshiftpipeline table of DB
	query := fmt.Sprintf("SELECT id,sha,ref,status,web_url,kibana_url FROM %s ORDER BY build_pipelineid DESC;", pipelineTableName)
	pipelinerows, err := database.Db.Query(query)
	if err != nil {
		return err
	}
	// Close DB connection after r/w operation
	defer pipelinerows.Close()
	// Iterate on each rows of pipeline table data for perform more operation related to pipeline Data
	for pipelinerows.Next() {
		pipelineData := pipelineSummary{}
		err = pipelinerows.Scan(
			&pipelineData.ID,
			&pipelineData.Sha,
			&pipelineData.Ref,
			&pipelineData.Status,
			&pipelineData.WebURL,
			&pipelineData.LogURL,
		)
		if err != nil {
			return err
		}
		// Query Openshiftjobs data of respective pipeline using pipelineID from Openshiftjobs table
		jobsQuery := fmt.Sprintf("SELECT * FROM %s WHERE pipelineid = $1 ORDER BY id;", jobTableName)
		jobsRows, err := database.Db.Query(jobsQuery, pipelineData.ID)
		if err != nil {
			return err
		}
		// Close DB connection after r/w operation
		defer jobsRows.Close()
		jobsDataArray := []Jobssummary{}
		// Iterate on each rows of table data for perform more operation related to pipelineJobsData
		for jobsRows.Next() {
			jobsData := Jobssummary{}
			err = jobsRows.Scan(
				&jobsData.PipelineID,
				&jobsData.ID,
				&jobsData.Status,
				&jobsData.Stage,
				&jobsData.Name,
				&jobsData.Ref,
				&jobsData.CreatedAt,
				&jobsData.StartedAt,
				&jobsData.FinishedAt,
				&jobsData.Message,
				&jobsData.AuthorName,
			)
			if err != nil {
				return err
			}
			// Append each row data to an array(jobsDataArray)
			jobsDataArray = append(jobsDataArray, jobsData)
			// Add jobs details of pipeline into jobs field of pipelineData
			pipelineData.Jobs = jobsDataArray
		}
		// Append each pipeline data to datas of field Dashobard
		datas.Dashboard = append(datas.Dashboard, pipelineData)
	}
	err = pipelinerows.Err()
	if err != nil {
		return err
	}
	return nil
}
