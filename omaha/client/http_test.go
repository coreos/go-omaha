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
	"bytes"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/coreos/go-omaha/omaha"
)

const (
	sampleRequest = `<?xml version="1.0" encoding="UTF-8"?>
<request protocol="3.0" version="ChromeOSUpdateEngine-0.1.0.0" updaterversion="ChromeOSUpdateEngine-0.1.0.0" installsource="ondemandupdate" ismachine="1">
<os version="Indy" platform="Chrome OS" sp="ForcedUpdate_x86_64"></os>
<app appid="{87efface-864d-49a5-9bb3-4b050a7c227a}" bootid="{7D52A1CC-7066-40F0-91C7-7CB6A871BFDE}" machineid="{8BDE4C4D-9083-4D61-B41C-3253212C0C37}" oem="ec3000" version="ForcedUpdate" track="dev-channel" from_track="developer-build" lang="en-US" board="amd64-generic" hardware_class="" delta_okay="false" >
<ping active="1" a="-1" r="-1"></ping>
<updatecheck targetversionprefix=""></updatecheck>
<event eventtype="3" eventresult="2" previousversion=""></event>
</app>
</request>
`
)

func TestHTTPClientDoPost(t *testing.T) {
	s, err := omaha.NewTrivialServer("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Destroy()
	go s.Serve()

	c := newHTTPClient()
	url := "http://" + s.Addr().String() + "/v1/update/"

	resp, err := c.doPost(url, []byte(sampleRequest))
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Apps) != 1 {
		t.Fatalf("Should be 1 app, not %d", len(resp.Apps))
	}
	if resp.Apps[0].Status != omaha.AppOK {
		t.Fatalf("Bad apps status: %q", resp.Apps[0].Status)
	}
}

type flakyHandler struct {
	omaha.OmahaHandler
	flakes int
	reqs   int
}

func (f *flakyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.reqs++
	if f.flakes > 0 {
		f.flakes--
		http.Error(w, "Flake!", http.StatusInternalServerError)
		return
	}
	f.OmahaHandler.ServeHTTP(w, r)
}

type flakyServer struct {
	l net.Listener
	s *http.Server
	h *flakyHandler
}

func newFlakyServer() (*flakyServer, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	f := &flakyServer{
		l: l,
		s: &http.Server{},
		h: &flakyHandler{
			OmahaHandler: omaha.OmahaHandler{
				Updater: omaha.UpdaterStub{},
			},
			flakes: 1,
		},
	}
	f.s.Handler = f.h

	go f.s.Serve(l)
	return f, nil
}

func TestHTTPClientError(t *testing.T) {
	f, err := newFlakyServer()
	if err != nil {
		t.Fatal(err)
	}
	defer f.l.Close()

	c := newHTTPClient()
	url := "http://" + f.l.Addr().String()

	_, err = c.doPost(url, []byte(sampleRequest))
	switch err := err.(type) {
	case nil:
		t.Fatal("doPost succeeded but should have failed")
	case *httpError:
		if err.StatusCode != http.StatusInternalServerError {
			t.Fatalf("Unexpected http error: %v", err)
		}
		if err.Timeout() {
			t.Fatal("http 500 error reported as timeout")
		}
		if !err.Temporary() {
			t.Fatal("http 500 error not reported as temporary")
		}
	default:
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestHTTPClientRetry(t *testing.T) {
	f, err := newFlakyServer()
	if err != nil {
		t.Fatal(err)
	}
	defer f.l.Close()

	req, err := omaha.ParseRequest("", strings.NewReader(sampleRequest))
	if err != nil {
		t.Fatal(err)
	}

	c := newHTTPClient()
	url := "http://" + f.l.Addr().String()

	resp, err := c.Omaha(url, req)
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Apps) != 1 {
		t.Fatalf("Should be 1 app, not %d", len(resp.Apps))
	}

	if resp.Apps[0].Status != omaha.AppOK {
		t.Fatalf("Bad apps status: %q", resp.Apps[0].Status)
	}

	if f.h.reqs != 2 {
		t.Fatalf("Server received %d requests, not 2", f.h.reqs)
	}
}

// should result in an unexected EOF
func largeHandler1(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><response protocol="3.0">`))
	w.Write(bytes.Repeat([]byte{' '}, 2*1024*1024))
	w.Write([]byte(`</response>`))
}

// should result in an EOF
func largeHandler2(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.Write(bytes.Repeat([]byte{' '}, 2*1024*1024))
}

func TestHTTPClientLarge(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	s := &http.Server{
		Handler: http.HandlerFunc(largeHandler1),
	}
	go s.Serve(l)

	c := newHTTPClient()
	url := "http://" + l.Addr().String()

	_, err = c.doPost(url, []byte(sampleRequest))
	if err != bodySizeError {
		t.Errorf("Unexpected error: %v", err)
	}

	// switch to failing before XML is read instead of half-way
	// through (which results in a different error internally)
	s.Handler = http.HandlerFunc(largeHandler2)

	_, err = c.doPost(url, []byte(sampleRequest))
	if err != bodyEmptyError {
		t.Errorf("Unexpected error: %v", err)
	}
}
