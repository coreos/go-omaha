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

	"github.com/satori/go.uuid"
)

// Client supports managing multiple apps using a single server.
type Client struct {
	apiEndpoint   string
	clientVersion string
	userID        string
	sessionID     string
	isMachine     bool
	apps          map[string]*AppClient
}

// AppClient supports managing a single application.
type AppClient struct {
	*Client
	appID   string
	track   string
	version string
}

// New creates an omaha client for updating one or more applications.
// userID must be a persistent unique identifier of this update client.
func New(serverURL, userID string) (*Client, error) {
	if userID == "" {
		return nil, errors.New("omaha: empty user identifier")
	}

	c := &Client{
		clientVersion: "go-omaha",
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
