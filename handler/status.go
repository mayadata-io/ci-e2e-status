package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang/glog"
)

type GitLabStatus struct {
	BaseURL  string    `json:"url"`
	Status   string    `json:"status"`
	Response int       `json:"response"`
	Updated  time.Time `json:"updated"`
	Message  string    `json:"message"`
	Version  string    `json:"version"`
}

type GitLabVersion struct {
	Version string `json:"version"`
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
	v := GitLabVersion{}
	url := fmt.Sprintf("%s/api/v4/version", BaseURL)
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return
	}
	token, ok := os.LookupEnv(token)
	if !ok {
		glog.Fatalf("TOKEN environment variable required")
	}

	req.Header.Add("PRIVATE-TOKEN", token)
	req.Header.Add("Content-Type", "application/json")
	// resp, err := http.Get(fmt.Sprintf("%s/api/v4/version", BaseURL))
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	if err != nil {
		// err := json.Unmarshal(resp., v)

		s = GitLabStatus{
			BaseURL:  BaseURL,
			Status:   "offline",
			Response: 000,
			Updated:  time.Now(),
			Message:  err.Error(),
			// Version:  v.Version,
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
			body, readErr := ioutil.ReadAll(resp.Body)
			if readErr != nil {
				log.Fatal(readErr)
			}
			jsonErr := json.Unmarshal(body, &v)
			if jsonErr != nil {
				log.Fatal(jsonErr)
			}
			s = GitLabStatus{
				BaseURL:  BaseURL,
				Status:   "online",
				Response: resp.StatusCode,
				Updated:  time.Now(),
				Message:  resp.Status,
				Version:  v.Version,
			}
		} else {

			s = GitLabStatus{
				BaseURL:  BaseURL,
				Status:   "offline",
				Response: resp.StatusCode,
				Updated:  time.Now(),
				Message:  err.Error(),
				// Version:  v.Version,
			}
		}
		out, err := json.Marshal(s)
		if err != nil {
			log.Fatal(err)
		}
		w.Write(out)
	}
}
