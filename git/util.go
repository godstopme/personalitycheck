package git

import (
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	githttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

func pushRefSpec(branchName string) []config.RefSpec {
	return []config.RefSpec{
		config.RefSpec(branchName),
	}
}

func push(repository *git.Repository, refSpecs []config.RefSpec, credentials *githttp.BasicAuth) error {
	return repository.Push(&git.PushOptions{
		RemoteName: cloneRemoteName,
		RefSpecs:   refSpecs,
		Auth:       credentials,
	})
}
