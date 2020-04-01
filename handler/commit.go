package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/golang/glog"
	"github.com/mayadata-io/ci-e2e-status/database"
)

// CommitHandler return eks pipeline data to /commit path
func CommitHandler(w http.ResponseWriter, r *http.Request) {
	// Allow cross origin request
	(w).Header().Set("Access-Control-Allow-Origin", "*")
	datas := Builddashboard{}
	err := queryBuildDataa(&datas)
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

func commitData(token string) {
	repo := []int{1, 6, 14, 15, 16, 17, 18, 38}

	for _, repo := range repo {
		pipelineData, err := pipelineDataa(repo, token)
		if err != nil {
			glog.Error(err)
			return
		}
		for i := range pipelineData {
			commitDetails, err := getCommitData(pipelineData[i].ID, token, repo)
			if err != nil {
				glog.Error(err)
				return
			}
			sqlStatement := `
			INSERT INTO commit_detail (project, id, sha, ref, status, web_url, committeddate, author_name, author_email, comitter_name, commit_title, commit_message)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			ON CONFLICT (id) DO UPDATE
			SET status = $5	RETURNING id`
			id := 0
			err = database.Db.QueryRow(sqlStatement,
				commitDetails.ProjectID,
				commitDetails.CommitPipeline.ID,
				commitDetails.Sha,
				commitDetails.CommitPipeline.Ref,
				commitDetails.CommitPipeline.Status,
				commitDetails.CommitPipeline.WebURL,
				commitDetails.CommittedDate,
				commitDetails.AuthorName,
				commitDetails.AuthorEmail,
				commitDetails.CommitterName,
				commitDetails.CommitterEmail,
				commitDetails.CommitMessage,
			).Scan(&id)
			if err != nil {
				glog.Error(err)
			}
			glog.Infoln("New Commit for " + strconv.Itoa(repo) + " Project : " + commitDetails.Sha)
		}
	}
}

func queryBuildDataa(datas *Builddashboard) error {
	pipelinerows, err := database.Db.Query(`SELECT * FROM commit_detail WHERE ref='staging' OR ref='master' ORDER BY id DESC FETCH FIRST 50 ROWS ONLY;`)
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
			&pipelinedata.Committeddate,
			&pipelinedata.AuthorName,
			&pipelinedata.AuthorEmail,
			&pipelinedata.ComitterName,
			&pipelinedata.CommitTitle,
			&pipelinedata.CommitMessage,
		)
		if err != nil {
			return err
		}
		datas.Dashboard = append(datas.Dashboard, pipelinedata)
	}
	err = pipelinerows.Err()
	if err != nil {
		return err
	}
	return nil
}

// // pipelineData will fetch the data from gitlab API
func pipelineDataa(project int, token string) (Commit, error) {
	URL := "https://gitlab.mayadata.io/api/v4/projects/" + strconv.Itoa(project) + "/repository/commits"
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

	var obj Commit
	json.Unmarshal(data, &obj)
	return obj, nil
}
func getCommitData(commitID, token string, project int) (CommitData, error) {
	var commit CommitData
	URL := "https://gitlab.mayadata.io/api/v4/projects/" + strconv.Itoa(project) + "/repository/commits/" + commitID
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return commit, err
	}
	req.Close = true
	req.Header.Set("Connection", "close")
	req.Header.Add("PRIVATE-TOKEN", token)
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return commit, err
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return commit, err
	}

	var obj CommitData
	json.Unmarshal(data, &obj)
	return obj, nil
}
