// Copyright (c) 2021 - 2022 Red Hat, Inc.
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

import "fmt"

type UnauthorizedError struct {
}

func (e *UnauthorizedError) Error() string {
	return "Request to SCM server was unauthorized or resource is not found"
}

type InvalidRequestError struct {
	message  string
	repoUrl  string
	filePath string
}

func (e *InvalidRequestError) Error() string {
	return fmt.Sprintf("%s. Repository URL: %s, file path: %s", e.message, e.repoUrl, e.filePath)
}

type InternalError struct {
	message string
	cause   error
}

func (e *InternalError) Error() string {
	return "Service internal error happened. Base message: " + e.message
}
