package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/mayadata-io/ci-e2e-status/database"
)

// PipelineHandler return packet pipeline data to /packet path
func PipelineHandler(w http.ResponseWriter, r *http.Request) {
	// Allow cross origin request
	(w).Header().Set("Access-Control-Allow-Origin", "*")
	datas := dashboard{}
	err := QueryPipelineData(&datas)
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

// QueryPipelineData fetch the pipeline data as well as jobs data form Packet table of database
func QueryPipelineData(datas *dashboard) error {
	// Select all data from packetpipeline table of DB
	query := fmt.Sprintf("SELECT pipelineid,projectid,sha,ref,status,web_url,kibana_url,author_name,author_email,message FROM oep_build WHERE ref='staging' OR ref='master' ORDER BY pipelineid DESC;")
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
			&pipelinedata.ProjectID,
			&pipelinedata.Sha,
			&pipelinedata.Ref,
			&pipelinedata.Status,
			&pipelinedata.WebURL,
			&pipelinedata.LogURL,
			&pipelinedata.AuthorName,
			&pipelinedata.AuthorEmail,
			&pipelinedata.Message,
		)
		if err != nil {
			return err
		}
		// Query packetjobs data of respective pipeline using pipelineID from packetjobs table
		jobsquery := fmt.Sprintf("SELECT * FROM oep_build_jobs WHERE pipelineid = $1 ORDER BY id;")
		jobsrows, err := database.Db.Query(jobsquery, pipelinedata.PipelineID)
		if err != nil {
			return err
		}
		// Close DB connection after r/w operation
		defer jobsrows.Close()
		jobsdataarray := []Jobssummary{}
		// Iterate on each rows of table data for perform more operation related to pipelineJobsDatage api (#1511)ge api (#1511)
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
				&jobsdata.JobLogURL,
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

// oepData from gitlab api for oep and dump to database
func pipelineData(token string) {
	query := fmt.Sprintf("SELECT project,id,author_name,author_email,commit_message FROM commit_detail ORDER BY id DESC FETCH FIRST 20 ROWS ONLY;")
	oepPipelineID, err := database.Db.Query(query)
	if err != nil {
		glog.Error("OEP pipeline quering data Error:", err)
		return
	}
	for oepPipelineID.Next() {
		var logURL string
		pipelinedata := TriggredID{}
		err = oepPipelineID.Scan(
			&pipelinedata.ProjectID,
			&pipelinedata.ID,
			&pipelinedata.AuthorName,
			&pipelinedata.AuthorEmail,
			&pipelinedata.Message,
		)
		defer oepPipelineID.Close()
		oepPipelineData, err := oepPipeline(token, pipelinedata.ID, pipelinedata.ProjectID)
		if err != nil {
			glog.Error(err)
			return
		}
		pipelineJobsdata, err := oepPipelineJobs(oepPipelineData.ID, token, pipelinedata.ProjectID)
		if err != nil {
			glog.Error(err)
			return
		}
		if pipelinedata.ID != 0 && len(pipelineJobsdata) != 0 {
			jobStartedAt := pipelineJobsdata[0].StartedAt
			JobFinishedAt := pipelineJobsdata[len(pipelineJobsdata)-1].FinishedAt
			logURL = Kibanaloglink(oepPipelineData.Sha, oepPipelineData.ID, oepPipelineData.Status, jobStartedAt, JobFinishedAt)
		}
		sqlStatement := fmt.Sprintf(`INSERT INTO oep_build (projectid, pipelineid, sha, ref, status, web_url, kibana_url, author_name, author_email, message )
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 ON CONFLICT (pipelineid) DO UPDATE SET sha = $3, ref = $4, status = $5, web_url = $6, kibana_url = $7 RETURNING pipelineid;`)
		pipelineid := 0
		err = database.Db.QueryRow(sqlStatement,
			pipelinedata.ProjectID,
			oepPipelineData.ID,
			oepPipelineData.Sha,
			oepPipelineData.Ref,
			oepPipelineData.Status,
			oepPipelineData.WebURL,
			logURL,
			pipelinedata.AuthorName,
			pipelinedata.AuthorEmail,
			pipelinedata.Message,
		).Scan(&pipelineid)
		if err != nil {
			glog.Error(err)
		}
		glog.Infoln("New record ID for OEP Pipeline:", oepPipelineData.ID)
		if pipelinedata.ID != 0 {
			for j := range pipelineJobsdata {
				// var jobLogURL string
				var TriggeredGCP, TriggeredKonvoy, TriggeredRancher string

				sqlStatement := fmt.Sprintf("INSERT INTO oep_build_jobs (pipelineid, id, status, stage, name, ref, created_at, started_at, finished_at, job_log_url) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)" +
					"ON CONFLICT (id) DO UPDATE SET status = $3, stage = $4, name = $5, ref = $6, created_at = $7, started_at = $8, finished_at = $9, job_log_url = $10 RETURNING id;")
				id := 0
				if len(pipelineJobsdata) != 0 {
					// jobLogURL = Kibanaloglink(oepPipelineData.Sha, oepPipelineData.ID, oepPipelineData.Status, pipelineJobsdata[j].StartedAt, pipelineJobsdata[j].FinishedAt)
				}
				if pipelineJobsdata[j].Stage == "BASELINE" || pipelineJobsdata[j].Stage == "TRIGGER-E2E" {
					glog.Infoln("baseline stage job : = -  [ ", len(pipelineJobsdata), " ] ", pipelineJobsdata[j].Name)
					if len(pipelineJobsdata) == 4 {
						if pipelineJobsdata[j].Name == "baseline-image" || pipelineJobsdata[j].Name == "gcp-e2e" {
							TriggeredGCP, err = getTriggerPipelineFromBuild(pipelineJobsdata[j].ID, token, pipelinedata.ProjectID)
							goPipeOep(token, TriggeredGCP, pipelinedata.AuthorName, pipelinedata.AuthorEmail, pipelinedata.Message, oepPipelineData.ID, oepPipelineData.Sha, "oep", 5)
							if err != nil {
								glog.Error(err)
							}
						} else if pipelineJobsdata[j].Name == "konvoy-e2e" {
							TriggeredKonvoy, err = getTriggerPipelineFromBuild(pipelineJobsdata[j].ID, token, pipelinedata.ProjectID)
							goPipeOep(token, TriggeredKonvoy, pipelinedata.AuthorName, pipelinedata.AuthorEmail, pipelinedata.Message, oepPipelineData.ID, oepPipelineData.Sha, "konvoy", 37)
							if err != nil {
								glog.Error(err)
							}
							glog.Infoln("Konvoy-pipeline-trigger", TriggeredKonvoy)
						} else if pipelineJobsdata[j].Name == "rancher-e2e" {
							TriggeredRancher, err = getTriggerPipelineFromBuild(pipelineJobsdata[j].ID, token, pipelinedata.ProjectID)
							goPipeOep(token, TriggeredRancher, pipelinedata.AuthorName, pipelinedata.AuthorEmail, pipelinedata.Message, oepPipelineData.ID, oepPipelineData.Sha, "rancher", 36)
							if err != nil {
								glog.Error(err)
							}
						} else {
							goPipeOep(token, "dummy", pipelinedata.AuthorName, pipelinedata.AuthorEmail, pipelinedata.Message, oepPipelineData.ID, oepPipelineData.Sha, "dummy", 0)
						}
					} else {
						if pipelineJobsdata[j].Name == "baseline-image" {
							TriggeredGCP, err = getTriggerPipelineFromBuild(pipelineJobsdata[j].ID, token, pipelinedata.ProjectID)
							goPipeOep(token, TriggeredGCP, pipelinedata.AuthorName, pipelinedata.AuthorEmail, pipelinedata.Message, oepPipelineData.ID, oepPipelineData.Sha, "oep", 5)
							if err != nil {
								glog.Error(err)
							}
						}

						goPipeOep(token, "dummy", pipelinedata.AuthorName, pipelinedata.AuthorEmail, pipelinedata.Message, oepPipelineData.ID, oepPipelineData.Sha, "konvoy", 0)
						goPipeOep(token, "dummy", pipelinedata.AuthorName, pipelinedata.AuthorEmail, pipelinedata.Message, oepPipelineData.ID, oepPipelineData.Sha, "rancher", 0)

					}
				}
				err = database.Db.QueryRow(sqlStatement,
					oepPipelineData.ID,
					pipelineJobsdata[j].ID,
					pipelineJobsdata[j].Status,
					pipelineJobsdata[j].Stage,
					pipelineJobsdata[j].Name,
					pipelineJobsdata[j].Ref,
					pipelineJobsdata[j].CreatedAt,
					pipelineJobsdata[j].StartedAt,
					pipelineJobsdata[j].FinishedAt,
					TriggeredGCP,
				).Scan(&id)
				if err != nil {
					glog.Error(err)
				}
				glog.Infoln("New record ID for OEP Jobs:", id)
			}
		}
	}
}

func getTriggerPipelineFromBuild(jobid int, token string, proID int) (string, error) {
	// curl --location --header "PRIVATE-TOKEN: <your_access_token>" "https://gitlab.example.com/api/v4/projects/1/jobs/8/trace"
	url := BaseURL + "api/v4/projects/" + strconv.Itoa(proID) + "/jobs/" + strconv.Itoa(jobid) + "/trace"
	glog.Infoln("trigger job ", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "0", err
	}
	req.Close = true
	req.Header.Set("Connection", "close")
	req.Header.Add("PRIVATE-TOKEN", token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "0", err
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	data := string(body)
	if data == "" {
		return "0", nil
	}
	// grep := exec.Command("grep", "-oP", "(?<=master)[^ ]*")
	// ps := exec.Command("echo", data)

	// // Get ps's stdout and attach it to grep's stdin.
	// pipe, _ := ps.StdoutPipe()
	// defer pipe.Close()
	// grep.Stdin = pipe
	// ps.Start()

	// // Run and get the output of grep.
	// value, _ := grep.Output()

	re := regexp.MustCompile("master[^ ]*")
	value := re.FindString(data)
	if string(value) == "" {
		return "0", nil
	}
	result := strings.Split(string(value), "\"")
	result = strings.Split(string(result[8]), "/")
	if result[6] == "" {
		return "0", nil
	}
	return result[6], nil
}

// oepPipeline will get data from gitlab api and store to DB
func oepPipeline(token string, pipelineID int, projectID int) (*PlatformPipeline, error) {
	dummyJSON := []byte(`{"id":0,"sha":"00000000000000000000","ref":"none","status":"none","web_url":"none"}`)
	if pipelineID == 0 {
		glog.Infoln("pipelineURL : DummyJSON")
		var obj PlatformPipeline
		json.Unmarshal(dummyJSON, &obj)
		return &obj, nil
	}
	// Store packet pipeline data form gitlab api to packetObj
	url := BaseURL + "api/v4/projects/" + strconv.Itoa(projectID) + "/pipelines/" + strconv.Itoa(pipelineID)
	glog.Infoln("pipelineURL : ", url)
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
	var obj PlatformPipeline
	json.Unmarshal(body, &obj)
	if obj.ID == 0 {
		return nil, fmt.Errorf("Pipeline data not found")
	}
	return &obj, nil
}

// oepPipelineJobs will get pipeline jobs details from gitlab jobs api
func oepPipelineJobs(pipelineID int, token string, projectID int) (Jobs, error) {
	// Generate pipeline jobs api url using BaseURL, pipelineID and oepID
	if pipelineID == 0 {
		return nil, nil
	}
	url := BaseURL + "api/v4/projects/" + strconv.Itoa(projectID) + "/pipelines/" + strconv.Itoa(pipelineID) + "/jobs?per_page=75"
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
