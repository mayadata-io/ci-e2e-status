package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/golang/glog"
)

type GitLabRunner struct {
	Active      bool   `json:"active"`
	Description string `json:"description"`
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Online      bool   `json:"online"`
	Status      string `json:"status"`
}

func HandleRunners(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers:", "Origin, Content-Type, X-Auth-Token, Authorization")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// s := GitLabRunner{}
	v := []GitLabRunner{}
	url := fmt.Sprintf("%s/api/v4/runners/all", BaseURL)
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
	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}
	jsonErr := json.Unmarshal(body, &v)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	fmt.Printf("body: \n %#v \n", v)
	out, err := json.Marshal(v)
	if err != nil {
		log.Fatal(err)
	}
	w.Write(out)

}
