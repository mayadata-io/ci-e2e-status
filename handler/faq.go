package handler

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// FaqHandler print readme content in /faq endpoint
func FaqHandler(w http.ResponseWriter, r *http.Request) {

	url := "https://raw.githubusercontent.com/mayadata-io/ci-e2e-dashboard/staging/README.md"
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	responseString := string(responseData)
	fmt.Fprint(w, responseString)
}
