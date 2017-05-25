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
	"io/ioutil"
	"testing"
)

// skip test if external file isn't readable
func readOrSkip(t *testing.T, name string) string {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		t.Skip(err)
	}
	return string(bytes.TrimSpace(data))
}

func TestNewMachine(t *testing.T) {
	userID := readOrSkip(t, machineIDPath)
	sessionID := readOrSkip(t, bootIDPath)

	c, err := NewMachineClient("https://example.com")
	if err != nil {
		t.Fatal(err)
	}
	if c.userID != userID {
		t.Errorf("%q != %q", c.userID, userID)
	}
	if c.sessionID != sessionID {
		t.Errorf("%q != %q", c.sessionID, sessionID)
	}
}
