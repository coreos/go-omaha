package omaha

import (
	"testing"
	"fmt"
	"encoding/xml"
	"io/ioutil"
	"os"
)

func TestOmahaRequestUpdateCheck(t *testing.T) {
	file, err := os.Open("fixtures/update-engine/update/request.xml")
	if err != nil {
		t.Error(err)
	}
	fix, err := ioutil.ReadAll(file)
	if err != nil {
		t.Error(err)
	}
	v := Request{}
	xml.Unmarshal(fix, &v)

	if v.Os.Version != "Indy" {
		t.Error("Unexpected version", v.Os.Version)
	}

	if v.Apps[0].Id != "{87efface-864d-49a5-9bb3-4b050a7c227a}" {
		t.Error("Expected an App Id")
	}

	if v.Apps[0].UpdateCheck == nil {
		t.Error("Expected an UpdateCheck")
	}

	if v.Apps[0].Version != "ForcedUpdate" {
		t.Error("Verison is ForcedUpdate")
	}

	if v.Apps[0].FromTrack != "developer-build" {
		t.Error("developer-build")
	}

	if v.Apps[0].Event.Type != "3" {
		t.Error("developer-build")
	}
}

func ExampleOmaha_NewRequest() {
	os := NewOs("linux", "3.0", "", "x64")

	app := NewApp("{27BD862E-8AE8-4886-A055-F7F1A6460627}", "1.0.0.0")
	app.AddUpdateCheck()

	request := NewRequest(os, app)

	if raw, err := xml.MarshalIndent(request, "", " "); err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Printf("%s%s\n", xml.Header, raw)
	}

	// Output:
	// <?xml version="1.0" encoding="UTF-8"?>
	//
	// <Request protocol="3.0">
	//  <os platform="linux" version="3.0" arch="x64"></os>
	//  <app appid="{27BD862E-8AE8-4886-A055-F7F1A6460627}" version="1.0.0.0">
	//   <updatecheck></updatecheck>
	//  </app>
	// </Request>
}
