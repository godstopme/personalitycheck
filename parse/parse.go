package parse

import (
	"fmt"
	"net/url"
	"regexp"
)

var userProfileRegex *regexp.Regexp

func init() {
	userProfileRegex = regexp.MustCompile(`<.*?"AvatarStack-body".*?>\s*<a.*?href="(.*?)"`)
}

type VerifiedLinkResult struct {
	Verified    bool   `json:"verified"`
	ProfileLink string `json:"profile"`
}

func ExtractProfileLink(repoURI *url.URL, commitHash string) (*VerifiedLinkResult, error) {
	uriTemplate := "%s/commit/%s"
	uri := fmt.Sprintf(uriTemplate, repoURI, commitHash)

	html, err := robustHTTPGet(uri)
	if err != nil {
		return nil, err
	}

	var result VerifiedLinkResult

	matches := userProfileRegex.FindStringSubmatch(html)
	if matches == nil {
		return &result, nil
	}

	profileLink, _ := url.Parse(matches[1]) // take the actual capturing group

	result.Verified = true
	result.ProfileLink = repoURI.ResolveReference(profileLink).String()

	return &result, nil
}
