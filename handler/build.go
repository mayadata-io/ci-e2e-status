package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/openebs/ci-e2e-dashboard-go-backend/database"
)

// Buildhandler return packet pipeline data to api
func Buildhandler(w http.ResponseWriter, r *http.Request) {
	(w).Header().Set("Access-Control-Allow-Origin", "*")
	datas := Builddashboard{}
	err := queryBuildData(&datas)
	if err != nil {
		http.Error(w, err.Error(), 500)
		fmt.Println(err)
		return
	}
	out, err := json.Marshal(datas)
	if err != nil {
		http.Error(w, err.Error(), 500)
		fmt.Println(err)
		return
	}
	w.Write(out)
}

// jivaPipelineJobs will get pipeline jobs details from gitlab api
func jivaPipelineJobs(id int, token string) BuildJobs {
	url := BaseURL + "api/v4/projects/" + PlatformID["jiva"] + "/pipelines/" + strconv.Itoa(id) + "/jobs?per_page=50"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	req.Close = true
	req.Header.Set("Connection", "close")
	req.Header.Add("PRIVATE-TOKEN", token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	var obj BuildJobs
	json.Unmarshal(body, &obj)
	return obj
}

// mayaPipelineJobs will get pipeline jobs details from gitlab api
func mayaPipelineJobs(id int, token string) BuildJobs {
	url := BaseURL + "api/v4/projects/" + PlatformID["maya"] + "/pipelines/" + strconv.Itoa(id) + "/jobs?per_page=50"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	req.Close = true
	req.Header.Set("Connection", "close")
	req.Header.Add("PRIVATE-TOKEN", token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	var obj BuildJobs
	json.Unmarshal(body, &obj)
	return obj
}

// BuildData from gitlab api for packet and dump to database
func BuildData(token string) {
	// Fetch jiva data from gitlab
	jivaURL := BaseURL + "api/v4/projects/" + PlatformID["jiva"] + "/pipelines?ref=master"
	req, err := http.NewRequest("GET", jivaURL, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Close = true
	req.Header.Set("Connection", "close")
	req.Header.Add("PRIVATE-TOKEN", token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	jivaData, _ := ioutil.ReadAll(res.Body)
	var jivaObj Pipeline
	json.Unmarshal(jivaData, &jivaObj)

	for i := range jivaObj {
		jivaJobsData := jivaPipelineJobs(jivaObj[i].ID, token)
		if err != nil {
			fmt.Println(err)
		}
		// Add jiva pipelines data to Database
		sqlStatement := `
			INSERT INTO buildpipeline (id, sha, ref, status, web_url)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO UPDATE
			SET status = $4
			RETURNING id`
		id := 0
		err = database.Db.QueryRow(sqlStatement, jivaObj[i].ID, jivaObj[i].Sha, jivaObj[i].Ref, jivaObj[i].Status, jivaObj[i].WebURL).Scan(&id)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("New record ID for jiva build Pipeline:", id)

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
				jivaObj[i].ID,
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
				fmt.Println(err)
			}
			fmt.Println("New record ID for jiva pipeline Jobs: ", id)
		}
	}

	// Fetch maya pipeline data from gitlab
	mayaURL := BaseURL + "api/v4/projects/" + PlatformID["maya"] + "/pipelines?ref=master"
	req, err = http.NewRequest("GET", mayaURL, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("PRIVATE-TOKEN", token)
	req.Close = true
	req.Header.Set("Connection", "close")
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	mayaData, _ := ioutil.ReadAll(res.Body)
	var mayaObj Pipeline
	json.Unmarshal(mayaData, &mayaObj)

	for i := range mayaObj {
		mayaJobsData := mayaPipelineJobs(mayaObj[i].ID, token)
		if err != nil {
			fmt.Println(err)
		}
		// Add Maya pipelines data to Database
		sqlStatement := `
			INSERT INTO buildpipeline (id, sha, ref, status, web_url)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO UPDATE
			SET status = $4
			RETURNING id`
		id := 0
		err = database.Db.QueryRow(sqlStatement, mayaObj[i].ID, mayaObj[i].Sha, mayaObj[i].Ref, mayaObj[i].Status, mayaObj[i].WebURL).Scan(&id)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("New record ID for jiva build Pipeline:", id)

		// Add Maya jobs data to Database
		for j := range mayaJobsData {
			sqlStatement := `
				INSERT INTO buildjobs (pipelineid, id, status, stage, name, ref, created_at, started_at, finished_at, message, author_name)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
				ON CONFLICT (id) DO UPDATE
				SET status = $3, stage = $4, name = $5, ref = $6, created_at = $7, started_at = $8, finished_at = $9
				RETURNING id`
			id := 0
			err = database.Db.QueryRow(sqlStatement,
				mayaObj[i].ID,
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
				fmt.Println(err)
			}
			fmt.Println("New record ID for maya pipeline Jobs: ", id)
		}
	}
}

// queryBuildData fetches the builddashboard data from the db
func queryBuildData(datas *Builddashboard) error {
	pipelinerows, err := database.Db.Query(`SELECT * FROM buildpipeline ORDER BY id DESC`)
	if err != nil {
		fmt.Println(err)
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
		)
		if err != nil {
			fmt.Println(err)
		}

		jobsquery := `SELECT * FROM buildjobs WHERE pipelineid = $1 ORDER BY id`
		jobsrows, err := database.Db.Query(jobsquery, pipelinedata.ID)
		if err != nil {
			fmt.Println(err)
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
				fmt.Println(err)
			}
			jobsdataarray = append(jobsdataarray, jobsdata)
			pipelinedata.Jobs = jobsdataarray
		}
		datas.Dashboard = append(datas.Dashboard, pipelinedata)
	}
	err = pipelinerows.Err()
	if err != nil {
		fmt.Println(err)
	}
	return nil
}
