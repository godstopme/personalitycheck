package git

import (
	"encoding/hex"
	"net/url"
	"path"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-git.v4/storage/memory"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	githttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

const dummyAuthorName = ""
const cloneRemoteName = "origin"

type Repository struct {
	repository  *git.Repository
	uri         *url.URL
	credentials *githttp.BasicAuth
}

func PrepareRepository(uri string, username string, password string) (*Repository, error) {
	uri = strings.TrimSuffix(uri, path.Ext(uri))
	validatedURI, err := url.Parse(uri)

	repository, err := git.Init(memory.NewStorage(), memfs.New())
	if err != nil {
		return nil, err
	}

	_, err = repository.CreateRemote(&config.RemoteConfig{
		Name: cloneRemoteName,
		URLs: []string{validatedURI.String() + ".git"},
	})
	if err != nil {
		return nil, err
	}

	object := &Repository{
		repository: repository,
		uri:        validatedURI,
		credentials: &githttp.BasicAuth{
			Username: username,
			Password: password,
		},
	}

	return object, nil
}

func (r *Repository) CreateCommit(authorEmail string) (string, error) {
	worktree, err := r.repository.Worktree()
	if err != nil {
		return "", err
	}

	hash, err := worktree.Commit("", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "author",
			Email: authorEmail,
			When:  time.Now(),
		},
	})
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hash[:]), nil
}

func (r *Repository) PushToNewBranch() (string, error) {
	branchName := uuid.NewV4().String()
	fullName := "refs/heads/" + branchName

	refSpec := pushRefSpec("refs/heads/master:" + fullName)

	err := push(r.repository, refSpec, r.credentials)
	if err != nil {
		return "", err
	}

	return fullName, nil
}

func (r *Repository) DeleteBranch(branchName string) error {
	refSpec := pushRefSpec(":" + branchName)

	err := push(r.repository, refSpec, r.credentials)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) URL() *url.URL {
	return r.uri
}
