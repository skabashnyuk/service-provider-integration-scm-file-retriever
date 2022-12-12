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
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/imroc/req/v3"
)

const defaultHttpTimeout = 5 * time.Second

type GitFile struct {
	fetcher TokenFetcher
	client  *req.Client
}

type errorMessage struct {
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
	var errMsg errorMessage
	resp, err := request.
		SetError(&errMsg).
		Get(fileUrl)
	if err != nil {
		zap.L().Error("Failed to make file content request", zap.Error(err))
		return nil, &InternalError{fmt.Sprintf("File content request failed: %s", err.Error()), err}
	}
	zap.L().Debug(fmt.Sprintf(
		"GitHub file call response code: %d", resp.GetStatusCode()))
	// 2xx
	if resp.IsSuccess() {
		return io.NopCloser(bytes.NewBuffer(resp.Bytes())), nil
	}
	// >= 400
	if resp.IsError() {
		switch resp.StatusCode {
		case http.StatusBadRequest:
			return nil, &InternalError{"File content request has wrong format", nil}
		case http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
			return nil, &UnauthorizedError{}
		default:
			return nil, &InternalError{fmt.Sprintf("Unexpected status code returned when make file content request: %d. Message: %s", resp.StatusCode, errMsg.Message), nil}
		}
	}
	// strange cases like 3xx etc
	return nil, &InternalError{fmt.Sprintf("File content request returned unexpected code: %d. Content dump: %s", resp.StatusCode, resp.Dump()), nil}
}

// New creates a new *GitFile instance
func New(fetcher TokenFetcher) *GitFile {
	return &GitFile{fetcher: fetcher, client: req.C().SetTimeout(defaultHttpTimeout)}
}

func Default() *GitFile {
	return &GitFile{fetcher: NewSpiTokenFetcher(), client: req.C().SetTimeout(defaultHttpTimeout)}
}
