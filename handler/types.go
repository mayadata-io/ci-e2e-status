package handler

// PlatformID assignment from gitlab repository
const (
	PACKETID    = "27"
	KONVOYID    = "34"
	OPENSHIFTID = "36"
	MAYAID      = "7"
	JIVAID      = "6"
	ISTGTID     = "5"
	ZFSID       = "8"
	NATIVEK8SID = "43"
)

// BranchName assignment from gitlab repository
const (
	GROUPNAME   = "openebs"
	MAYABRANCH  = "master"
	JIVABRANCH  = "master"
	ISTGTBRANCH = "replication"
	ZFSBRANCH   = "develop"
)

var project string

const (
	token = "TOKEN"
)

// BaseURL for gitlab
var BaseURL = "https://gitlab.openebs.ci/"

// Pipeline struct
type Pipeline []struct {
	ID        int    `json:"id"`
	Sha       string `json:"sha"`
	Ref       string `json:"ref"`
	Status    string `json:"status"`
	WebURL    string `json:"web_url"`
	Jobs      Jobs   `json:"jobs"`
	CreatedAt string `json:"created_at"`
}

// Jobs struct
type Jobs []struct {
	ID         int    `json:"id"`
	Status     string `json:"status"`
	Stage      string `json:"stage"`
	Name       string `json:"name"`
	Ref        string `json:"ref"`
	CreatedAt  string `json:"created_at"`
	StartedAt  string `json:"started_at"`
	FinishedAt string `json:"finished_at"`
	WebURL     string `json:"web_url"`
}

// BuildJobssummary contains the details of builds job for database
type BuildJobssummary struct {
	PipelineID   int    `json:"pipelineid"`
	ID           int    `json:"id"`
	Status       string `json:"status"`
	Stage        string `json:"stage"`
	Name         string `json:"name"`
	Ref          string `json:"ref"`
	GithubReadme string `json:"github_readme"`
	CreatedAt    string `json:"created_at"`
	StartedAt    string `json:"started_at"`
	FinishedAt   string `json:"finished_at"`
	WebURL       string `json:"web_url"`
}

// OpenshiftpipelineSummary contains the details of a openshifts pipelines
type OpenshiftpipelineSummary struct {
	Project      string             `json:"project"`
	ID           int                `json:"id"`
	Sha          string             `json:"sha"`
	Ref          string             `json:"ref"`
	Status       string             `json:"status"`
	WebURL       string             `json:"web_url"`
	OpenshiftPID string             `json:"openshift_pid"`
	LogURL       string             `json:"kibana_url" `
	ReleaseTag   string             `json:"release_tag"`
	CreatedAt    string             `json:"created_at"`
	Jobs         []BuildJobssummary `json:"jobs"`
}

// Openshiftdashboard contains the details related to openshifts
type Openshiftdashboard struct {
	Dashboard []OpenshiftpipelineSummary `json:"dashboard"`
}
type PipeData struct {
	Pipeline OpenshiftpipelineSummary `json:"pipeline"`
}
