package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/openebs/ci-e2e-status/database"
)

// Buildhandler return eks pipeline data to /build path
func Buildhandler(w http.ResponseWriter, r *http.Request) {
	// Allow cross origin request
	(w).Header().Set("Access-Control-Allow-Origin", "*")
	datas := Builddashboard{}
	err := queryBuildData(&datas)
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

// BuildData from gitlab api and store to database
func BuildData(token, project string) {
	pipelineData, err := getPipelineData(project, token)
	if err != nil {
		glog.Error(err)
		return
	}
	for i := range pipelineData {
		pipelineJobsData, err := getPipelineJobsData(pipelineData[i].ID, token, project)
		if err != nil {
			glog.Error(err)
			return
		}
		// Getting webURL link for getting triggredID
		baselineJobsWebURL := getBaselineJobWebURL(pipelineJobsData)
		// Get Openshift, Triggred pipeline ID for sepecified project
		openshiftPID, err := getTriggerPipelineid(baselineJobsWebURL, "e2e-openshift")
		if err != nil {
			glog.Error(err)
		}
		// Add pipelines data to Database
		sqlStatement := `
			INSERT INTO build_pipeline (project, id, sha, ref, status, web_url, openshift_pid)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id) DO UPDATE
			SET status = $5, openshift_pid = $7
			RETURNING id`
		id := 0
		err = database.Db.QueryRow(sqlStatement,
			project,
			pipelineData[i].ID,
			pipelineData[i].Sha,
			pipelineData[i].Ref,
			pipelineData[i].Status,
			pipelineData[i].WebURL,
			openshiftPID,
		).Scan(&id)
		if err != nil {
			glog.Error(err)
		}
		glog.Infof("New record ID for %s Pipeline: %d", project, id)

		// Add pipeline jobs data to Database
		for j := range pipelineJobsData {
			sqlStatement := `
				INSERT INTO build_jobs (pipelineid, id, status, stage, name, ref, created_at, started_at, finished_at, message, author_name)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
				ON CONFLICT (id) DO UPDATE
				SET status = $3, stage = $4, name = $5, ref = $6, created_at = $7, started_at = $8, finished_at = $9
				RETURNING id`
			id := 0
			err = database.Db.QueryRow(sqlStatement,
				pipelineData[i].ID,
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
	err = modifyBuildData()
	if err != nil {
		glog.Error(err)
	}
}

func modifyBuildData() error {
	query, err := database.Db.Query(`DELETE FROM build_pipeline WHERE id < (SELECT id FROM build_pipeline ORDER BY id DESC LIMIT 1 OFFSET 19)`)
	if err != nil {
		return err
	}
	defer query.Close()
	return nil
}

// queryBuildData fetches the builddashboard data from the db
func queryBuildData(datas *Builddashboard) error {
	pipelinerows, err := database.Db.Query(`SELECT * FROM build_pipeline ORDER BY id DESC`)
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
		)
		if err != nil {
			return err
		}

		jobsquery := `SELECT * FROM build_jobs WHERE pipelineid = $1 ORDER BY id`
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

// getTriggerPipelineid wil fetch the triggred pipeline ID using filter of raw file
func getTriggerPipelineid(jobURL, filter string) (string, error) {
	url := jobURL + "/raw"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Close = true
	req.Header.Set("Connection", "close")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	data := string(body)
	if data == "" {
		return "0", nil
	}
	grep := exec.Command("grep", "-oP", "(?<="+filter+"/pipelines/)[^ ]*")
	ps := exec.Command("echo", data)

	// Get ps's stdout and attach it to grep's stdin.
	pipe, _ := ps.StdoutPipe()
	defer pipe.Close()
	grep.Stdin = pipe
	ps.Start()

	// Run and get the output of grep.
	value, _ := grep.Output()
	result := strings.Split(string(value), "\"")
	if result[0] == "" {
		return "0", nil
	}
	return result[0], nil
}

// pipelineJobsData will get pipeline jobs details from gitlab api
func getPipelineJobsData(id int, token string, project string) (BuildJobs, error) {
	url := jobURLGenerator(id, project)
	req, err := http.NewRequest("GET", url, nil)
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
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var obj BuildJobs
	json.Unmarshal(body, &obj)
	return obj, nil
}

// pipelineData will fetch the data from gitlab API
func getPipelineData(project, token string) (Pipeline, error) {
	URL := pipelineURLGenerator(project)
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

// genearete pipeline url according to project name
func pipelineURLGenerator(project string) string {
	var projectID, Branch string
	if project == "maya" {
		projectID = MAYAID
		Branch = MAYABRANCH
	} else if project == "jiva" {
		projectID = JIVAID
		Branch = JIVABRANCH
	} else if project == "istgt" {
		projectID = ISTGTID
		Branch = ISTGTBRANCH
	} else if project == "zfs" {
		projectID = ZFSID
		Branch = ZFSBRANCH
	}
	generatedURL := BaseURL + "api/v4/projects/" + projectID + "/pipelines?ref=" + Branch
	return generatedURL
}

// genearete pipeline job url according to project name
func jobURLGenerator(id int, project string) string {
	var projectID string
	if project == "maya" {
		projectID = MAYAID
	} else if project == "jiva" {
		projectID = JIVAID
	} else if project == "istgt" {
		projectID = ISTGTID
	} else if project == "zfs" {
		projectID = ZFSID
	}
	generatedURL := BaseURL + "api/v4/projects/" + projectID + "/pipelines/" + strconv.Itoa(id) + "/jobs?per_page=50"
	return generatedURL
}

// Generate joburl of baseline stage
func getBaselineJobWebURL(data BuildJobs) string {
	var maxJobID = 0
	var jobURL string
	for _, value := range data {
		if value.Stage == "baseline" {
			if value.ID > maxJobID {
				maxJobID = value.ID
				jobURL = value.WebURL
			}
		}
	}
	return jobURL
}
