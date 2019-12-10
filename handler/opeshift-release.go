package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/mayadata-io/ci-e2e-status/database"
)

// OpenshiftHandlerRelease return eks pipeline data to /build path
func OpenshiftHandlerRelease(w http.ResponseWriter, r *http.Request) {
	// Allow cross origin request
	(w).Header().Set("Access-Control-Allow-Origin", "*")
	datas := Openshiftdashboard{}
	err := QueryReleasePipelineData(&datas, "release_pipeline_data", "release_jobs_data")
	if err != nil {
		http.Error(w, err.Error(), 500)
		glog.Error(err)
		return
	}
	out, err := json.Marshal(datas)
	if err != nil {
		http.Error(w, err.Error(), 500)
		glog.Error(err)
		return
	}
	w.Write(out)
}

// QueryReleasePipelineData fetches the builddashboard data from the db
func QueryReleasePipelineData(datas *Openshiftdashboard, pipelineTable string, jobsTable string) error {
	pipelineQuery := fmt.Sprintf("SELECT * FROM %s ORDER BY id DESC;", pipelineTable)
	pipelinerows, err := database.Db.Query(pipelineQuery)
	if err != nil {
		return err
	}
	defer pipelinerows.Close()
	for pipelinerows.Next() {
		pipelinedata := OpenshiftpipelineSummary{}
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
		jobsquery := fmt.Sprintf("SELECT pipelineid, id, status , stage , name , ref , created_at , started_at , finished_at  FROM %s WHERE pipelineid = $1 ORDER BY id;", jobsTable)
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
	url := BaseURL + "api/v4/projects/36/pipelines/" + strconv.Itoa(pipelineID) + "/jobs?per_page=100"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Close = true
	// Set header for api request
	req.Header.Set("Connection", "close")
	req.Header.Add("PRIVATE-TOKEN", token)
	client := http.Client{
		Timeout: time.Minute * time.Duration(1),
	}
	res, err := client.Do(req)
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
	var releaseTag string
	releaseData, err := getReleaseData(token, branch)
	if err != nil {
		glog.Error(err)
		return
	}
	for i := range releaseData {
		pipelineJobsData, err := releasePipelineJobs(releaseData[i].ID, token)
		if err != nil {
			glog.Error(err)
			return
		}
		glog.Infoln("pipelieID :->  " + strconv.Itoa(releaseData[i].ID) + " || JobSLegth :-> " + strconv.Itoa(len(pipelineJobsData)))
		if len(pipelineJobsData) != 0 {
			jobStartedAt := pipelineJobsData[0].StartedAt
			JobFinishedAt := pipelineJobsData[len(pipelineJobsData)-1].FinishedAt
			logURL = Kibanaloglink(releaseData[i].Sha, releaseData[i].ID, releaseData[i].Status, jobStartedAt, JobFinishedAt)
		}
		releaseTag, err = getReleaseTag(pipelineJobsData, token)
		if err != nil {
			glog.Error(err)
		}

		glog.Infoln("releaseTagFuc Result : - > : ", releaseTag)
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
			sqlStatement := fmt.Sprintf("INSERT INTO %s (pipelineid, id, status, stage, name, ref, created_at, started_at, finished_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)"+
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
			).Scan(&id)
			if err != nil {
				glog.Error(err)
			}
			glog.Infof("New record ID for %s pipeline Jobs: %d", project, id)
		}
	}
}

func getReleaseTag(jobsData Jobs, token string) (string, error) {
	var jobURL string
	for _, value := range jobsData {
		if value.Name == "K9YC-OpenEBS" {
			jobURL = value.WebURL + "/raw"
		}
	}
	// url := jobURL
	glog.Infoln("url----->", jobURL)
	req, err := http.NewRequest("GET", jobURL, nil)
	if err != nil {
		return "NA", err
	}
	req.Close = true
	req.Header.Set("Connection", "close")
	// req.Header.Add("PRIVATE-TOKEN", token
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
	glog.Infoln("result----->", result)
	if result != nil && len(result) > 1 {
		if result[1] == "" {
			return "NA", nil
		}
		result = strings.Split(result[1], ":")
		return result[1], nil
	}
	return "NA", nil
}
