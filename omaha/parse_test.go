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

package omaha

import (
	"strings"
	"testing"
)

func TestCheckContentType(t *testing.T) {
	for _, tt := range []struct {
		ct string
		ok bool
	}{
		{"", true},
		{"text/xml", true},
		{"text/XML", true},
		{"application/xml", true},
		{"text/plain", false},
		{"xml", false},
		{"text/xml; charset=utf-8", true},
		{"text/xml; charset=UTF-8", true},
		{"text/xml; charset=ascii", false},
	} {
		err := checkContentType(tt.ct)
		if tt.ok && err != nil {
			t.Errorf("%q failed: %v", tt.ct, err)
		}
		if !tt.ok && err == nil {
			t.Errorf("%q was not rejected", tt.ct)
		}
	}
}

func TestParseBadVersion(t *testing.T) {
	r := strings.NewReader(`<request protocol="2.0"></request>`)
	err := parseReqOrResp(r, &Request{})
	if err == nil {
		t.Error("Bad protocol version was accepted")
	} else if err.Error() != `unsupported omaha protocol: "2.0"` {
		t.Errorf("Wrong error: %v", err)
	}
}
