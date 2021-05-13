package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/golang/glog"
)

type Config struct {
	Projects []Project `json:"projects"`
}
type Project struct {
	Name     string   `json:"name"`
	ID       string   `json:"id"`
	Branches []Branch `json:"branches"`
}
type Branch struct {
	Name          string `json:"name"`
	ReleaseTagJob string `json:"releaseTagJob"`
}

func ReadConfig() Config {
	file, err := ioutil.ReadFile("./config/config.json")
	if err != nil {
		glog.Infoln(err)
	}
	config := Config{}
	err = json.Unmarshal(file, &config)
	if err != nil {
		glog.Infoln(err)
	}
	return config
}

func ViewConfig(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers:", "Origin, Content-Type, X-Auth-Token, Authorization")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	file, err := ioutil.ReadFile("./config/config.json")
	if err != nil {
		glog.Infoln(err)
	}
	config := Config{}
	err = json.Unmarshal(file, &config)
	if err != nil {
		glog.Infoln(err)
	}
	out, err := json.Marshal(config)
	if err != nil {
		log.Fatal(err)
	}
	w.Write(out)

}
