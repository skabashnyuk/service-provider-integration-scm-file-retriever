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
	"errors"
	"fmt"
	"regexp"

	"github.com/imroc/req/v3"
	"go.uber.org/zap"
)

var (
	unexpectedStatusCodeError = errors.New("unexpected status code from GitHub API")
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
	var errMsg ErrorMessage
	resp, err := request.
		SetResult(&file).
		SetError(&errMsg).
		Get(fmt.Sprintf(GithubAPITemplate, m["repoUser"], m["repoName"], filepath))
	if err != nil {
		zap.L().Error("Failed to make GitHub API call", zap.Error(err))
		return true, "", fmt.Errorf("GitHub API call failed: %w", err)
	}
	statusCode := resp.StatusCode
	zap.L().Debug(fmt.Sprintf(
		"GitHub API call response code: %d", statusCode))
	if !resp.IsSuccess() {
		return true, "", fmt.Errorf("%w: %d. Response: %s", unexpectedStatusCodeError, statusCode, resp.String())
	}
	if resp.IsError() {
		return true, "", fmt.Errorf("%w: %d. Error message: %s", unexpectedStatusCodeError, statusCode, errMsg.Message)
	}
	return true, file.DownloadUrl, nil
}
