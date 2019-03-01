package main

import (
	"encoding/json"

	"github.com/godstopme/personalitycheck/parse"

	"github.com/godstopme/personalitycheck/git"
)

// integrational concurrency test, gae integration
func main() {
	repo, err := git.PrepareRepository("https://github.com/godstopme/wolframalphakiller.git", "", "")
	if err != nil {
		panic(err)
	}

	commitHash, err := repo.CreateCommit("nosov@nodeart.io")
	if err != nil {
		panic(err)
	}

	// git push origin master:branchName
	branch, err := repo.PushToNewBranch()
	if err != nil {
		panic(err)
	}

	verifyResult, err := parse.ExtractProfileLink(repo.URL(), commitHash)
	if err != nil {
		panic(err)
	}

	js, err := json.MarshalIndent(verifyResult, "", "  ")
	if err != nil {
		panic(err)
	}

	println(string(js))

	// git push --delete origin branchName
	//err = push(repository, pushRefSpec(":refs/heads/"+branchName), credentials)
	err = repo.DeleteBranch(branch)
	if err != nil {
		panic(err)
	}
}
