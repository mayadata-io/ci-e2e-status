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
	K8sVersion    string `json:"k8sVersion"`
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

// {
//     "projects" : [
//         {
//             "name": "nativek8s",
//             "id": "43",
//             "branches":[
//                 {
//                     "name": "release-branch",
//                     "releaseTagJob": "2P01-ZFS-LOCALPV-PROVISIONER-DEPLOY",
//                     "k8sVersion": "1S01-cluster-setup"
//                 },
//                 {
//                     "name": "lvm-localpv",
//                     "releaseTagJob": "2P01-LVM-LOCALPV-PROVISIONER-DEPLOY",
//                     "k8sVersion": "1S01-cluster-setup"
//                 }
//             ]
//         },
//         {
//             "name": "openshift",
//             "id": "36",
//             "branches":[
//                 {
//                     "name": "openebs-localpv",
//                     "releaseTagJob": "2LP01-OPENEBS-DEPLOY",
//                     "k8sVersion": "1LS01-CLUSTER-SETUP"
//                 },
//                 {
//                     "name": "openebs-cstor-csi",
//                     "releaseTagJob": "1CCO01-cluster-setup",
//                     "k8sVersion": "1JS01-cluster-setup"
//                 },
//                 {
//                     "name": "openebs-cstor",
//                     "releaseTagJob": "K9YC-OpenEBS",
//                     "k8sVersion": "PCZD-cluster-setup"
//                 },
//                 {
//                     "name": "jiva-operator",
//                     "releaseTagJob": "2IJO01-JIVA-OPERATOR",
//                     "k8sVersion": "1CJO01-cluster-setup"
//                 },
//                 {
//                     "name": "openebs-jiva",
//                     "releaseTagJob": "2JP01-OPENEBS-DEPLOY",
//                     "k8sVersion": "1JS01-cluster-setup"
//                 }
//             ]
//         },{
//             "name": "konvoy",
//             "id": "34",
//             "branches":[
//                 {
//                     "name": "jiva-operator",
//                     "releaseTagJob": "2IJO01-JIVA-OPERATOR",
//                     "k8sVersion": "1CJO01-cluster-setup"
//                 },
//                 {
//                     "name": "openebs-jiva",
//                     "releaseTagJob": "2JP01-OPENEBS-KONVOY-DEPLOY",
//                     "k8sVersion": "1JS01-cluster-setup"
//                 },
//                 {
//                     "name": "openebs-cstor-csi",
//                     "releaseTagJob": "2ICO01-CSTOR-OPERATOR",
//                     "k8sVersion": "1CCO01-cluster-setup"
//                 },
//                 {
//                     "name": "openebs-cstor",
//                     "releaseTagJob": "2IC01-OPENEBS-KONVOY-DEPLOY",
//                     "k8sVersion": "1CC01-cluster-setup"
//                 },
//                 {
//                     "name": "openebs-localpv",
//                     "releaseTagJob": "2LP01-OPENEBS-DEPLOY",
//                     "k8sVersion": "1LS01-cluster-setup"
//                 }
//             ]
//         }
//     ]
// }
