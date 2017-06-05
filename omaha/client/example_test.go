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
	"fmt"
	"os"
	//"os/signal"
	"syscall"

	"github.com/coreos/go-omaha/omaha"
)

func Example() {
	// Launch a dummy server for our client to talk to.
	s, err := omaha.NewTrivialServer("127.0.0.1:0")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer s.Destroy()
	go s.Serve()

	// Configure our client. userID should be random but preserved
	// across restarts. version is the current version of our app.
	var (
		serverURL = "http://" + s.Addr().String()
		userID    = "8b10fc6d-30ca-49b2-b1a2-8185f03d522b"
		appID     = "5ca607f8-61b5-4692-90ce-30380ba05a98"
		version   = "1.0.0"
	)
	c, err := NewAppClient(serverURL, userID, appID, version)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Client version is the name and version of this updater.
	c.SetClientVersion("example-0.0.1")

	// Use SIGUSR1 to trigger immediate update checks.
	sigc := make(chan os.Signal, 1)
	//signal.Notify(sigc, syscall.SIGUSR1)
	sigc <- syscall.SIGUSR1 // Fake it

	//for {
	var source string
	select {
	case <-sigc:
		source = "ondemandupdate"
	case <-c.NextPing():
		source = "scheduler"
	}

	// TODO: pass source to UpdateCheck
	_ = source
	// If updates are disabled call c.Ping() instead.
	update, err := c.UpdateCheck()
	if err != nil {
		fmt.Println(err)
		//continue
		return
	}

	// Download new application version.
	c.Event(&omaha.EventRequest{
		Type:   omaha.EventTypeUpdateDownloadFinished,
		Result: omaha.EventResultSuccess,
	})

	// Install new application version here.
	c.Event(&omaha.EventRequest{
		Type:   omaha.EventTypeUpdateComplete,
		Result: omaha.EventResultSuccess,
	})

	// Restart, new application is now running.
	c.SetVersion(update.Manifest.Version)
	c.Event(&omaha.EventRequest{
		Type:   omaha.EventTypeUpdateComplete,
		Result: omaha.EventResultSuccessReboot,
	})

	//}

	// Output:
	// omaha: update status noupdate
}
