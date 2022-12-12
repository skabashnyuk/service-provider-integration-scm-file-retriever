// Copyright (c) 2022 Red Hat, Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gitfile

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/imroc/req/v3"
	"go.uber.org/zap"
)

type GithubFile struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Size        int32  `json:"size"`
	Encoding    string `json:"encoding"`
	DownloadUrl string `json:"download_url"`
}

var GithubAPITemplate = "https://api.github.com/repos/%s/%s/contents/%s"
var GithubURLRegexp = regexp.MustCompile(`(?Um)^(?:https)(?:\:\/\/)github.com/(?P<repoUser>[^/]+)/(?P<repoName>[^/]+)(.git)?$`)
var GithubURLRegexpNames = GithubURLRegexp.SubexpNames()

// GitHubScmProvider implements Detector to detect GitHub URLs.
type GitHubScmProvider struct {
}

func (d *GitHubScmProvider) detect(ctx context.Context, repoUrl, filepath, ref string, cl *req.Client, auth HeaderStruct) (bool, string, error) {
	if len(repoUrl) == 0 || !GithubURLRegexp.MatchString(repoUrl) {
		return false, "", nil
	}

	result := GithubURLRegexp.FindAllStringSubmatch(repoUrl, -1)
	m := map[string]string{}
	for i, n := range result[0] {
		m[GithubURLRegexpNames[i]] = n
	}
	request := cl.R().SetContext(ctx).SetBearerAuthToken(auth.Authorization)
	if ref != "" {
		request.SetQueryParam("ref", ref)
	}

	var file GithubFile
	var errMsg errorMessage
	resp, err := request.
		SetResult(&file).
		SetError(&errMsg).
		Get(fmt.Sprintf(GithubAPITemplate, m["repoUser"], m["repoName"], filepath))
	if err != nil {
		zap.L().Error("Failed to make GitHub API call", zap.Error(err))
		return true, "", &InternalError{fmt.Sprintf("GitHub API request failed: %s", err.Error()), err}
	}
	statusCode := resp.StatusCode
	zap.L().Debug(fmt.Sprintf(
		"GitHub API request response code: %d", statusCode))
	// 2xx
	if resp.IsSuccess() {
		return true, file.DownloadUrl, nil
	}
	// 4xx and 5xx
	if resp.IsError() {
		switch statusCode {
		case http.StatusBadRequest:
			return true, "", &InternalError{"GitHub API request has wrong format", nil}
		case http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
			return true, "", &UnauthorizedError{}
		default:
			return true, "", &InternalError{fmt.Sprintf("Unexpected status code returned from GitHub API: %d. Message: %s", statusCode, errMsg.Message), nil}
		}
	}
	// strange cases like 3xx etc
	return true, "", &InternalError{fmt.Sprintf("GitHub API request returned unexpected code: %d. Content dump: %s", statusCode, resp.Dump()), nil}

}
