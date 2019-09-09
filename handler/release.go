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

func getReleaseData(token, branch string) (Pipeline, error) {
	URL := "https://gitlab.openebs.ci/api/v4/projects/36/pipelines?ref=release-branch"
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
	var obj Pipeline
	json.Unmarshal(data, &obj)
	return obj, nil
}

func releasePipelineJobs(pipelineID int, token string) (Jobs, error) {
	// Generate pipeline jobs api url using BaseURL, pipelineID and OPENSHIFTID
	url := BaseURL + "api/v4/projects/" + OPENSHIFTID + "/pipelines/" + strconv.Itoa(pipelineID) + "/jobs?per_page=100"
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

// openshiftCommit from gitlab api and store to database
func releaseBranch(token, project, branch, pipelineTable, jobTable string) {
	var logURL string
	releaseData, err := getReleaseData(token, branch)
	if err != nil {
		glog.Error(err)
		return
	}
	for i := range releaseData {
		// pipelineDetail, err := getTriggredPipelineDetail(releaseData[i].ID, token)
		// if err != nil {
		// 	glog.Error(err)
		// 	return
		// }
		// pipelineJobsData store the jobs details related to pipeline id
		pipelineJobsData, err := releasePipelineJobs(releaseData[i].ID, token)
		if err != nil {
			glog.Error(err)
			return
		}
		if len(pipelineJobsData) != 0 {
			jobStartedAt := pipelineJobsData[0].StartedAt
			JobFinishedAt := pipelineJobsData[len(pipelineJobsData)-1].FinishedAt
			logURL = Kibanaloglink(releaseData[i].Sha, releaseData[i].ID, releaseData[i].Status, jobStartedAt, JobFinishedAt)
		}
		var releaseTag = ""
		releaseTag, err = getReleaseTag(pipelineJobsData, token)
		if err != nil {
			glog.Error(err)
		}
		// Add pipelines data to Database
		sqlStatement := fmt.Sprintf("INSERT INTO %s (project, id, sha, ref, status, web_url, openshift_pid, kibana_url, release_tag) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)"+
			"ON CONFLICT (id) DO UPDATE SET status = $5, openshift_pid = $7, kibana_url = $8, release_tag = $9 RETURNING id;", pipelineTable)
		id := 0
		err = database.Db.QueryRow(sqlStatement,
			project,
			releaseData[i].ID,
			releaseData[i].Sha,
			releaseData[i].Ref,
			releaseData[i].Status,
			releaseData[i].WebURL,
			releaseData[i].ID,
			logURL,
			releaseTag,
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
				releaseData[i].ID,
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
