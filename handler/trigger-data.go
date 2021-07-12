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

// QueryData fetches the builddashboard data from the db
func QueryData(datas *Openshiftdashboard, pipelineTable string, jobsTable string) error {
	pipelineQuery := fmt.Sprintf("SELECT * FROM %s ORDER BY id DESC LIMIT 20;", pipelineTable)
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
			&pipelinedata.CreatedAt,
		)
		if err != nil {
			return err
		}
		jobsquery := fmt.Sprintf("SELECT pipelineid, id, status , stage , name , ref ,github_readme, created_at , started_at , finished_at, author_name  FROM %s WHERE pipelineid = $1 ORDER BY id;", jobsTable)
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
		datas.Dashboard = append(datas.Dashboard, pipelinedata)
	}
	err = pipelinerows.Err()
	if err != nil {
		return err
	}
	return nil
}

func getPipelineData(token, project, branch string) (Pipeline, error) {
	URL := "https://gitlab.openebs.ci/api/v4/projects/" + project + "/pipelines?ref=" + branch
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

func releasePipelineJobs(pipelineID int, token string, project string) (Jobs, error) {
	// Generate pipeline jobs api url using BaseURL, pipelineID and OPENSHIFTID
	urlTmp := BaseURL + "api/v4/projects/" + project + "/pipelines/" + strconv.Itoa(pipelineID) + "/jobs?page="
	var obj Jobs
	for i := 1; ; i++ {
		url := urlTmp + strconv.Itoa(i)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Close = true
		// Set header for api request
		req.Header.Set("Connection", "close")
		req.Header.Add("PRIVATE-TOKEN", token)
		client := http.Client{
			Timeout: time.Minute * time.Duration(2),
		}
		res, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)
		if string(body) == "[]" {
			break
		}
		var tmpObj Jobs
		err = json.Unmarshal(body, &tmpObj)
		// glog.Infoln("error ", err)
		obj = append(obj, tmpObj...)
	}
	return obj, nil
}

// openshiftCommit from gitlab api and store to database
func getPlatformData(token, project, branch, pipelineTable, jobTable, releaseTagJob string) {
	pipelineData, err := getPipelineData(token, project, branch)
	if err != nil {
		glog.Error(err)
		return
	}
	for i := range pipelineData {
		var imageTag, getURLString, createdAt string
		checkPipelinePresent := CheckExists{
			Id:        pipelineData[i].ID,
			TableName: pipelineTable,
		}
		if !CheckUpdateRequire(checkPipelinePresent) {
			glog.Infoln("%d pipeline update not require", checkPipelinePresent.Id)
			return
		}

		pipelineJobsData, err := releasePipelineJobs(pipelineData[i].ID, token, project)
		if err != nil {
			glog.Error(err)
			return
		}
		glog.Infoln("pipelieID :->  " + strconv.Itoa(pipelineData[i].ID) + " || JobSLegth :-> " + strconv.Itoa(len(pipelineJobsData)))
		//check is releaseTag already exixts on pipelineTable
		checkImageTag := VerifyImageTagExists(ImageTagCheck{pipelineTable, pipelineData[i].ID})
		if checkImageTag == "NA" { //if tag not exists (NA) then check and fetch tag from jobLogs
			imageTag, err = getImageTag(pipelineJobsData, token, project, branch, releaseTagJob)
			if err != nil {
				glog.Error(err)
			}
		} else {
			imageTag = checkImageTag //Tag exixts rewrite the same tag again
		}
		if len(pipelineJobsData) != 0 {
			createdAt = pipelineJobsData[0].CreatedAt
		}
		// Add pipelines data to Database
		sqlStatement := fmt.Sprintf("INSERT INTO %s (project, id, sha, ref, status, web_url, openshift_pid, release_tag, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)"+
			"ON CONFLICT (id) DO UPDATE SET status = $5, openshift_pid = $7, release_tag = $8, created_at = $9 RETURNING id;", pipelineTable)
		id := 0
		err = database.Db.QueryRow(sqlStatement,
			project,
			pipelineData[i].ID,
			pipelineData[i].Sha,
			pipelineData[i].Ref,
			pipelineData[i].Status,
			pipelineData[i].WebURL,
			pipelineData[i].ID,
			imageTag,
			createdAt,
		).Scan(&id)
		if err != nil {
			glog.Error(err)
		}
		glog.Infof("New record ID for %s Pipeline: %d CreatedAt : %s", project, id, createdAt)

		// Add pipeline jobs data to Database
		for j := range pipelineJobsData {
			getURLString, err = getURL(pipelineJobsData[j].WebURL, token)
			if err != nil {
				glog.Error("error in getting JobUrl", err)
			}
			sqlStatement := fmt.Sprintf("INSERT INTO %s (pipelineid, id, status, stage, name, ref, github_readme, created_at, started_at, finished_at, author_name ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)"+
				"ON CONFLICT (id) DO UPDATE SET status = $3, stage = $4, name = $5, ref = $6, github_readme = $7, created_at = $8, started_at = $9, finished_at = $10, author_name = $11 RETURNING id;", jobTable)
			id := 0
			err = database.Db.QueryRow(sqlStatement,
				pipelineData[i].ID,
				pipelineJobsData[j].ID,
				pipelineJobsData[j].Status,
				pipelineJobsData[j].Stage,
				pipelineJobsData[j].Name,
				pipelineJobsData[j].Ref,
				getURLString,
				pipelineJobsData[j].CreatedAt,
				pipelineJobsData[j].StartedAt,
				pipelineJobsData[j].FinishedAt,
				pipelineJobsData[j].WebURL,
			).Scan(&id)
			if err != nil {
				glog.Error(err)
			}
			glog.Infof("New record ID for %s pipeline Jobs: %d", project, id)
		}
	}
}

func getImageTag(jobsData Jobs, token, project, branch, releaseTagJob string) (string, error) {
	var jobURL, tagJob string
	tagJob = releaseTagJob
	glog.Infoln(fmt.Sprintf("\n platform : %s \n branch : %s \n tagJob : %s \n", project, branch, tagJob))
	for _, value := range jobsData {
		if value.Name == tagJob {
			jobURL = value.WebURL + "/raw"
		}
	}
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
	re := regexp.MustCompile("releaseTag[^ ]*")
	value := re.FindString(data)
	result := strings.Split(string(value), "=")
	if result != nil && len(result) > 1 {
		if result[1] == "" {
			return "NA", nil
		}
		releaseVersion := strings.Split(result[1], "\n")
		return releaseVersion[0], nil
	}
	return "NA", nil
}

func getURL(jobsData string, token string) (string, error) {
	var jobURL string
	var baseURL string
	var searchPlatform string
	if strings.Contains(jobsData, "konvoy") {
		baseURL = "https://github.com/mayadata-io/e2e-konvoy/tree/master/"
		searchPlatform = "openebs-konvoy-e2e/"

	} else if strings.Contains(jobsData, "openshift") {
		baseURL = "https://github.com/mayadata-io/e2e-openshift/tree/master/"
		searchPlatform = "Openshift-EE/"
	} else {
		return "NA", nil
	}
	jobURL = jobsData + "/raw"
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
	re := regexp.MustCompile(searchPlatform + "[^ ]*")
	value := re.FindString(data)
	result := strings.Split(string(value), "\n")
	if result != nil && len(result) > 1 {
		if result[0] == "" {
			return "NA", nil
		}
		resultSlice := strings.Split(string(result[0]), "/")
		resultSlice[len(resultSlice)-1] = ""
		url := strings.Join(resultSlice, "/")
		return baseURL + url + "README.md", nil
	}
	return "NA", nil
}
