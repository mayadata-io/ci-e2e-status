package handler

// PlatformID assignment from gitlab repository
const (
	PACKETID = "27"
	KONVOYID = "34"
	MAYAID   = "7"
	JIVAID   = "6"
	ISTGTID  = "5"
	ZFSID    = "8"
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

// PlatformPipeline struct
type PlatformPipeline struct {
	ID     int    `json:"id"`
	Sha    string `json:"sha"`
	Ref    string `json:"ref"`
	Status string `json:"status"`
	WebURL string `json:"web_url"`
}

// Pipeline struct
type Pipeline []struct {
	ID     int    `json:"id"`
	Sha    string `json:"sha"`
	Ref    string `json:"ref"`
	Status string `json:"status"`
	WebURL string `json:"web_url"`
	Jobs   Jobs   `json:"jobs"`
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
}

// Jobssummary struct
type Jobssummary struct {
	PipelineID int    `json:"pipelineid"`
	ID         int    `json:"id"`
	Status     string `json:"status"`
	Stage      string `json:"stage"`
	Name       string `json:"name"`
	Ref        string `json:"ref"`
	CreatedAt  string `json:"created_at"`
	StartedAt  string `json:"started_at"`
	FinishedAt string `json:"finished_at"`
	JobLogURL  string `json:"job_log_url"`
}

// TriggredID contains the details of a gitlab pipelines
type TriggredID struct {
	ID       int `json:"id"`
	BuildPID int `json:"gke_trigger_pid"`
}

// pipelineSummary contains the details of a gitlab pipelines
type pipelineSummary struct {
	ID     int           `json:"id"`
	Sha    string        `json:"sha"`
	Ref    string        `json:"ref"`
	Status string        `json:"status"`
	WebURL string        `json:"web_url"`
	LogURL string        `json:"kibana_url"`
	Jobs   []Jobssummary `json:"jobs"`
}

type dashboard struct {
	Dashboard []pipelineSummary `json:"dashboard"`
}

// BuildJobs struct
type BuildJobs []struct {
	ID         int    `json:"id"`
	Status     string `json:"status"`
	Stage      string `json:"stage"`
	Name       string `json:"name"`
	Ref        string `json:"ref"`
	CreatedAt  string `json:"created_at"`
	StartedAt  string `json:"started_at"`
	FinishedAt string `json:"finished_at"`
	Message    string `json:"message"`
	AuthorName string `json:"author_name"`
	WebURL     string `json:"web_url"`
	Commit     struct {
		Message    string `json:"message"`
		AuthorName string `json:"author_name"`
	} `json:"commit"`
}

// BuildJobssummary contains the details of builds job for database
type BuildJobssummary struct {
	PipelineID int    `json:"pipelineid"`
	ID         int    `json:"id"`
	Status     string `json:"status"`
	Stage      string `json:"stage"`
	Name       string `json:"name"`
	Ref        string `json:"ref"`
	CreatedAt  string `json:"created_at"`
	StartedAt  string `json:"started_at"`
	FinishedAt string `json:"finished_at"`
	Message    string `json:"message"`
	AuthorName string `json:"author_name"`
}

// BuildpipelineSummary contains the details of a builds pipelines
type BuildpipelineSummary struct {
	Project      string             `json:"project"`
	ID           int                `json:"id"`
	Sha          string             `json:"sha"`
	Ref          string             `json:"ref"`
	Status       string             `json:"status"`
	WebURL       string             `json:"web_url"`
	PacketV11PID string             `json:"packet_v11_pid"`
	PacketV12PID string             `json:"packet_v12_pid"`
	PacketV13PID string             `json:"packet_v13_pid"`
	KonvoyPID    string             `json:"konvoy_pid"`
	Jobs         []BuildJobssummary `json:"jobs"`
}

// Builddashboard contains the details related to builds
type Builddashboard struct {
	Dashboard []BuildpipelineSummary `json:"dashboard"`
}
