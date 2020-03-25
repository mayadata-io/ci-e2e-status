package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/golang/glog"
	"github.com/mayadata-io/ci-e2e-status/database"
)

//KonvoyPipelineHandler to export data in /api/pipeline/konvoy endpoint
func KonvoyPipelineHandler(w http.ResponseWriter, r *http.Request) {
	// Allow cross origin request
	(w).Header().Set("Access-Control-Allow-Origin", "*")
	datas := dashboard{}
	err := KonvoyQueryPipelineData(&datas)
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

// KonvoyQueryPipelineData fetch the pipeline data as well as jobs data form Packet table of database
func KonvoyQueryPipelineData(datas *dashboard) error {
	// Select all data from packetpipeline table of DB
	query := fmt.Sprintf("SELECT pipelineid,sha,ref,status,web_url,author_name,author_email,message,percentage_coverage FROM konvoy_pipeline ORDER BY build_pipeline_id DESC")
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
		jobsquery := fmt.Sprintf("SELECT pipelineid,id,status,stage,name,ref,created_at,started_at,finished_at FROM konvoy_pipeline_jobs WHERE pipelineid = $1 ORDER BY id;")
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
