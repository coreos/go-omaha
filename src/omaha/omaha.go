/*
    Package that implements the Google omaha protocol.

    Omaha is a request/response protocol using XML. Requests are made by
    clients and responses are given by the Omaha server.
    http://code.google.com/p/omaha/wiki/ServerProtocol
    The 
*/
package omaha

import "encoding/xml"

type Request struct {
	Os Os
	Apps []App `xml:"app"`
	Protocol string `xml:"protocol,attr"`
	Version string `xml:"version,attr,omitempty"`
	IsMachine string `xml:"ismachine,attr,omitempty"`
	SessionId string `xml:"sessionid,attr,omitempty"`
	UserId string `xml:"userid,attr,omitempty"`
	InstallSource string `xml:"installsource,attr,omitempty"`
	TestSource string `xml:"testsource,attr,omitempty"`
	RequestId string `xml:"requestid,attr,omitempty"`
	UpdaterVersion string `xml:"updaterversion,attr,omitempty"`
}

func NewRequest(os *Os, app *App) *Request {
	r := new(Request)
	r.Protocol = "3.0"
	r.AddApp(app)
	r.Os = *os
	return r
}

func (r *Request) AddApp(a *App) {
	r.Apps = append(r.Apps, *a)
}

// app element
type App struct {
	XMLName xml.Name `xml:"app"`
	UpdateCheck *UpdateCheck `xml:"updatecheck"`
	Event *Event `xml:"event"`
	Ping *Ping `xml:"ping"`
	Id string `xml:"appid,attr,omitempty"`
	Version string `xml:"version,attr,omitempty"`
	NextVersion string `xml:"nextversion,attr,omitempty"`
	Lang string `xml:"lang,attr,omitempty"`
	Client string `xml:"client,attr,omitempty"`
	InstallAge string `xml:"installage,attr,omitempty"`
	FromTrack string `xml:"from_track,attr,omitempty"`
}

func NewApp(id string, version string) *App {
	a := new(App)
	a.Id = id
	a.Version = version
	return a
}

type UpdateCheck struct {
	XMLName xml.Name `xml:"updatecheck"`
	TargetVersionPrefix string `xml:"targetversionprefix,attr,omitempty"`
}

func (a *App) AddUpdateCheck() {
	a.UpdateCheck = new(UpdateCheck)
}

type Ping struct {
	XMLName xml.Name `xml:"ping"`
	LastReportDays string `xml:"r,attr,omitempty"`
}

type Os struct {
	XMLName xml.Name `xml:"os"`
	Platform string `xml:"platform,attr,omitempty"`
	Version string `xml:"version,attr,omitempty"`
	Sp string `xml:"sp,attr,omitempty"`
	Arch string `xml:"arch,attr,omitempty"`
}

func NewOs(platform string, version string, sp string, arch string) *Os {
	o := new(Os)
	o.Platform = platform
	o.Version = version
	o.Sp = sp
	o.Arch = arch
	return o
}

func (a *App) AddPing() {
}

type Event struct {
	XMLName xml.Name `xml:"event"`
	Type string `xml:"eventtype,attr,omitempty"`
	Result string `xml:"eventresult,attr,omitempty"`
	PreviousVersion string `xml:"previousversion,attr,omitempty"`
}

var EventTypes = map[int] string {
	0: "unknown",
	1: "download complete",
	2: "install complete",
	3: "update complete",
	4: "uninstall",
	5: "download started",
	6: "install started",
	9: "new application install started",
	10: "setup started",
	11: "setup finished",
	12: "update application started",
	13: "update download started",
	14: "update download finished",
	15: "update installer started",
	16: "setup update begin",
	17: "setup update complete",
	20: "register product complete",
	30: "OEM install first check",
	40: "app-specific command started",
	41: "app-specific command ended",
	100: "setup failure",
	102: "COM server failure",
	103: "setup update failure",
}

var EventResults = map[int] string {
	0: "error",
	1: "success",
	2: "success reboot",
	3: "success restart browser",
	4: "cancelled",
	5: "error installer MSI",
	6: "error installer other",
	7: "noupdate",
	8: "error installer system",
	9: "update deferred",
	10: "handoff error",
}
