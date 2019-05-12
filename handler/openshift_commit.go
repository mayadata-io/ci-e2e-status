package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
	"github.com/openebs/ci-e2e-status/database"
)

// openshiftCommit from gitlab api and store to database
func openshiftCommit(token, project, branch, pipelineTable, jobTable string) {
	commitData, err := getcommitData(token, branch)
	if err != nil {
		glog.Error(err)
		return
	}
	for i := range commitData {
		pipelineDetail, err := getTriggredPipelineDetail(commitData[i].ID, token)
		if err != nil {
			glog.Error(err)
			return
		}
		// pipelineJobsData store the jobs details related to pipeline id
		pipelineJobsData, err := openshiftPipelineJobs(pipelineDetail.ID, token)
		if err != nil {
			glog.Error(err)
			return
		}
		// Add pipelines data to Database
		sqlStatement := fmt.Sprintf("INSERT INTO %s (project, id, sha, ref, status, web_url, openshift_pid) VALUES ($1, $2, $3, $4, $5, $6, $7)"+
			"ON CONFLICT (id) DO UPDATE SET status = $5, openshift_pid = $7 RETURNING id;", pipelineTable)
		id := 0
		err = database.Db.QueryRow(sqlStatement,
			project,
			pipelineDetail.ID,
			pipelineDetail.Sha,
			pipelineDetail.Ref,
			pipelineDetail.Status,
			pipelineDetail.WebURL,
			pipelineDetail.ID,
		).Scan(&id)
		if err != nil {
			glog.Error(err)
		}
		glog.Infof("New record ID for %s Pipeline: %d", project, id)

		// Add pipeline jobs data to Database
		for j := range pipelineJobsData {
			sqlStatement := fmt.Sprintf("INSERT INTO %s (pipelineid, id, status, stage, name, ref, created_at, started_at, finished_at, message, author_name) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)"+
				"ON CONFLICT (id) DO UPDATE SET status = $3, stage = $4, name = $5, ref = $6, created_at = $7, started_at = $8, finished_at = $9 RETURNING id;", jobTable)
			id := 0
			err = database.Db.QueryRow(sqlStatement,
				pipelineDetail.ID,
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
			glog.Infof("New record ID for %s pipeline Jobs: %d", project, id)
		}
	}
}

// getcommitData will fetch the commit data from gitlab API
func getcommitData(token, branch string) (commit, error) {
	URL := BaseURL + "api/v4/projects/" + OPENSHIFTID + "/repository/commits?ref_name=" + branch
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, err
	}
	req.Close = true
	req.Header.Set("Connection", "close")
	req.Header.Add("PRIVATE-TOKEN", token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var obj commit
	json.Unmarshal(data, &obj)
	return obj, nil
}

// pipelineJobsData will get pipeline jobs details from gitlab api
func getTriggredPipelineDetail(id, token string) (PlatformPipeline, error) {
	url := BaseURL + "api/v4/projects/" + OPENSHIFTID + "/repository/commits/" + id
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		q := PlatformPipeline{}
		return q, err
	}
	req.Close = true
	req.Header.Set("Connection", "close")
	req.Header.Add("PRIVATE-TOKEN", token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		q := PlatformPipeline{}
		return q, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		q := PlatformPipeline{}
		return q, err
	}
	var obj commitPipeline
	json.Unmarshal(body, &obj)
	return obj.LastPipeline, nil
}
