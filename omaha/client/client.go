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

// Package client provides a general purpose Omaha update client implementation.
package client

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/satori/go.uuid"

	"github.com/coreos/go-omaha/omaha"
)

const (
	defaultClientVersion = "go-omaha"

	// periodic update check and ping intervals
	pingFuzz     = 10 * time.Minute
	pingDelay    = 7 * time.Minute  // first check after 2-12 minutes
	pingInterval = 45 * time.Minute // check in every 40-50 minutes
)

// Client supports managing multiple apps using a single server.
type Client struct {
	apiClient     *httpClient
	apiEndpoint   string
	clientVersion string
	userID        string
	sessionID     string
	isMachine     bool
	sentPing      bool
	apps          map[string]*AppClient
}

// AppClient supports managing a single application.
type AppClient struct {
	*Client
	appID   string
	track   string
	version string
	oem     string
}

// New creates an omaha client for updating one or more applications.
// userID must be a persistent unique identifier of this update client.
func New(serverURL, userID string) (*Client, error) {
	if userID == "" {
		return nil, errors.New("omaha: empty user identifier")
	}

	c := &Client{
		apiClient:     newHTTPClient(),
		clientVersion: defaultClientVersion,
		userID:        userID,
		sessionID:     uuid.NewV4().String(),
		apps:          make(map[string]*AppClient),
	}

	if err := c.SetServerURL(serverURL); err != nil {
		return nil, err
	}

	return c, nil
}

// SetServerURL changes the Omaha server this client talks to.
// If the URL does not include a path component /v1/update/ is assumed.
func (c *Client) SetServerURL(serverURL string) error {
	u, err := url.Parse(serverURL)
	if err != nil {
		return fmt.Errorf("omaha: invalid server URL: %v", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("omaha: invalid server protocol: %s", u)
	}
	if u.Host == "" {
		return fmt.Errorf("omaha: invalid server host: %s", u)
	}
	if u.Path == "" || u.Path == "/" {
		u.Path = "/v1/update/"
	}

	c.apiEndpoint = u.String()
	return nil
}

// SetClientVersion sets the identifier of this updater application.
// e.g. "update_engine-0.1.0".  Default is "go-omaha".
func (c *Client) SetClientVersion(clientVersion string) {
	c.clientVersion = clientVersion
}

// NextPing returns a timer channel that will fire when the next update
// check or ping should be sent.
func (c *Client) NextPing() <-chan time.Time {
	d := pingDelay
	if c.sentPing {
		d = pingInterval
	}
	return FuzzyAfter(d, pingFuzz)
}

// AppClient gets the application client for the given application ID.
func (c *Client) AppClient(appID string) (*AppClient, error) {
	if app, ok := c.apps[appID]; ok {
		return app, nil
	}

	return nil, fmt.Errorf("omaha: missing app client %q", appID)
}

// NewAppClient creates a new application client.
func (c *Client) NewAppClient(appID, appVersion string) (*AppClient, error) {
	if _, ok := c.apps[appID]; ok {
		return nil, fmt.Errorf("omaha: duplicate app client %q", appID)
	}

	ac := &AppClient{
		Client: c,
		appID:  appID,
	}
	c.apps[appID] = ac

	return ac, nil
}

// NewAppClient creates a single application client.
// Shorthand for New(serverURL, userID).NewAppClient(appID, appVersion).
func NewAppClient(serverURL, userID, appID, appVersion string) (*AppClient, error) {
	c, err := New(serverURL, userID)
	if err != nil {
		return nil, err
	}

	ac, err := c.NewAppClient(appID, appVersion)
	if err := ac.SetVersion(appVersion); err != nil {
		return nil, err
	}

	return ac, nil
}

func (ac *AppClient) SetAppID(appID string) error {
	if appID == ac.appID {
		return nil
	}

	if _, ok := ac.apps[appID]; ok {
		return fmt.Errorf("omaha: duplicate app %q", appID)
	}

	delete(ac.apps, ac.appID)
	ac.appID = appID
	ac.apps[appID] = ac
	return nil
}

// SetVersion changes the application version.
func (ac *AppClient) SetVersion(version string) error {
	if version == "" {
		return errors.New("omaha: empty application version")
	}

	ac.version = version
	return nil
}

// SetTrack sets the application update track or group.
// This is a update_engine/Core Update protocol extension.
func (ac *AppClient) SetTrack(track string) error {
	// Although track is an omaha extension and theoretically not required
	// our Core Update server requires track to be set to a valid id/name.
	// TODO: deprecate track and use the standard cohort protocol fields.
	if track == "" {
		return errors.New("omaha: empty application update track/group")
	}

	ac.track = track
	return nil
}

// SetOEM sets the application OEM name.
// This is a update_engine/Core Update protocol extension.
func (ac *AppClient) SetOEM(oem string) {
	ac.oem = oem
}

func (ac *AppClient) UpdateCheck() (*omaha.UpdateResponse, error) {
	req := ac.newReq()
	app := req.Apps[0]
	app.AddPing()
	app.AddUpdateCheck()

	ac.sentPing = true

	appResp, err := ac.doReq(ac.apiEndpoint, req)
	if err, ok := err.(ErrorEvent); ok {
		ac.Event(err.ErrorEvent())
		return nil, err
	} else if err != nil {
		ac.Event(NewErrorEvent(ExitCodeOmahaRequestError))
		return nil, err
	}

	if appResp.Ping == nil {
		ac.Event(NewErrorEvent(ExitCodeOmahaResponseInvalid))
		return nil, fmt.Errorf("omaha: ping status missing from response")
	}

	if appResp.Ping.Status != "ok" {
		return nil, fmt.Errorf("omaha: ping status %s", appResp.Ping.Status)
	}

	if appResp.UpdateCheck == nil {
		ac.Event(NewErrorEvent(ExitCodeOmahaResponseInvalid))
		return nil, fmt.Errorf("omaha: update check missing from response")
	}

	if appResp.UpdateCheck.Status != omaha.UpdateOK {
		return nil, appResp.UpdateCheck.Status
	}

	return appResp.UpdateCheck, nil
}

func (ac *AppClient) Ping() error {
	req := ac.newReq()
	app := req.Apps[0]
	app.AddPing()

	ac.sentPing = true

	appResp, err := ac.doReq(ac.apiEndpoint, req)
	if err, ok := err.(ErrorEvent); ok {
		ac.Event(err.ErrorEvent())
		return err
	} else if err != nil {
		ac.Event(NewErrorEvent(ExitCodeOmahaRequestError))
		return err
	}

	if appResp.Ping == nil {
		ac.Event(NewErrorEvent(ExitCodeOmahaResponseInvalid))
		return fmt.Errorf("omaha: ping status missing from response")
	}

	if appResp.Ping.Status != "ok" {
		return fmt.Errorf("omaha: ping status %s", appResp.Ping.Status)
	}

	return nil
}

// Event asynchronously sends the given omaha event.
// Reading the error channel is optional.
func (ac *AppClient) Event(event *omaha.EventRequest) <-chan error {
	errc := make(chan error, 1)
	url := ac.apiEndpoint
	req := ac.newReq()
	app := req.Apps[0]
	app.Events = append(app.Events, event)

	go func() {
		appResp, err := ac.doReq(url, req)
		if err != nil {
			errc <- err
			return
		}

		if len(appResp.Events) == 0 {
			errc <- fmt.Errorf("omaha: event status missing from response")
			return
		}

		if appResp.Events[0].Status != "ok" {
			errc <- fmt.Errorf("omaha: event status %s", appResp.Events[0].Status)
			return
		}

		errc <- nil
		return
	}()

	return errc
}

func (ac *AppClient) newReq() *omaha.Request {
	req := omaha.NewRequest()
	req.Version = ac.clientVersion
	req.UserID = ac.userID
	req.SessionID = ac.sessionID
	if ac.isMachine {
		req.IsMachine = 1
	}

	app := req.AddApp(ac.appID, ac.version)
	app.Track = ac.track
	app.OEM = ac.oem

	// MachineID and BootID are non-standard fields used by CoreOS'
	// update_engine and Core Update. Copy their values from the
	// standard UserID and SessionID. Eventually the non-standard
	// fields should be deprecated.
	app.MachineID = req.UserID
	app.BootID = req.SessionID

	return req
}

// doReq posts an omaha request. It may be called in its own goroutine so
// it should not touch any mutable data in AppClient, but apiClient is ok.
func (ac *AppClient) doReq(url string, req *omaha.Request) (*omaha.AppResponse, error) {
	if len(req.Apps) != 1 {
		panic(fmt.Errorf("unexpected number of apps: %d", len(req.Apps)))
	}
	appID := req.Apps[0].ID
	resp, err := ac.apiClient.Omaha(url, req)
	if err != nil {
		return nil, err
	}

	appResp := resp.GetApp(appID)
	if appResp == nil {
		return nil, &omahaError{
			Err:  fmt.Errorf("app %s missing from response", appID),
			Code: ExitCodeOmahaResponseInvalid,
		}
	}

	if appResp.Status != omaha.AppOK {
		return nil, appResp.Status
	}

	return appResp, nil
}
