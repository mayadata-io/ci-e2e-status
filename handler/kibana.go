package handler

import "strconv"

// Kibanaloglink generate log link for platform pipelines
func Kibanaloglink(sha string, pipelineID int, status string, startedAt string, finishedAt string) string {
	if status == "success" || status == "failed" {
		link := "https://e2elogs.openebs.ci/app/kibana#/discover?_g=(refreshInterval:('$$hashKey':'object:188',display:Off,pause:!f,section:0,value:0),time:(from:'" + startedAt + "',mode:absolute,to:'" + finishedAt + "'))&_a=(columns:!(_source),filters:!(('$state':(store:appState),meta:(alias:!n,disabled:!f,index:'cluster-logs',key:commit_id,negate:!f,params:(query:'" + sha + "',type:phrase),type:phrase,value:'" + sha + "'),query:(match:(commit_id:(query:'" + sha + "',type:phrase)))),('$state':(store:appState),meta:(alias:!n,disabled:!f,index:'cluster-logs',key:pipeline_id,negate:!f,params:(query:'" + strconv.Itoa(pipelineID) + "',type:phrase),type:phrase,value:'" + strconv.Itoa(pipelineID) + "'),query:(match:(pipeline_id:(query:'" + strconv.Itoa(pipelineID) + "',type:phrase))))),index:'cluster-logs',interval:auto,query:(language:lucene,query:''),sort:!('@timestamp',desc))"
		return link
	}
	link := "https://e2elogs.openebs.ci/app/kibana#/discover?_g=(refreshInterval:('$$hashKey':'object:2232',display:'10+seconds',pause:!f,section:1,value:10000),time:(from:now-3h,mode:quick,to:now))&_a=(columns:!(_source),filters:!(('$state':(store:appState),meta:(alias:!n,disabled:!f,index:'cluster-logs',key:commit_id,negate:!f,params:(query:'" + sha + "',type:phrase),type:phrase,value:'" + sha + "'),query:(match:(commit_id:(query:'" + sha + "',type:phrase)))),('$state':(store:appState),meta:(alias:!n,disabled:!f,index:'cluster-logs',key:pipeline_id,negate:!f,params:(query:'" + strconv.Itoa(pipelineID) + "',type:phrase),type:phrase,value:'" + strconv.Itoa(pipelineID) + "'),query:(match:(pipeline_id:(query:'" + strconv.Itoa(pipelineID) + "',type:phrase))))),index:'cluster-logs',interval:auto,query:(language:lucene,query:''),sort:!('@timestamp',desc))"
	return link
}
