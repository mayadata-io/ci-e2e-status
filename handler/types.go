package handler

// PlatformID assignment from gitlab repository
const (
	AKSID    = "2"
	AWSID    = "1"
	EKSID    = "3"
	GCPID    = "4"
	GKEID    = "5"
	PACKETID = "6"
	MAYAID   = "8"
	JIVAID   = "7"
	ISTGTID  = "18"
	ZFSID    = "17"
)

// var token string

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
	ID            int                `json:"id"`
	Sha           string             `json:"sha"`
	Ref           string             `json:"ref"`
	Status        string             `json:"status"`
	WebURL        string             `json:"web_url"`
	GKETriggerPID string             `json:"gke_trigger_pid"`
	EKSTriggerPID string             `json:"eks_trigger_pid"`
	Jobs          []BuildJobssummary `json:"jobs"`
}

// Builddashboard contains the details related to builds
type Builddashboard struct {
	Dashboard []BuildpipelineSummary `json:"dashboard"`
}
