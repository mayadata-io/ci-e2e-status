package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/mayadata-io/ci-e2e-status/database"
)

// OepPipelineHandler return packet pipeline data to /packet path
func OepPipelineHandler(w http.ResponseWriter, r *http.Request) {
	// Allow cross origin request
	(w).Header().Set("Access-Control-Allow-Origin", "*")
	datas := dashboard{}
	err := OepQueryPipelineData(&datas)
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

// OepQueryPipelineData fetch the pipeline data as well as jobs data form Packet table of database
func OepQueryPipelineData(datas *dashboard) error {
	// Select all data from packetpipeline table of DB
	query := fmt.Sprintf("SELECT pipelineid,sha,ref,status,web_url,author_name,author_email,message,percentage_coverage FROM oep_pipeline ORDER BY build_pipeline_id DESC")
	pipelinerows, err := database.Db.Query(query)
	if err != nil {
		return err
	}
	// Close DB connection after r/w operation
	defer pipelinerows.Close()
	// Iterate on each rows of pipeline table data for perform more operation related to pipeline Data
	for pipelinerows.Next() {
		pipelinedata := pipelineSummary{}
		err = pipelinerows.Scan(
			&pipelinedata.PipelineID,
			&pipelinedata.Sha,
			&pipelinedata.Ref,
			&pipelinedata.Status,
			&pipelinedata.WebURL,
			&pipelinedata.AuthorName,
			&pipelinedata.AuthorEmail,
			&pipelinedata.Message,
			&pipelinedata.Percentage,
		)
		if err != nil {
			return err
		}
		// Query packetjobs data of respective pipeline using pipelineID from packetjobs table
		jobsquery := fmt.Sprintf("SELECT pipelineid,id,status,stage,name,ref,created_at,started_at,finished_at FROM oep_pipeline_jobs WHERE pipelineid = $1 ORDER BY id;")
		jobsrows, err := database.Db.Query(jobsquery, pipelinedata.PipelineID)
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

// oepData from gitlab api for oep and dump to database
func goPipeOep(token string, triggerID string, pA string, pE string, pM string, buildID int, commitSha string, platform string, platformID int) {
	trID, err := strconv.Atoi(triggerID)
	glog.Infoln("OEp triggered by build id :", trID)
	oepPipelineData, err := oepPipeline(token, trID, platformID)
	if err != nil {
		glog.Error(err)
		return
	}
	pipelineJobsdata, err := oepPipelineJobs(oepPipelineData.ID, token, platformID)
	if err != nil {
		glog.Error(err)
		return
	}
	percentageCoverage, err := percentageCoverageFunc(pipelineJobsdata, token)
	if err != nil {
		glog.Error(err)
		return
	}

	sqlStatement := fmt.Sprintf(`INSERT INTO ` + platform + `_pipeline ( pipelineid, sha, ref, status, web_url, author_name, author_email, message, build_pipeline_id, percentage_coverage)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 ON CONFLICT (build_pipeline_id) DO UPDATE SET pipelineid = $1, sha = $2, ref = $3, status = $4, web_url = $5, percentage_coverage = $10 RETURNING build_pipeline_id;`)
	pipelineid := 0
	err = database.Db.QueryRow(sqlStatement,
		oepPipelineData.ID,
		commitSha,
		oepPipelineData.Ref,
		oepPipelineData.Status,
		oepPipelineData.WebURL,
		pA,
		pE,
		pM,
		buildID,
		percentageCoverage,
	).Scan(&pipelineid)
	if err != nil {
		glog.Error(err)
	}
	glog.Infoln("New record ID for build Triggered OEP Pipeline:", oepPipelineData.ID)

	// if pipelinedata.ID != 0 {
	for j := range pipelineJobsdata {
		// var jobLogURL string
		sqlStatement := fmt.Sprintf("INSERT INTO " + platform + "_pipeline_jobs (pipelineid, id, status, stage, name, ref, created_at, started_at, finished_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)" +
			"ON CONFLICT (id) DO UPDATE SET status = $3, stage = $4, name = $5, ref = $6, created_at = $7, started_at = $8, finished_at = $9 RETURNING id;")
		id := 0
		if len(pipelineJobsdata) != 0 {
			// jobLogURL = Kibanaloglink(oepPipelineData.Sha, oepPipelineData.ID, oepPipelineData.Status, pipelineJobsdata[j].StartedAt, pipelineJobsdata[j].FinishedAt)
		}
		err = database.Db.QueryRow(sqlStatement,
			oepPipelineData.ID,
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
		glog.Infoln("New record ID for OEP Jobs:", id)
	}
	// }
	// }
}

func percentageCoverageFunc(jobsData Jobs, token string) (string, error) {
	// var jobURL = "https://gitlab.mayadata.io/oep/oep-e2e-gcp/-/jobs/38871/raw"
	var jobURL string
	for _, value := range jobsData {
		if value.Name == "e2e-metrics" {
			jobURL = value.WebURL + "/raw"
		}
	}
	if jobURL != "" {
		req, err := http.NewRequest("GET", jobURL, nil)
		if err != nil {
			return "NA", err
		}
		req.Close = true
		req.Header.Set("Connection", "close")
		client := http.Client{
			Timeout: time.Minute * time.Duration(1),
		}
		res, err := client.Do(req)
		if err != nil {
			return "NA", err
		}
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)
		data := string(body)
		if data == "" {
			return "NA", err
		}
		re := regexp.MustCompile("coverage: [^ ]*")
		value := re.FindString(data)
		result := strings.Split(string(value), ":")
		if result != nil && len(result) > 1 {
			if result[1] == "" {
				return "NA", nil
			}
			releaseVersion := result[1]
			return releaseVersion, nil
		}
		return "NA", nil
	}
	return "NA", nil
}
