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
   Each check in should include a ping and optionally an update check.

 - Various protocol extensions/abuses.
   update_engine, likely due to earlier limitations of the protocol and Google's server implementation, uses a number of non-standard fields.
   For example, packing a lot of extra attributes such as the package's SHA-256 hash into a "postinstall" action.
   As much as possible the code includes comments about these extensions.

 - Many data fields not used by CoreUpdate are omitted.
