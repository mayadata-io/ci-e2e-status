package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/mayadata-io/ci-e2e-status/database"
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
func BuildData(token string) {
	jivaPipelineData, err := pipelineData("jiva", token)
	if err != nil {
		glog.Error(err)
		return
	}
	project = "jiva"
	for i := range jivaPipelineData {
		jivaJobsData, err := pipelineJobsData(jivaPipelineData[i].ID, token, "jiva")
		if err != nil {
			glog.Error(err)
			return
		}
		// Getting webURL link for getting triggredID
		baselineJobsWebURL := getBaselineJobWebURL(jivaJobsData)
		// Get GKE, Triggred pipeline ID for jiva build
		packetV11PID, err := getTriggerPipelineid(baselineJobsWebURL, "k8s-1-11")
		if err != nil {
			glog.Error(err)
		}
		// Get EKS, Triggred pipeline ID for jiva build
		packetV12PID, err := getTriggerPipelineid(baselineJobsWebURL, "k8s-1-12")
		if err != nil {
			glog.Error(err)
		}
		// Get AKS, Triggred pipeline ID for jiva build
		packetV13PID, err := getTriggerPipelineid(baselineJobsWebURL, "k8s-1-13")
		if err != nil {
			glog.Error(err)
		}
		// Add jiva pipelines data to Database
		sqlStatement := `
			INSERT INTO build_pipeline (project, id, sha, ref, status, web_url, packet_v11_pid, packet_v12_pid, packet_v13_pid)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (id) DO UPDATE
			SET status = $5, packet_v11_pid = $7, packet_v12_pid = $8, packet_v13_pid = $9
			RETURNING id`
		id := 0
		err = database.Db.QueryRow(sqlStatement,
			project,
			jivaPipelineData[i].ID,
			jivaPipelineData[i].Sha,
			jivaPipelineData[i].Ref,
			jivaPipelineData[i].Status,
			jivaPipelineData[i].WebURL,
			packetV11PID,
			packetV12PID,
			packetV13PID,
		).Scan(&id)
		if err != nil {
			glog.Error(err)
		}
		glog.Infoln("New record ID for jiva Pipeline:", id)

		// Add jiva jobs data to Database
		for j := range jivaJobsData {
			sqlStatement := `
				INSERT INTO build_jobs (pipelineid, id, status, stage, name, ref, created_at, started_at, finished_at, message, author_name)
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

	project = "maya"
	mayaPipelineData, err := pipelineData("maya", token)
	if err != nil {
		glog.Error(err)
		return
	}
	for i := range mayaPipelineData {
		mayaJobsData, err := pipelineJobsData(mayaPipelineData[i].ID, token, "maya")
		if err != nil {
			glog.Error(err)
			return
		}
		// Getting webURL link for getting triggredID
		baselineJobsWebURL := getBaselineJobWebURL(mayaJobsData)
		// Get GKE, Triggred pipeline ID for maya build
		packetV11PID, err := getTriggerPipelineid(baselineJobsWebURL, "k8s-1-11")
		if err != nil {
			glog.Error(err)
		}
		// Get EKS, Triggred pipeline ID for maya build
		packetV12PID, err := getTriggerPipelineid(baselineJobsWebURL, "k8s-1-12")
		if err != nil {
			glog.Error(err)
		}
		// Get AKS, Triggred pipeline ID for maya build
		packetV13PID, err := getTriggerPipelineid(baselineJobsWebURL, "k8s-1-13")
		if err != nil {
			glog.Error(err)
		}
		// Add maya pipelines data to Database
		sqlStatement := `
			INSERT INTO build_pipeline (project, id, sha, ref, status, web_url, packet_v11_pid, packet_v12_pid, packet_v13_pid)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (id) DO UPDATE
			SET status = $5, packet_v11_pid = $7, packet_v12_pid = $8, packet_v13_pid = $9
			RETURNING id`
		id := 0
		err = database.Db.QueryRow(sqlStatement,
			project,
			mayaPipelineData[i].ID,
			mayaPipelineData[i].Sha,
			mayaPipelineData[i].Ref,
			mayaPipelineData[i].Status,
			mayaPipelineData[i].WebURL,
			packetV11PID,
			packetV12PID,
			packetV13PID,
		).Scan(&id)
		if err != nil {
			glog.Error(err)
		}
		glog.Infoln("New record ID for maya Pipeline:", id)

		// Add maya jobs data to Database
		for j := range mayaJobsData {
			sqlStatement := `
				INSERT INTO build_jobs (pipelineid, id, status, stage, name, ref, created_at, started_at, finished_at, message, author_name)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
				ON CONFLICT (id) DO UPDATE
				SET status = $3, stage = $4, name = $5, ref = $6, created_at = $7, started_at = $8, finished_at = $9
				RETURNING id`
			id := 0
			err = database.Db.QueryRow(sqlStatement,
				mayaPipelineData[i].ID,
				mayaJobsData[j].ID,
				mayaJobsData[j].Status,
				mayaJobsData[j].Stage,
				mayaJobsData[j].Name,
				mayaJobsData[j].Ref,
				mayaJobsData[j].CreatedAt,
				mayaJobsData[j].StartedAt,
				mayaJobsData[j].FinishedAt,
				mayaJobsData[j].Commit.Message,
				mayaJobsData[j].Commit.AuthorName,
			).Scan(&id)
			if err != nil {
				glog.Error(err)
			}
			glog.Infoln("New record ID for maya pipeline Jobs: ", id)
		}
	}

	project = "zfs"
	zfsPipelineData, err := pipelineData("zfs", token)
	if err != nil {
		glog.Error(err)
		return
	}
	for i := range zfsPipelineData {
		zfsJobsData, err := pipelineJobsData(zfsPipelineData[i].ID, token, "zfs")
		if err != nil {
			glog.Error(err)
			return
		}
		// Getting webURL link for getting triggredID
		baselineJobsWebURL := getBaselineJobWebURL(zfsJobsData)
		// Get GKE, Triggred pipeline ID for zfs build
		packetV11PID, err := getTriggerPipelineid(baselineJobsWebURL, "k8s-1-11")
		if err != nil {
			glog.Error(err)
		}
		// Get EKS, Triggred pipeline ID for zfs build
		packetV12PID, err := getTriggerPipelineid(baselineJobsWebURL, "k8s-1-12")
		if err != nil {
			glog.Error(err)
		}
		// Get AKS, Triggred pipeline ID for zfs build
		packetV13PID, err := getTriggerPipelineid(baselineJobsWebURL, "k8s-1-13")
		if err != nil {
			glog.Error(err)
		}
		// Add zfs pipelines data to Database
		sqlStatement := `
			INSERT INTO build_pipeline (project, id, sha, ref, status, web_url, packet_v11_pid, packet_v12_pid, packet_v13_pid)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (id) DO UPDATE
			SET status = $5, packet_v11_pid = $7, packet_v12_pid = $8, packet_v13_pid = $9
			RETURNING id`
		id := 0
		err = database.Db.QueryRow(sqlStatement,
			project,
			zfsPipelineData[i].ID,
			zfsPipelineData[i].Sha,
			zfsPipelineData[i].Ref,
			zfsPipelineData[i].Status,
			zfsPipelineData[i].WebURL,
			packetV11PID,
			packetV12PID,
			packetV13PID,
		).Scan(&id)
		if err != nil {
			glog.Error(err)
		}
		glog.Infoln("New record ID for zfs Pipeline:", id)

		// Add zfs jobs data to Database
		for j := range zfsJobsData {
			sqlStatement := `
				INSERT INTO build_jobs (pipelineid, id, status, stage, name, ref, created_at, started_at, finished_at, message, author_name)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
				ON CONFLICT (id) DO UPDATE
				SET status = $3, stage = $4, name = $5, ref = $6, created_at = $7, started_at = $8, finished_at = $9
				RETURNING id`
			id := 0
			err = database.Db.QueryRow(sqlStatement,
				zfsPipelineData[i].ID,
				zfsJobsData[j].ID,
				zfsJobsData[j].Status,
				zfsJobsData[j].Stage,
				zfsJobsData[j].Name,
				zfsJobsData[j].Ref,
				zfsJobsData[j].CreatedAt,
				zfsJobsData[j].StartedAt,
				zfsJobsData[j].FinishedAt,
				zfsJobsData[j].Commit.Message,
				zfsJobsData[j].Commit.AuthorName,
			).Scan(&id)
			if err != nil {
				glog.Error(err)
			}
			glog.Infoln("New record ID for zfs pipeline Jobs: ", id)
		}
	}

	project = "istgt"
	istgtPipelineData, err := pipelineData("istgt", token)
	if err != nil {
		glog.Error(err)
		return
	}
	for i := range istgtPipelineData {
		istgtJobsData, err := pipelineJobsData(istgtPipelineData[i].ID, token, "istgt")
		if err != nil {
			glog.Error(err)
			return
		}
		// Getting webURL link for getting triggredID
		baselineJobsWebURL := getBaselineJobWebURL(istgtJobsData)
		// Get GKE, Triggred pipeline ID for istgt build
		packetV11PID, err := getTriggerPipelineid(baselineJobsWebURL, "k8s-1-11")
		if err != nil {
			glog.Error(err)
		}
		// Get EKS, Triggred pipeline ID for istgt build
		packetV12PID, err := getTriggerPipelineid(baselineJobsWebURL, "k8s-1-12")
		if err != nil {
			glog.Error(err)
		}
		// Get AKS, Triggred pipeline ID for istgt build
		packetV13PID, err := getTriggerPipelineid(baselineJobsWebURL, "k8s-1-13")
		if err != nil {
			glog.Error(err)
		}
		// Add istgt pipelines data to Database
		sqlStatement := `
			INSERT INTO build_pipeline (project, id, sha, ref, status, web_url, packet_v11_pid, packet_v12_pid, packet_v13_pid)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (id) DO UPDATE
			SET status = $5, packet_v11_pid = $7, packet_v12_pid = $8, packet_v13_pid = $9
			RETURNING id`
		id := 0
		err = database.Db.QueryRow(sqlStatement,
			project,
			istgtPipelineData[i].ID,
			istgtPipelineData[i].Sha,
			istgtPipelineData[i].Ref,
			istgtPipelineData[i].Status,
			istgtPipelineData[i].WebURL,
			packetV11PID,
			packetV12PID,
			packetV13PID,
		).Scan(&id)
		if err != nil {
			glog.Error(err)
		}
		glog.Infoln("New record ID for istgt Pipeline:", id)

		// Add istgt jobs data to Database
		for j := range istgtJobsData {
			sqlStatement := `
				INSERT INTO build_jobs (pipelineid, id, status, stage, name, ref, created_at, started_at, finished_at, message, author_name)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
				ON CONFLICT (id) DO UPDATE
				SET status = $3, stage = $4, name = $5, ref = $6, created_at = $7, started_at = $8, finished_at = $9
				RETURNING id`
			id := 0
			err = database.Db.QueryRow(sqlStatement,
				istgtPipelineData[i].ID,
				istgtJobsData[j].ID,
				istgtJobsData[j].Status,
				istgtJobsData[j].Stage,
				istgtJobsData[j].Name,
				istgtJobsData[j].Ref,
				istgtJobsData[j].CreatedAt,
				istgtJobsData[j].StartedAt,
				istgtJobsData[j].FinishedAt,
				istgtJobsData[j].Commit.Message,
				istgtJobsData[j].Commit.AuthorName,
			).Scan(&id)
			if err != nil {
				glog.Error(err)
			}
			glog.Infoln("New record ID for istgt pipeline Jobs:", id)
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
			&pipelinedata.PacketV11PID,
			&pipelinedata.PacketV12PID,
			&pipelinedata.PacketV13PID,
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

func getTriggerPipelineid(jobURL, k8sVersion string) (string, error) {
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
	grep := exec.Command("grep", "-oP", "(?<="+k8sVersion+")[^ ]*")
	ps := exec.Command("echo", data)

	// Get ps's stdout and attach it to grep's stdin.
	pipe, _ := ps.StdoutPipe()
	defer pipe.Close()
	grep.Stdin = pipe
	ps.Start()

	// Run and get the output of grep.
	value, _ := grep.Output()
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

// // jivaPipelineJobs will get pipeline jobs details from gitlab api
func pipelineJobsData(id int, token string, project string) (BuildJobs, error) {
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
func pipelineData(project, token string) (Pipeline, error) {
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
	generatedURL := BaseURL + "api/v4/projects/" + projectID + "/pipelines/" + strconv.Itoa(id) + "/jobs"
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
