package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"

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
		if branch == "release-branch" {
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
				pipelineDetail.ID,
				pipelineDetail.Sha,
				pipelineDetail.Ref,
				pipelineDetail.Status,
				pipelineDetail.WebURL,
				pipelineDetail.ID,
				logURL,
				releaseTag,
			).Scan(&id)
			if err != nil {
				glog.Error(err)
			}
			glog.Infof("New record ID for %s Pipeline: %d", project, id)
		} else {
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
		}

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
	URL := BaseURL + "api/v4/projects/" + OPENSHIFTID + "/repository/commits?per_page=50&&ref_name=" + branch
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

func getReleaseTag(jobsData Jobs, token string) (string, error) {
	var jobURL string
	for _, value := range jobsData {
		if value.Name == "K9YC-OpenEBS" {
			jobURL = value.WebURL
		}
	}
	url := jobURL + "/raw"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "NA", err
	}
	req.Close = true
	req.Header.Set("Connection", "close")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "NA", err
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	data := string(body)
	if data == "" {
		return "NA", err
	}
	grep := exec.Command("grep", "-oP", "(?<=openebs/m-apiserver)[^ ]*")
	ps := exec.Command("echo", data)

	// Get ps's stdout and attach it to grep's stdin.
	pipe, _ := ps.StdoutPipe()
	defer pipe.Close()
	grep.Stdin = pipe
	ps.Start()

	// Run and get the output of grep.
	value, _ := grep.Output()
	result := strings.Split(string(value), "\n")
	result = strings.Split(result[1], ":")
	if result[1] == "" {
		return "NA", nil
	}
	return result[1], nil
}

// QueryReleasePipelineData fetches the builddashboard data from the db
func QueryReleasePipelineData(datas *Builddashboard, pipelineTable string, jobsTable string) error {
	pipelineQuery := fmt.Sprintf("SELECT * FROM %s ORDER BY id DESC;", pipelineTable)
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
			&pipelinedata.ReleaseTag,
		)
		if err != nil {
			return err
		}

		jobsquery := fmt.Sprintf("SELECT * FROM %s WHERE pipelineid = $1 ORDER BY id;", jobsTable)
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
