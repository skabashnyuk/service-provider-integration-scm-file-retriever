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
	"bytes"
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"io"
	"time"

	"github.com/imroc/req/v3"
)

var (
	failedFileRequestError     = errors.New("file request failed")
	undefinedFileResponseError = errors.New("file request returned an unknown result")
)

const defaultHttpTimeout = 5 * time.Second

type GitFile struct {
	fetcher TokenFetcher
	client  *req.Client
}

type ErrorMessage struct {
	Message string `json:"message"`
}

// GetFileContents is a main entry function allowing to retrieve file content from the SCM provider.
// It expects three file location parameters, from which the repository URL and path to the file are mandatory,
// and optional Git reference for the branch/tags/commitIds.
// Function type parameter is a callback used when user authentication is needed in order to retrieve the file,
// that function will be called with the URL to OAuth service, where user need to be redirected.
func (g *GitFile) GetFileContents(ctx context.Context, namespace, repoUrl, filepath, ref string, callback func(ctx context.Context, url string)) (io.ReadCloser, error) {
	headerStruct, err := buildAuthHeader(ctx, namespace, repoUrl, g.fetcher, callback)
	if err != nil {
		return nil, err
	}
	fileUrl, err := detect(ctx, repoUrl, filepath, ref, g.client, *headerStruct)
	if err != nil {
		return nil, err
	}

	request := g.client.R().SetContext(ctx).SetBearerAuthToken(headerStruct.Authorization)
	var errMsg ErrorMessage
	resp, err := request.
		SetError(&errMsg).
		Get(fileUrl)
	if err != nil {
		zap.L().Error("Failed to make GitHub URL call", zap.Error(err))
		return nil, fmt.Errorf("GitHub file call failed: %w", err)
	}
	zap.L().Debug(fmt.Sprintf(
		"GitHub file call response code: %d", resp.GetStatusCode()))
	if resp.IsSuccess() {
		return io.NopCloser(bytes.NewBuffer(resp.Bytes())), nil
	}
	if resp.IsError() {
		return nil, fmt.Errorf("%w. Status code: %d. Message: %s", failedFileRequestError, resp.GetStatusCode(), errMsg.Message)
	}
	return nil, fmt.Errorf("%w. Status code: %d. Content: %s", undefinedFileResponseError, resp.GetStatusCode(), resp.Dump())
}

// New creates a new *GitFile instance
func New(fetcher TokenFetcher) *GitFile {
	return &GitFile{fetcher: fetcher, client: req.C().SetTimeout(defaultHttpTimeout)}
}

func Default() *GitFile {
	return &GitFile{fetcher: NewSpiTokenFetcher(), client: req.C().SetTimeout(defaultHttpTimeout)}
}
