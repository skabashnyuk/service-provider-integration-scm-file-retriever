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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/imroc/req/v3"
	"github.com/jarcoal/httpmock"

	"github.com/stretchr/testify/assert"
)

func TestGetFileHead(t *testing.T) {
	mockResponse, _ := json.Marshal(map[string]interface{}{
		"name":         "myfile",
		"size":         582,
		"download_url": "https://raw.githubusercontent.com/foo-user/foo-repo/HEAD/myfile",
	})

	client := req.C()
	httpmock.ActivateNonDefault(client.GetClient())
	httpmock.RegisterResponder("GET", "https://api.github.com/repos/foo-user/foo-repo/contents/myfile?ref=HEAD", func(request *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(http.StatusOK, string(mockResponse))
		resp.Header.Set("Content-Type", "application/json; charset=utf-8")
		return resp, nil
	})

	r1, err := detect(context.TODO(), "https://github.com/foo-user/foo-repo", "myfile", "HEAD", client, HeaderStruct{Authorization: "foo"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	assert.Equal(t, "https://raw.githubusercontent.com/foo-user/foo-repo/HEAD/myfile", r1)
}

func TestGetFileHeadGitSuffix(t *testing.T) {
	mockResponse, _ := json.Marshal(map[string]interface{}{
		"name":         "myfile",
		"size":         582,
		"download_url": "https://raw.githubusercontent.com/foo-user/foo-repo/HEAD/myfile",
	})

	client := req.C()
	httpmock.ActivateNonDefault(client.GetClient())
	httpmock.RegisterResponder("GET", "https://api.github.com/repos/foo-user/foo-repo/contents/myfile?ref=HEAD", func(request *http.Request) (*http.Response, error) {
		respBody := string(mockResponse)
		resp := httpmock.NewStringResponse(http.StatusOK, respBody)
		resp.Header.Set("Content-Type", "application/json; charset=utf-8")
		return resp, nil
	})
	r1, err := detect(context.TODO(), "https://github.com/foo-user/foo-repo.git", "myfile", "HEAD", client, HeaderStruct{Authorization: "foo"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	assert.Equal(t, "https://raw.githubusercontent.com/foo-user/foo-repo/HEAD/myfile", r1)
}

func TestGetFileOnBranch(t *testing.T) {
	mockResponse, _ := json.Marshal(map[string]interface{}{
		"name":         "myfile",
		"size":         582,
		"download_url": "https://raw.githubusercontent.com/foo-user/foo-repo/v0.1.0/myfile",
	})

	client := req.C()
	httpmock.ActivateNonDefault(client.GetClient())
	httpmock.RegisterResponder("GET", "https://api.github.com/repos/foo-user/foo-repo/contents/myfile?ref=v0.1.0", func(request *http.Request) (*http.Response, error) {
		respBody := string(mockResponse)
		resp := httpmock.NewStringResponse(http.StatusOK, respBody)
		resp.Header.Set("Content-Type", "application/json; charset=utf-8")
		return resp, nil
	})

	r1, err := detect(context.TODO(), "https://github.com/foo-user/foo-repo", "myfile", "v0.1.0", client, HeaderStruct{Authorization: "foo"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	assert.Equal(t, "https://raw.githubusercontent.com/foo-user/foo-repo/v0.1.0/myfile", r1)
}

func TestGetFileOnCommitId(t *testing.T) {
	mockResponse, _ := json.Marshal(map[string]interface{}{
		"name":         "myfile",
		"size":         582,
		"download_url": "https://raw.githubusercontent.com/foo-user/foo-repo/efaf08a367921ae130c524db4a531b7696b7d967/myfile",
	})

	client := req.C()
	httpmock.ActivateNonDefault(client.GetClient())
	httpmock.RegisterResponder("GET", "https://api.github.com/repos/foo-user/foo-repo/contents/myfile?ref=efaf08a367921ae130c524db4a531b7696b7d967", func(request *http.Request) (*http.Response, error) {
		respBody := string(mockResponse)
		resp := httpmock.NewStringResponse(http.StatusOK, respBody)
		resp.Header.Set("Content-Type", "application/json; charset=utf-8")
		return resp, nil
	})

	r1, err := detect(context.TODO(), "https://github.com/foo-user/foo-repo", "myfile", "efaf08a367921ae130c524db4a531b7696b7d967", client, HeaderStruct{Authorization: "foo"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	assert.Equal(t, "https://raw.githubusercontent.com/foo-user/foo-repo/efaf08a367921ae130c524db4a531b7696b7d967/myfile", r1)
}

func TestGetUnexistingFile(t *testing.T) {
	mockResponse := "{\"message\":\"File Is Not Found\"}"

	client := req.C()
	httpmock.ActivateNonDefault(client.GetClient())
	httpmock.RegisterResponder("GET", "https://api.github.com/repos/foo-user/foo-repo/contents/myfile?ref=efaf08a367921ae130c524db4a531b7696b7d967", func(request *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(http.StatusNotFound, mockResponse)
		resp.Header.Set("Content-Type", "application/json; charset=utf-8")
		return resp, nil
	})

	_, err := detect(context.TODO(), "https://github.com/foo-user/foo-repo", "myfile", "efaf08a367921ae130c524db4a531b7696b7d967", client, HeaderStruct{Authorization: "foo"})
	if err == nil {
		t.Error("error expected")
	}
	assert.True(t, errors.Is(err, &UnauthorizedError{}))
	assert.Equal(t, "detection failed: Request to SCM server was unauthorized or resource is not found", fmt.Sprint(err))

}
