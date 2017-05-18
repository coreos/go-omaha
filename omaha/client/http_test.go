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
