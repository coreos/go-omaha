// Copyright 2017 CoreOS, Inc.
//
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

package client

import (
	"net/http"
)

// httpError implements error and net.Error for http responses.
type httpError struct {
	*http.Response
}

func (he *httpError) Error() string {
	return "http error: " + he.Status
}

func (he *httpError) Timeout() bool {
	switch he.StatusCode {
	case http.StatusRequestTimeout: // 408
		return true
	case http.StatusGatewayTimeout: // 504
		return true
	default:
		return false
	}
}

func (he *httpError) Temporary() bool {
	if he.Timeout() {
		return true
	}
	switch he.StatusCode {
	case http.StatusTooManyRequests: // 429
		return true
	case http.StatusInternalServerError: // 500
		return true
	case http.StatusBadGateway: // 502
		return true
	case http.StatusServiceUnavailable: // 503
		return true
	default:
		return false
	}
}
