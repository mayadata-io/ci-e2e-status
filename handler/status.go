package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang/glog"
)

type GitLabStatus struct {
	BaseURL  string    `json:"url"`
	Status   string    `json:"status"`
	Response int       `json:"response"`
	Updated  time.Time `json:"updated"`
	Message  string    `json:"message"`
}

func gitLabStatus(url string) bool {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
		return false
	}
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return true
	} else {
		return false
	}
}

func StatusGitLab(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers:", "Origin, Content-Type, X-Auth-Token, Authorization")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	s := GitLabStatus{}
	resp, err := http.Get(BaseURL)
	if err != nil {
		s = GitLabStatus{
			BaseURL:  BaseURL,
			Status:   "offline",
			Response: 000,
			Updated:  time.Now(),
			Message:  err.Error(),
		}
		jOut, err := json.Marshal(s)
		if err != nil {
			glog.Infoln(err)
		}
		w.Write(jOut)
	} else {
		// Print the HTTP Status Code and Status Name
		fmt.Println("HTTP Response Status:", resp.StatusCode, http.StatusText(resp.StatusCode))

		if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
			s = GitLabStatus{
				BaseURL:  BaseURL,
				Status:   "online",
				Response: resp.StatusCode,
				Updated:  time.Now(),
				Message:  resp.Status,
			}
		} else {
			s = GitLabStatus{
				BaseURL:  BaseURL,
				Status:   "offline",
				Response: resp.StatusCode,
				Updated:  time.Now(),
				Message:  err.Error(),
			}
		}
		out, err := json.Marshal(s)
		if err != nil {
			log.Fatal(err)
		}
		w.Write(out)
	}
}
