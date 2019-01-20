package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/openebs/ci-e2e-dashboard-go-backend/database"
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

// BuildData from gitlab api for packet and dump to database
func BuildData(token string) {
	jivaPipelineData, err := jivaPipeline(token)
	if err != nil {
		glog.Error(err)
		return
	}
	// mayaObj, err := mayaPipeline(token)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	for i := range jivaPipelineData {
		jivaJobsData, err := jivaPipelineJobs(jivaPipelineData[i].ID, token)
		if err != nil {
			glog.Error(err)
			return
		}
		// Get GKE, Triggred pipeline ID for jiva build
		gkeTriggerID, err := getTriggerPipelineid(jivaJobsData[1].WebURL, "e2e-gke")
		if err != nil {
			glog.Error(err)
		}
		// Get EKS, Triggred pipeline ID for jiva build
		eksTriggerID, err := getTriggerPipelineid(jivaJobsData[1].WebURL, "e2e-eks")
		if err != nil {
			glog.Error(err)
		}
		// Add jiva pipelines data to Database
		sqlStatement := `
			INSERT INTO buildpipeline (id, sha, ref, status, web_url, gke_trigger_pid, eks_trigger_pid)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id) DO UPDATE
			SET status = $4, gke_trigger_pid = $6, eks_trigger_pid = $7
			RETURNING id`
		id := 0
		err = database.Db.QueryRow(sqlStatement,
			jivaPipelineData[i].ID,
			jivaPipelineData[i].Sha,
			jivaPipelineData[i].Ref,
			jivaPipelineData[i].Status,
			jivaPipelineData[i].WebURL,
			gkeTriggerID,
			eksTriggerID,
		).Scan(&id)
		if err != nil {
			glog.Error(err)
		}
		glog.Infoln("New record ID for jiva Pipeline:", id)

		// Add jiva jobs data to Database
		for j := range jivaJobsData {
			sqlStatement := `
				INSERT INTO buildjobs (pipelineid, id, status, stage, name, ref, created_at, started_at, finished_at, message, author_name)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
				ON CONFLICT (id) DO UPDATE
				SET status = $3, stage = $4, name = $5, ref = $6, created_at = $7, started_at = $8, finished_at = $9
				RETURNING id`
			id := 0
			err = database.Db.QueryRow(sqlStatement,
				jivaPipelineData[i].ID,
				jivaJobsData[j].ID,
				jivaJobsData[j].Status,
				jivaJobsData[j].Stage,
				jivaJobsData[j].Name,
				jivaJobsData[j].Ref,
				jivaJobsData[j].CreatedAt,
				jivaJobsData[j].StartedAt,
				jivaJobsData[j].FinishedAt,
				jivaJobsData[j].Commit.Message,
				jivaJobsData[j].Commit.AuthorName,
			).Scan(&id)
			if err != nil {
				glog.Error(err)
			}
			glog.Infoln("New record ID for jiva pipeline Jobs: ", id)
		}
	}
	// for i := range mayaObj {
	// 	mayaJobsData, err := mayaPipelineJobs(mayaObj[i].ID, token)
	// 	if err != nil {
	// 		glog.Error(err)
	// 		return
	// 	}
	// 	// Add Maya pipelines data to Database
	// 	sqlStatement := `
	// 		INSERT INTO buildpipeline (id, sha, ref, status, web_url)
	// 		VALUES ($1, $2, $3, $4, $5)
	// 		ON CONFLICT (id) DO UPDATE
	// 		SET status = $4
	// 		RETURNING id`
	// 	id := 0
	// 	err = database.Db.QueryRow(sqlStatement, mayaObj[i].ID, mayaObj[i].Sha, mayaObj[i].Ref, mayaObj[i].Status, mayaObj[i].WebURL).Scan(&id)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// 	fmt.Println("New record ID for jiva build Pipeline:", id)

	// 	// Add Maya jobs data to Database
	// 	for j := range mayaJobsData {
	// 		sqlStatement := `
	// 			INSERT INTO buildjobs (pipelineid, id, status, stage, name, ref, created_at, started_at, finished_at, message, author_name)
	// 			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	// 			ON CONFLICT (id) DO UPDATE
	// 			SET status = $3, stage = $4, name = $5, ref = $6, created_at = $7, started_at = $8, finished_at = $9
	// 			RETURNING id`
	// 		id := 0
	// 		err = database.Db.QueryRow(sqlStatement,
	// 			mayaObj[i].ID,
	// 			mayaJobsData[j].ID,
	// 			mayaJobsData[j].Status,
	// 			mayaJobsData[j].Stage,
	// 			mayaJobsData[j].Name,
	// 			mayaJobsData[j].Ref,
	// 			mayaJobsData[j].CreatedAt,
	// 			mayaJobsData[j].StartedAt,
	// 			mayaJobsData[j].FinishedAt,
	// 			mayaJobsData[j].Commit.Message,
	// 			mayaJobsData[j].Commit.AuthorName,
	// 		).Scan(&id)
	// 		if err != nil {
	// 			fmt.Println(err)
	// 		}
	// 		fmt.Println("New record ID for maya pipeline Jobs: ", id)
	// 	}
	// }
}

// queryBuildData fetches the builddashboard data from the db
func queryBuildData(datas *Builddashboard) error {
	pipelinerows, err := database.Db.Query(`SELECT * FROM buildpipeline ORDER BY id DESC`)
	if err != nil {
		return err
	}
	defer pipelinerows.Close()
	for pipelinerows.Next() {
		pipelinedata := BuildpipelineSummary{}
		err = pipelinerows.Scan(
			&pipelinedata.ID,
			&pipelinedata.Sha,
			&pipelinedata.Ref,
			&pipelinedata.Status,
			&pipelinedata.WebURL,
			&pipelinedata.GKETriggerPID,
			&pipelinedata.EKSTriggerPID,
		)
		if err != nil {
			return err
		}

		jobsquery := `SELECT * FROM buildjobs WHERE pipelineid = $1 ORDER BY id`
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

func getTriggerPipelineid(jobURL, platform string) (string, error) {
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

	grep := exec.Command("grep", "-oP", "(?<="+platform+"/pipelines/)[^ ]*")
	ps := exec.Command("echo", data)

	// Get ps's stdout and attach it to grep's stdin.
	pipe, _ := ps.StdoutPipe()
	defer pipe.Close()
	grep.Stdin = pipe
	ps.Start()

	// Run and get the output of grep.
	value, err := grep.Output()
	if err != nil {
		return "", err
	}
	result := strings.Split(string(value), "\"")
	if result[0] == "" {
		return "0", nil
	}
	return result[0], nil
}

// // jivaPipelineJobs will get pipeline jobs details from gitlab api
func jivaPipelineJobs(id int, token string) (BuildJobs, error) {
	url := BaseURL + "api/v4/projects/" + JIVAID + "/pipelines/" + strconv.Itoa(id) + "/jobs"
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
	body, _ := ioutil.ReadAll(res.Body)
	var obj BuildJobs
	json.Unmarshal(body, &obj)
	return obj, nil
}

// mayaPipelineJobs will get pipeline jobs details from gitlab api
// func mayaPipelineJobs(id int, token string) (BuildJobs, error) {
// 	url := BaseURL + "api/v4/projects/" + MAYAID + "/pipelines/" + strconv.Itoa(id) + "/jobs"
// 	req, err := http.NewRequest("GET", url, nil)
// 	if err != nil {
// 		fmt.Println(err)
// 		return nil, err
// 	}
// 	req.Close = true
// 	req.Header.Set("Connection", "close")
// 	req.Header.Add("PRIVATE-TOKEN", token)
// 	res, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		fmt.Println(err)
// 		return nil, err
// 	}
// 	defer res.Body.Close()
// 	body, _ := ioutil.ReadAll(res.Body)
// 	var obj BuildJobs
// 	json.Unmarshal(body, &obj)
// 	return obj, nil
// }

// jivaPipeline get jiva pipeline data from gitlab
func jivaPipeline(token string) (Pipeline, error) {
	jivaURL := BaseURL + "api/v4/projects/" + JIVAID + "/pipelines?ref=master"
	req, err := http.NewRequest("GET", jivaURL, nil)
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
	jivaData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var obj Pipeline
	json.Unmarshal(jivaData, &obj)
	return obj, nil
}

// mayaPipeline get maya pipeline data from gitlab
// func mayaPipeline(token string) (Pipeline, error) {
// 	mayaURL := BaseURL + "api/v4/projects/" + MAYAID + "/pipelines?ref=master"
// 	req, err := http.NewRequest("GET", mayaURL, nil)
// 	if err != nil {
// 		fmt.Println(err)
// 		return nil, err
// 	}
// 	req.Header.Add("PRIVATE-TOKEN", token)
// 	req.Close = true
// 	req.Header.Set("Connection", "close")
// 	res, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		fmt.Println(err)
// 		return nil, err
// 	}
// 	defer res.Body.Close()
// 	mayaData, _ := ioutil.ReadAll(res.Body)
// 	var obj Pipeline
// 	json.Unmarshal(mayaData, &obj)
// 	return obj, nil
// }
