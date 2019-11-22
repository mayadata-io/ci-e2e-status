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
	repo := []int{6, 5, 1, 2}
	for _, repo := range repo {
		mayaUIPipelineData, err := pipelineDataa(repo, token)
		if err != nil {
			glog.Error(err)
			return
		}
		for i := range mayaUIPipelineData {
			// glog.Infoln(i, mayaUIPipelineData[i])
			commitDetails, err := getCommitData(mayaUIPipelineData[i].ID, token, repo)
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

			// pipelineJobss, err := pipelineJobsDataa(commitDetails.CommitPipeline.IDD, token, 6)
			// if err != nil {
			// 	glog.Error(err)
			// 	return
			// }
			// for i := range pipelineJobss {
			// 	glog.Infoln("PipelineDetails", pipelineJobss[i].Name)
			// }
		}
	}
}

func queryBuildDataa(datas *Builddashboard) error {
	pipelinerows, err := database.Db.Query(`SELECT * FROM commit_detail ORDER BY id DESC FETCH FIRST 30 ROWS ONLY;`)
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

// // jivaPipelineJobs will get pipeline jobs details from gitlab api
// func pipelineJobsDataa(id int, token string, project int) (BuildJobs, error) {
// 	// url := jobURLGenerator(id, project)
// 	url := BaseURL + "api/v4/projects/" + strconv.Itoa(project) + "/pipelines/" + strconv.Itoa(id) + "/jobs"
// 	glog.Infoln("url", url)
// 	req, err := http.NewRequest("GET", url, nil)
// 	if err != nil {
// 		return nil, err
// 	}
// 	req.Close = true
// 	req.Header.Set("Connection", "close")
// 	req.Header.Add("PRIVATE-TOKEN", token)
// 	res, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer res.Body.Close()
// 	body, err := ioutil.ReadAll(res.Body)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var obj BuildJobs
// 	json.Unmarshal(body, &obj)
// 	return obj, nil
// }

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
