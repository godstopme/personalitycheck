package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	githttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"

	"gopkg.in/src-d/go-billy.v4/memfs"

	uuid "github.com/satori/go.uuid"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

const cloneRemoteName = "origin"

var userProfileRegex *regexp.Regexp

func init() {
	userProfileRegex = regexp.MustCompile(`<.*"AvatarStack-body".*>\s*<a.*href="(.*)"`)
}

func createCommit(repository *git.Repository, email string) (plumbing.Hash, error) {
	worktree, err := repository.Worktree()
	if err != nil {
		return plumbing.ZeroHash, err
	}

	return worktree.Commit("", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "author",
			Email: email,
			When:  time.Now(),
		},
	})
}

func push(repository *git.Repository, refSpecs []config.RefSpec, credentials *githttp.BasicAuth) error {
	return repository.Push(&git.PushOptions{
		RemoteName: cloneRemoteName,
		RefSpecs:   refSpecs,
		Auth:       credentials,
	})
}

type VerifiedLinkResult struct {
	Verified    bool   `json:"verified"`
	ProfileLink string `json:"profile"`
}

func robustHTTPGet(uri string) (string, error) {
	for {
		time.Sleep(time.Millisecond * 500) // sleeping here because there is a timegap between push & github page availability
		response, err := http.Get(uri)
		if err != nil {
			return "", err
		}

		if response.StatusCode == 404 {
			continue
		}

		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return "", nil
		}

		return string(body), nil
	}
}

func checkProfileLink(repoURI string, commitHash plumbing.Hash) (*VerifiedLinkResult, error) {
	hash := hex.EncodeToString(commitHash[:])
	uriTemplate := "%s/commit/%s"
	uri := fmt.Sprintf(uriTemplate, strings.TrimSuffix(repoURI, path.Ext(repoURI)), hash)

	html, err := robustHTTPGet(uri)
	if err != nil {
		return nil, err
	}

	var result VerifiedLinkResult

	matches := userProfileRegex.FindStringSubmatch(html)
	if matches == nil {
		return &result, nil
	}

	url, _ := url.Parse(repoURI)            // TODO: validate uri in main func
	profileLink, _ := url.Parse(matches[1]) // take the actual capturing group

	result.Verified = true
	result.ProfileLink = url.ResolveReference(profileLink).String()

	return &result, nil
}

func pushRefSpec(branchName string) []config.RefSpec {
	return []config.RefSpec{
		config.RefSpec(branchName),
	}
}

func main() {
	repository, err := git.Init(memory.NewStorage(), memfs.New())
	if err != nil {
		panic(err)
	}

	_, err = repository.CreateRemote(&config.RemoteConfig{
		Name: cloneRemoteName,
		URLs: []string{"https://github.com/godstopme/wolframalphakiller.git"},
	})
	if err != nil {
		panic(err)
	}

	commitHash, err := createCommit(repository, "nosov@nodeart.io")
	if err != nil {
		panic(err)
	}

	credentials := &githttp.BasicAuth{
		Username: "",
		Password: "",
	}
	branchName := uuid.NewV4().String()

	// git push origin master:branchName
	err = push(repository, pushRefSpec("refs/heads/master:refs/heads/"+branchName), credentials)
	if err != nil {
		panic(err)
	}

	verifyResult, err := checkProfileLink("https://github.com/godstopme/wolframalphakiller.git", commitHash)
	if err != nil {
		panic(err)
	}
	js, err := json.MarshalIndent(verifyResult, "", "  ")
	if err != nil {
		panic(err)
	}
	println(string(js))
	// git push --delete origin branchName
	err = push(repository, pushRefSpec(":refs/heads/"+branchName), credentials)

	if err != nil {
		panic(err)
	}
}
