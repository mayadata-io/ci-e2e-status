package handler

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/mayadata-io/ci-e2e-status/database"
)

type ImageTagCheck struct {
	TableName string
	JobID     int
}

// VerifyImageTagExists functions checks if the releaseTag exists or not
func VerifyImageTagExists(t ImageTagCheck) string {
	var releaseTag string
	query := fmt.Sprintf("SELECT release_tag FROM %s WHERE id=%d", t.TableName, t.JobID)
	row := database.Db.QueryRow(query)
	switch err := row.Scan(&releaseTag); err {
	case sql.ErrNoRows:
		fmt.Printf("\nNo imageTag rows were returned for %s table of %d jobID \n", t.TableName, t.JobID)
		return "NA"
	case nil:
		return releaseTag
	default:
		fmt.Println(err)
		return "NA"
	}

}

// VerifyColumnDataExixts functions checks if the column exists or not
func VerifyColumnDataExixts(t ImageTagCheck) bool {
	var columnData string
	query := fmt.Sprintf("SELECT k8s_version FROM %s WHERE id=%d", t.TableName, t.JobID)
	row := database.Db.QueryRow(query)
	err := row.Scan(&columnData)
	switch err {
	case sql.ErrNoRows:
		fmt.Printf("\nNo rows were returned for %s table of %d jobID \n", t.TableName, t.JobID)
		return false
	case nil:
		fmt.Printf("\n\n\n\t\t\t K8s_version Column:%s,<--", columnData)
		if columnData == "" {
			return false
		}
		return true
	default:
		fmt.Println(err)
		return false
	}
}

func GrepFromRaw(jobsData Jobs, token, project, branch, jobName string) (string, error) {
	var jobURL string
	// glog.Infoln(fmt.Sprintf("\n platform : %s \n branch : %s \n jobName : %s \n", project, branch, jobName))
	for _, job := range jobsData {
		if job.Name == jobName {
			if job.Status == "success" || job.Status == "failed" {
				jobURL = job.WebURL + "/raw"
			}
		}
	}
	if jobURL == "" {
		return "NA", nil
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
	re := regexp.MustCompile("k8s_version=[^ ]*")
	value := re.FindString(data)
	result := strings.Split(string(value), "=")
	if result != nil && len(result) > 2 {
		if result[2] == "" {
			return "NA", nil
		}
		releaseVersion := strings.Split(result[2], "\n")
		return releaseVersion[0], nil
	}
	return "NA", nil
}
