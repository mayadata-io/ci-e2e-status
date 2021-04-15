package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/mayadata-io/ci-e2e-status/database"
)

// GetPipelineDetails return packet pipeline data to /packet path
func GetPipelineDataAPI(w http.ResponseWriter, r *http.Request) {
	// Allow cross origin request
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers:", "Origin, Content-Type, X-Auth-Token, Authorization")
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	platform := vars["platform"]                            //openshift
	branch := strings.Replace(vars["branch"], "-", "_", -1) // openebs-jiva
	id := vars["id"]
	glog.Infoln(fmt.Sprintf("\n\n Platform : %s \n branch : %s \n id : %s \n", platform, branch, id))
	pipe := PipeData{}
	err := GetPipelineData(&pipe, platform, branch, id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		glog.Error(err)
		return
	}
	body, err := json.Marshal(pipe)
	if err != nil {
		http.Error(w, err.Error(), 500)
		glog.Error(err)
		return
	}
	w.Write(body)

}

// GetPipelineData to get perticular pipeline data with jobs
func GetPipelineData(pipe *PipeData, platform, branch, id string) error {
	pipelineQuery := fmt.Sprintf("SELECT * FROM %s WHERE id=%s ;", fmt.Sprintf("%s_%s", platform, branch), id)
	glog.Infoln("\n \t %s \n", pipelineQuery)
	row := database.Db.QueryRow(pipelineQuery)
	pipelinedata := OpenshiftpipelineSummary{}
	err := row.Scan(
		&pipelinedata.Project,
		&pipelinedata.ID,
		&pipelinedata.Sha,
		&pipelinedata.Ref,
		&pipelinedata.Status,
		&pipelinedata.WebURL,
		&pipelinedata.OpenshiftPID,
		&pipelinedata.LogURL,
		&pipelinedata.ReleaseTag,
	)
	if err != nil {
		return err
	}

	jobsquery := fmt.Sprintf("SELECT pipelineid, id, status , stage , name , ref ,github_readme, created_at , started_at , finished_at  FROM %s WHERE pipelineid = $1 ORDER BY id;", fmt.Sprintf("%s_%s_jobs", platform, branch))
	jobsrows, err := database.Db.Query(jobsquery, pipelinedata.ID)
	if err != nil {
		return err
	}
	defer jobsrows.Close()
	jobsdataarray := []BuildJobssummary{}
	for jobsrows.Next() {
		jobsdata := BuildJobssummary{}
		err = jobsrows.Scan(
			&jobsdata.PipelineID,
			&jobsdata.ID,
			&jobsdata.Status,
			&jobsdata.Stage,
			&jobsdata.Name,
			&jobsdata.Ref,
			&jobsdata.GithubReadme,
			&jobsdata.CreatedAt,
			&jobsdata.StartedAt,
			&jobsdata.FinishedAt,
		)
		if err != nil {
			return err
		}
		jobsdataarray = append(jobsdataarray, jobsdata)
		pipelinedata.Jobs = jobsdataarray
	}
	pipe.Pipeline = pipelinedata
	return nil
}
