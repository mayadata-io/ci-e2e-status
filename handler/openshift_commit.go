package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
	"github.com/mayadata-io/ci-e2e-status/database"
)

// openshiftCommit from gitlab api and store to database
func openshiftCommit(token, project, branch, pipelineTable, jobTable string) {
	var logURL string
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
		if pipelineDetail.ID != 0 && len(pipelineJobsData) != 0 {
			jobStartedAt := pipelineJobsData[0].StartedAt
			JobFinishedAt := pipelineJobsData[len(pipelineJobsData)-1].FinishedAt
			logURL = Kibanaloglink(pipelineDetail.Sha, pipelineDetail.ID, pipelineDetail.Status, jobStartedAt, JobFinishedAt)
		}
		// Add pipelines data to Database
		sqlStatement := fmt.Sprintf("INSERT INTO %s (project, id, sha, ref, status, web_url, openshift_pid, kibana_url) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)"+
			"ON CONFLICT (id) DO UPDATE SET status = $5, openshift_pid = $7, kibana_url = $8 RETURNING id;", pipelineTable)
		id := 0
		err = database.Db.QueryRow(sqlStatement,
			project,
			pipelineDetail.ID,
			pipelineDetail.Sha,
			pipelineDetail.Ref,
			pipelineDetail.Status,
			pipelineDetail.WebURL,
			pipelineDetail.ID,
			logURL,
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

// QueryOpenshiftReleaseData fetch the pipeline data as well as jobs data form Openshift table of database
func QueryOpenshiftReleaseData(datas *Builddashboard, pipelineTableName string, jobTableName string) error {
	pipelineQuery := fmt.Sprintf("SELECT * FROM %s ORDER BY id DESC;", pipelineTableName)
	pipelinerows, err := database.Db.Query(pipelineQuery)
	if err != nil {
		return err
	}
	defer pipelinerows.Close()
	for pipelinerows.Next() {
		pipelinedata := BuildpipelineSummary{}
		err = pipelinerows.Scan(
			&pipelinedata.Project,
			&pipelinedata.ID,
			&pipelinedata.Sha,
			&pipelinedata.Ref,
			&pipelinedata.Status,
			&pipelinedata.WebURL,
			&pipelinedata.OpenshiftPID,
			&pipelinedata.LogURL,
		)
		if err != nil {
			return err
		}

		jobsquery := fmt.Sprintf("SELECT * FROM %s WHERE pipelineid = $1 ORDER BY id;", jobTableName)
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
				&jobsdata.CreatedAt,
				&jobsdata.StartedAt,
				&jobsdata.FinishedAt,
				&jobsdata.Message,
				&jobsdata.AuthorName,
			)
			if err != nil {
				return err
			}
			jobsdataarray = append(jobsdataarray, jobsdata)
			pipelinedata.Jobs = jobsdataarray
		}
		datas.Dashboard = append(datas.Dashboard, pipelinedata)
	}
	err = pipelinerows.Err()
	if err != nil {
		return err
	}
	return nil
}
