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
	"reflect"
	"testing"

	"github.com/coreos/go-omaha/omaha"
)

// implements omaha.Updater
type recorder struct {
	t      *testing.T
	update *omaha.Update
	checks []*omaha.UpdateRequest
	events []*omaha.EventRequest
	pings  []*omaha.PingRequest
}

func newRecordingServer(t *testing.T, u *omaha.Update) (*recorder, *omaha.Server) {
	r := &recorder{t: t, update: u}
	s, err := omaha.NewServer("127.0.0.1:0", r)
	if err != nil {
		t.Fatal(err)
	}
	go s.Serve()
	return r, s
}

func (r *recorder) CheckApp(req *omaha.Request, app *omaha.AppRequest) error {
	// CheckApp is meant for checking if app.ID is valid but we don't
	// care and accept any ID. Instead this is just a convenient place
	// to check that all requests are well formed.
	if len(req.SessionID) != 36 {
		r.t.Errorf("SessionID %q is not a UUID", req.SessionID)
	}
	if app.BootID != req.SessionID {
		r.t.Errorf("BootID %q != SessionID %q", app.BootID, req.SessionID)
	}
	if req.UserID == "" {
		r.t.Error("UserID is blank")
	}
	if app.MachineID != req.UserID {
		r.t.Errorf("MachineID %q != UserID %q", app.MachineID, req.UserID)
	}
	if app.Version == "" {
		r.t.Error("App Version is blank")
	}
	return nil
}

func (r *recorder) CheckUpdate(req *omaha.Request, app *omaha.AppRequest) (*omaha.Update, error) {
	r.checks = append(r.checks, app.UpdateCheck)
	if r.update == nil {
		return nil, omaha.NoUpdate
	} else {
		return r.update, nil
	}
}

func (r *recorder) Event(req *omaha.Request, app *omaha.AppRequest, event *omaha.EventRequest) {
	r.events = append(r.events, event)
}

func (r *recorder) Ping(req *omaha.Request, app *omaha.AppRequest) {
	r.pings = append(r.pings, app.Ping)
}

func TestClientNoUpdate(t *testing.T) {
	r, s := newRecordingServer(t, nil)
	defer s.Destroy()

	url := "http://" + s.Addr().String()
	ac, err := NewAppClient(url, "client-id", "app-id", "0.0.0")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := ac.UpdateCheck(); err != omaha.NoUpdate {
		t.Fatalf("UpdateCheck id not return NoUpdate: %v", err)
	}

	if len(r.pings) != 1 {
		t.Fatalf("expected 1 ping, not %d", len(r.pings))
	}

	if len(r.checks) != 1 {
		t.Fatalf("expected 1 update check, not %d", len(r.checks))
	}
}

func TestClientWithUpdate(t *testing.T) {
	r, s := newRecordingServer(t, &omaha.Update{
		Manifest: omaha.Manifest{
			Version: "1.1.1",
		},
	})
	defer s.Destroy()

	url := "http://" + s.Addr().String()
	ac, err := NewAppClient(url, "client-id", "app-id", "0.0.0")
	if err != nil {
		t.Fatal(err)
	}

	update, err := ac.UpdateCheck()
	if err != nil {
		t.Fatal(err)
	}

	if update.Manifest.Version != "1.1.1" {
		t.Fatalf("expected version 1.1.1, not %s", update.Manifest.Version)
	}

	if len(r.pings) != 1 {
		t.Fatalf("expected 1 ping, not %d", len(r.pings))
	}

	if len(r.checks) != 1 {
		t.Fatalf("expected 1 update check, not %d", len(r.checks))
	}
}

func TestClientPing(t *testing.T) {
	r, s := newRecordingServer(t, nil)
	defer s.Destroy()

	url := "http://" + s.Addr().String()
	ac, err := NewAppClient(url, "client-id", "app-id", "0.0.0")
	if err != nil {
		t.Fatal(err)
	}

	if err := ac.Ping(); err != nil {
		t.Fatal(err)
	}

	if len(r.pings) != 1 {
		t.Fatalf("expected 1 ping, not %d", len(r.pings))
	}
}

func TestClientEvent(t *testing.T) {
	r, s := newRecordingServer(t, nil)
	defer s.Destroy()

	url := "http://" + s.Addr().String()
	ac, err := NewAppClient(url, "client-id", "app-id", "0.0.0")
	if err != nil {
		t.Fatal(err)
	}

	event := &omaha.EventRequest{
		Type:   omaha.EventTypeDownloadComplete,
		Result: omaha.EventResultSuccess,
	}
	if err := ac.Event(event); err != nil {
		t.Fatal(err)
	}

	if len(r.events) != 1 {
		t.Fatalf("expected 1 event, not %d", len(r.events))
	}

	if !reflect.DeepEqual(event, r.events[0]) {
		t.Fatalf("sent != received:\n%#v\n%#v", event, r.events[0])
	}
}
