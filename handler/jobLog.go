package handler

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
)

func GetJobLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers:", "Origin, Content-Type, X-Auth-Token, Authorization")
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	platform := vars["platform"] //openshift
	platform = getPlatform(platform)
	// branch := strings.Replace(vars["branch"], "-", "_", -1) // openebs-jiva
	id := vars["id"]
	gitLabJob := fmt.Sprintf("https://gitlab.openebs.ci/openebs/%s/-/jobs/%s/raw", platform, id)
	resp, err := http.Get(gitLabJob)
	if err != nil {
		glog.Infoln(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			glog.Fatal(err)
		}
		// bodyString := string(bodyBytes)
		w.Write(bodyBytes)
		// glog.Info(bodyString)
	} // https://gitlab.openebs.ci/openebs/e2e-nativek8s/-/jobs/354396/raw

}

func getPlatform(p string) string {
	switch p {
	case "konvoy":
		return "e2e-konvoy"
	case "openshift":
		return "e2e-openshift"
	case "nativek8s":
		return "e2e-nativek8s"
	default:
		return "NA"

	}
}
