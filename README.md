# Go Omaha

[![Build Status](https://travis-ci.org/coreos/go-omaha.svg?branch=master)](https://travis-ci.org/coreos/go-omaha)
[![GoDoc](https://godoc.org/github.com/coreos/go-omaha/omaha?status.svg)](https://godoc.org/github.com/coreos/go-omaha/omaha)

Implementation of the [omaha update protocol](https://github.com/google/omaha) in Go.

## Status

This code is targeted for use with CoreOS's [CoreUpdate](https://coreos.com/products/coreupdate/) product and the Container Linux [update_engine](https://github.com/coreos/update_engine).
As a result this is not a complete implementation of the [protocol](https://github.com/google/omaha/blob/master/doc/ServerProtocolV3.md) and inherits a number of quirks from update_engine.
These differences include:

 - No offline activity tracking.
   The protocol's ping mechanism allows for tracking application usage, reporting the number of days since the last ping and how many of those days saw active usage.
   CoreUpdate does not use this, instead assuming update clients are always online and checking in once every ~45-50 minutes.
   Clients not actively updating should send only a ping, indicating CoreUpdate's "Instance-Hold" state.
   Clients requesting an update should send a ping, update check, and an UpdateComplete:SuccessReboot event indicating CoreUpdate's "Complete" state.

 - Various protocol extensions/abuses.
   update_engine, likely due to earlier limitations of the protocol and Google's server implementation, uses a number of non-standard fields.
   For example, packing a lot of extra attributes such as the package's SHA-256 hash into a "postinstall" action.
   As much as possible the code includes comments about these extensions.

 - Many data fields not used by CoreUpdate are omitted.
 
## `serve-package`

This project includes a very simple program designed to serve a single Container Linux package on the local host. It is intended to be used as a manual updater for a machine that is not able to use a full-fledged CoreUpdate instance. Binaries are available for each released version on the [releases page](https://github.com/coreos/go-omaha/releases). `serve-package` can also be built from source using the provided Makefile:

```bash
make
```

The binary will be available in the `bin/` folder.

It is recommended that the server be run directly on the machine you intend to update. Go to the [Container Linux release notes](https://coreos.com/releases/) and find the version number for the release you would like to update to. The update payload can be retrieved from

```
https://update.release.core-os.net/amd64-usr/<version>/update.gz
```

where `<version>` is the version number you retrieved from the releases page. For example, `https://update.release.core-os.net/amd64-usr/1576.4.0/update.gz` is the payload required to update to Container Linux version 1576.4.0.

Copy the update payload and the `serve-package` binary to the server you would like to update. `serve-package` can be run as follows:

```bash
./serve-package --package-file update.gz --package-version <version>
```

By default, the server listens on `localhost:8000`. This can be modified using the `--listen-address` option.

Next, `update_engine` needs to be configured to use the local server that was just set up:

```bash
echo "SERVER=http://localhost:8000/v1/update" | sudo tee -a /etc/coreos/update.conf
```

Restart `update_engine` and tell it to check for an update:

```bash
sudo systemctl restart update-engine.service
update_engine_client -check_for_update
```

If `locksmithd.service` is running, the machine will restart once it has updated to the latest version. Otherwise, watch the logs from `update-engine.service` to determine when the update is complete and the machine is ready to restart:

```bash
journalctl -u update-engine.service -f
# wait for a line that says "Update successfully applied, waiting for reboot"
sudo systemctl reboot
```
