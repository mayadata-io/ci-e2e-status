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

// PacketHandlerV11 return packet pipeline data to /packet path
func PacketHandlerV11(w http.ResponseWriter, r *http.Request) {
	// Allow cross origin request
	(w).Header().Set("Access-Control-Allow-Origin", "*")
	datas := dashboard{}
	err := QueryPacketData(&datas, "packet_pipeline_v11", "packet_jobs_v11")
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

// PacketHandlerV12 return packet pipeline data to /packet path
// TODO
func PacketHandlerV12(w http.ResponseWriter, r *http.Request) {
	// Allow cross origin request
	(w).Header().Set("Access-Control-Allow-Origin", "*")
	datas := dashboard{}
	err := QueryPacketData(&datas, "packet_pipeline_v12", "packet_jobs_v12")
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

// PacketHandlerV13 return packet pipeline data to /packet path
// TODO
func PacketHandlerV13(w http.ResponseWriter, r *http.Request) {
	// Allow cross origin request
	(w).Header().Set("Access-Control-Allow-Origin", "*")
	datas := dashboard{}
	err := QueryPacketData(&datas, "packet_pipeline_v13", "packet_jobs_v13")
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

// PacketData from gitlab api for Packet and dump to database
func PacketData(token, triggredIDColumnName, pipelineTableName, jobTableName string) {
	query := fmt.Sprintf("SELECT id,%s FROM build_pipeline ORDER BY id DESC FETCH FIRST 20 ROWS ONLY;", triggredIDColumnName)
	packetPipelineID, err := database.Db.Query(query)
	if err != nil {
		glog.Error("PACKET pipeline quering data Error:", err)
		return
	}
	for packetPipelineID.Next() {
		var logURL string
		pipelinedata := TriggredID{}
		err = packetPipelineID.Scan(
			&pipelinedata.BuildPID,
			&pipelinedata.ID,
		)
		defer packetPipelineID.Close()
		packetPipelineData, err := packetPipeline(token, pipelinedata.ID)
		if err != nil {
			glog.Error(err)
			return
		}
		pipelineJobsdata, err := packetPipelineJobs(packetPipelineData.ID, token)
		if err != nil {
			glog.Error(err)
			return
		}
		if pipelinedata.ID != 0 && len(pipelineJobsdata) != 0 {
			jobStartedAt := pipelineJobsdata[0].StartedAt
			JobFinishedAt := pipelineJobsdata[len(pipelineJobsdata)-1].FinishedAt
			logURL = Kibanaloglink(packetPipelineData.Sha, packetPipelineData.ID, packetPipelineData.Status, jobStartedAt, JobFinishedAt)
		}
		sqlStatement := fmt.Sprintf("INSERT INTO %s (build_pipelineid, id, sha, ref, status, web_url, kibana_url) VALUES ($1, $2, $3, $4, $5, $6, $7)"+
			"ON CONFLICT (build_pipelineid) DO UPDATE SET id = $2, sha = $3, ref = $4, status = $5, web_url = $6, kibana_url = $7 RETURNING id;", pipelineTableName)
		id := 0
		err = database.Db.QueryRow(sqlStatement,
			pipelinedata.BuildPID,
			packetPipelineData.ID,
			packetPipelineData.Sha,
			packetPipelineData.Ref,
			packetPipelineData.Status,
			packetPipelineData.WebURL,
			logURL,
		).Scan(&id)
		if err != nil {
			glog.Error(err)
		}
		glog.Infoln("New record ID for PACKET Pipeline:", id)
		if pipelinedata.ID != 0 {
			for j := range pipelineJobsdata {
				var jobLogURL string
				sqlStatement := fmt.Sprintf("INSERT INTO %s (pipelineid, id, status, stage, name, ref, created_at, started_at, finished_at, job_log_url) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)"+
					"ON CONFLICT (id) DO UPDATE SET status = $3, stage = $4, name = $5, ref = $6, created_at = $7, started_at = $8, finished_at = $9, job_log_url = $10 RETURNING id;", jobTableName)
				id := 0
				if len(pipelineJobsdata) != 0 {
					jobLogURL = Kibanaloglink(packetPipelineData.Sha, packetPipelineData.ID, packetPipelineData.Status, pipelineJobsdata[j].StartedAt, pipelineJobsdata[j].FinishedAt)
				}
				err = database.Db.QueryRow(sqlStatement,
					packetPipelineData.ID,
					pipelineJobsdata[j].ID,
					pipelineJobsdata[j].Status,
					pipelineJobsdata[j].Stage,
					pipelineJobsdata[j].Name,
					pipelineJobsdata[j].Ref,
					pipelineJobsdata[j].CreatedAt,
					pipelineJobsdata[j].StartedAt,
					pipelineJobsdata[j].FinishedAt,
					jobLogURL,
				).Scan(&id)
				if err != nil {
					glog.Error(err)
				}
				glog.Infoln("New record ID for PACKET Jobs:", id)
			}
		}
	}
}

// packetPipeline will get data from gitlab api and store to DB
func packetPipeline(token string, pipelineID int) (*PlatformPipeline, error) {
	dummyJSON := []byte(`{"id":0,"sha":"00000000000000000000","ref":"none","status":"none","web_url":"none"}`)
	if pipelineID == 0 {
		var obj PlatformPipeline
		json.Unmarshal(dummyJSON, &obj)
		return &obj, nil
	}
	// Store packet pipeline data form gitlab api to packetObj
	url := BaseURL + "api/v4/projects/" + PACKETID + "/pipelines/" + strconv.Itoa(pipelineID)
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

// packetPipelineJobs will get pipeline jobs details from gitlab jobs api
func packetPipelineJobs(pipelineID int, token string) (Jobs, error) {
	// Generate pipeline jobs api url using BaseURL, pipelineID and PACKETID
	if pipelineID == 0 {
		return nil, nil
	}
	url := BaseURL + "api/v4/projects/" + PACKETID + "/pipelines/" + strconv.Itoa(pipelineID) + "/jobs?per_page=50"
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

// QueryPacketData fetch the pipeline data as well as jobs data form Packet table of database
func QueryPacketData(datas *dashboard, pipelineTableName string, jobTableName string) error {
	// Select all data from packetpipeline table of DB
	query := fmt.Sprintf("SELECT id,sha,ref,status,web_url,kibana_url FROM %s ORDER BY build_pipelineid DESC;", pipelineTableName)
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
			&pipelinedata.ID,
			&pipelinedata.Sha,
			&pipelinedata.Ref,
			&pipelinedata.Status,
			&pipelinedata.WebURL,
			&pipelinedata.LogURL,
		)
		if err != nil {
			return err
		}
		// Query packetjobs data of respective pipeline using pipelineID from packetjobs table
		jobsquery := fmt.Sprintf("SELECT * FROM %s WHERE pipelineid = $1 ORDER BY id;", jobTableName)
		jobsrows, err := database.Db.Query(jobsquery, pipelinedata.ID)
		if err != nil {
			return err
		}
		// Close DB connection after r/w operation
		defer jobsrows.Close()
		jobsdataarray := []Jobssummary{}
		// Iterate on each rows of table data for perform more operation related to pipelineJobsData
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
