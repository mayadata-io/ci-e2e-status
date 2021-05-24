package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/mayadata-io/ci-e2e-status/config"
	"github.com/mayadata-io/ci-e2e-status/database"
)

type Recent struct {
	Branch   string     `json:"branch"`
	Pipeline []PipeData `json:"pipelines"`
}
type RecentAPI struct {
	Recent []Recent `json:"recent"`
}

func HandleRecentPipelines(w http.ResponseWriter, r *http.Request) {
	// Allow cross origin request
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers:", "Origin, Content-Type, X-Auth-Token, Authorization")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	con := config.ReadConfig()
	pipe := []PipeData{}
	recPip := []Recent{}
	rec := Recent{}
	pi := PipeData{}
	rApi := RecentAPI{}
	for _, p := range con.Projects {
		if p.Name == "openshift" {
			for _, b := range p.Branches {
				pipe = nil
				err := GetPipeline(&pi, p.Name, strings.Replace(b.Name, "-", "_", -1), &pipe)
				if err != nil {
					glog.Infoln(err)
				}
				err = GetPipeline(&pi, "konvoy", strings.Replace(b.Name, "-", "_", -1), &pipe)
				if err != nil {
					glog.Infoln(err)
				}

				rec.Branch = b.Name
				rec.Pipeline = pipe
				recPip = append(recPip, rec)
			}
		} else if p.Name == "nativek8s" {
			for _, b := range p.Branches {
				pipe = nil
				err := GetPipeline(&pi, p.Name, strings.Replace(b.Name, "-", "_", -1), &pipe)
				if err != nil {
					glog.Infoln(err)
				}
				rec.Branch = b.Name
				rec.Pipeline = pipe
				recPip = append(recPip, rec)
			}
		}
	}
	rApi.Recent = append(rApi.Recent, recPip...)
	body, err := json.Marshal(rApi)
	if err != nil {
		http.Error(w, err.Error(), 500)
		glog.Error(err)
		return
	}
	w.Write(body)
}

// GetPipelineData to get perticular pipeline data with jobs
func GetPipeline(pipe *PipeData, platform, branch string, pipeArray *[]PipeData) error {
	pipelineQuery := fmt.Sprintf("SELECT * FROM %s WHERE id=(select max(id) from %s) ;", fmt.Sprintf("%s_%s", platform, branch), fmt.Sprintf("%s_%s", platform, branch))
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
		&pipelinedata.CreatedAt,
	)
	if err != nil {
		return err
	}

	jobsquery := fmt.Sprintf("SELECT pipelineid, id, status , stage , name , ref ,github_readme, created_at , started_at , finished_at, author_name  FROM %s WHERE pipelineid = $1 ORDER BY id;", fmt.Sprintf("%s_%s_jobs", platform, branch))
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
			&jobsdata.WebURL,
		)
		if err != nil {
			return err
		}
		jobsdataarray = append(jobsdataarray, jobsdata)
		pipelinedata.Jobs = jobsdataarray
	}
	pipe.Pipeline = pipelinedata
	*pipeArray = append(*pipeArray, *pipe)
	return nil
}
