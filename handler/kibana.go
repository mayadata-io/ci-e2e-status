package handler

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/mayadata-io/ci-e2e-status/database"
)

// Kibanaloglink generate log link for platform pipelines
func Kibanaloglink(sha string, pipelineID int, status string, startedAt string, finishedAt string) string {
	if status == "success" || status == "failed" {
		link := "http://eck.openebs100.io:5603/app/kibana#/discover?_g=(refreshInterval:(pause:!t,value:5000),time:(from:'" + startedAt + "',to:'" + finishedAt + "'))&_a=(columns:!(log),filters:!(('$state':(store:appState),meta:(alias:!n,disabled:!f,index:faa59410-ca5f-11e9-834d-e7e11a373ae5,key:pipeline_id,negate:!f,params:(query:'" + strconv.Itoa(pipelineID) + "'),type:phrase,value:'" + strconv.Itoa(pipelineID) + "'),query:(match:(pipeline_id:(query:'" + strconv.Itoa(pipelineID) + "',type:phrase)))),('$state':(store:appState),meta:(alias:!n,disabled:!f,index:faa59410-ca5f-11e9-834d-e7e11a373ae5,key:commit_id,negate:!f,params:(query:'" + sha + "'),type:phrase,value:'" + sha + "'),query:(match:(commit_id:(query:'" + sha + "',type:phrase))))),index:faa59410-ca5f-11e9-834d-e7e11a373ae5,interval:auto,query:(language:kuery,query:''),sort:!('@timestamp',desc))"
		return link
	}
	link := "http://eck.openebs100.io:5603/app/kibana#/discover?_g=(refreshInterval:(pause:!f,value:5000),time:(from:now-3h,to:now))&_a=(columns:!(log),filters:!(('$state':(store:appState),meta:(alias:!n,disabled:!f,index:faa59410-ca5f-11e9-834d-e7e11a373ae5,key:pipeline_id,negate:!f,params:(query:'" + strconv.Itoa(pipelineID) + "'),type:phrase,value:'" + strconv.Itoa(pipelineID) + "'),query:(match:(pipeline_id:(query:'" + strconv.Itoa(pipelineID) + "',type:phrase)))),('$state':(store:appState),meta:(alias:!n,disabled:!f,index:faa59410-ca5f-11e9-834d-e7e11a373ae5,key:commit_id,negate:!f,params:(query:'" + sha + "'),type:phrase,value:'" + sha + "'),query:(match:(commit_id:(query:'" + sha + "',type:phrase))))),index:faa59410-ca5f-11e9-834d-e7e11a373ae5,interval:auto,query:(language:kuery,query:''),sort:!('@timestamp',desc))"
	return link
}

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
		fmt.Printf("No rows were returned for %s table of %d jobID \n", t.TableName, t.JobID)
		return "NA"
	case nil:
		return releaseTag
	default:
		fmt.Println(err)
		return "NA"
	}

}
