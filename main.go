package main

import (
	"encoding/json"
	"net/http"

	"github.com/godstopme/personalitycheck/git"
	"github.com/godstopme/personalitycheck/parse"
	"google.golang.org/appengine"
)

type personalityCheckRequest struct {
	GitURI string `json:"repository"`
	Author string `json:"email"`
}

func readRequestBody(request *http.Request) (*personalityCheckRequest, error) {
	decoder := json.NewDecoder(request.Body)

	var body personalityCheckRequest
	err := decoder.Decode(&body)
	if err != nil {
		return nil, err
	}

	return &body, nil
}

func handler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		http.NotFound(writer, request)

		return
	}

	body, err := readRequestBody(request)
	if err != nil {
		writer.WriteHeader(400)

		return
	}

	repo, err := git.PrepareRepository(body.GitURI, "", "")
	if err != nil {
		panic(err)
	}

	commitHash, err := repo.CreateCommit(body.Author)
	if err != nil {
		panic(err)
	}

	branch, err := repo.PushToNewBranch()
	if err != nil {
		panic(err)
	}

	verifyResult, err := parse.ExtractProfileLink(repo.URL(), commitHash)
	if err != nil {
		panic(err)
	}

	js, err := json.Marshal(verifyResult)
	if err != nil {
		panic(err)
	}

	writer.Write(js)

	err = repo.DeleteBranch(branch)
	if err != nil {
		panic(err)
	}
}

func main() {
	http.HandleFunc("/", handler)

	appengine.Main()
}
