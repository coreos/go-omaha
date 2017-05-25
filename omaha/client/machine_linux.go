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

// +build linux

package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
)

const (
	machineIDPath = "/etc/machine-id"
	bootIDPath    = "/proc/sys/kernel/random/boot_id"
)

// NewMachineClient creates a machine-wide client, updating applications
// that may be used by multiple users. On Linux the system's machine id
// is used as the user id, and boot id is used as the omaha session id.
func NewMachineClient(serverURL string) (*Client, error) {
	machineID, err := ioutil.ReadFile(machineIDPath)
	if err != nil {
		fmt.Errorf("omaha: failed to read machine id: %v", err)
	}

	machineID = bytes.TrimSpace(machineID)
	// Although machineID should be a UUID, it is formatted as a
	// plain hex string, omitting the normal '-' separators, so it
	// should be 32 bytes long. It would be nice to reformat it to
	// add the '-' chars but update_engine doesn't so stick with its
	// behavior for now.
	if len(machineID) < 32 {
		fmt.Errorf("omaha: incomplete machine id: %q",
			machineID)
	}

	bootID, err := ioutil.ReadFile(bootIDPath)
	if err != nil {
		fmt.Errorf("omaha: failed to read boot id: %v", err)
	}

	bootID = bytes.TrimSpace(bootID)
	// unlike machineID, bootID *does* include '-' chars.
	if len(bootID) < 36 {
		fmt.Errorf("omaha: incomplete boot id: %q", bootID)
	}

	c := &Client{
		apiClient:     newHTTPClient(),
		clientVersion: "go-omaha",
		userID:        string(machineID),
		sessionID:     string(bootID),
		isMachine:     true,
		apps:          make(map[string]*AppClient),
	}

	if err := c.SetServerURL(serverURL); err != nil {
		return nil, err
	}

	return c, nil
}
