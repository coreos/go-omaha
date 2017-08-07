// Copyright 2017 CoreOS, Inc.
// Copyright 2011 The Chromium OS Authors.
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

	"github.com/coreos/go-omaha/omaha"
)

var (
	// These events are what update_engine sends to CoreUpdate to
	// mark different steps in the update process.
	EventDownloading = &omaha.EventRequest{
		Type:   omaha.EventTypeUpdateDownloadStarted,
		Result: omaha.EventResultSuccess,
	}
	EventDownloaded = &omaha.EventRequest{
		Type:   omaha.EventTypeUpdateDownloadFinished,
		Result: omaha.EventResultSuccess,
	}
	EventInstalled = &omaha.EventRequest{
		Type:   omaha.EventTypeUpdateComplete,
		Result: omaha.EventResultSuccess,
	}
	EventComplete = &omaha.EventRequest{
		Type:   omaha.EventTypeUpdateComplete,
		Result: omaha.EventResultSuccessReboot,
	}
)

// ExitCode is used for omaha event error codes derived from update_engine
type ExitCode int

// These error codes are from CoreOS Container Linux update_engine 0.4.x
// https://github.com/coreos/update_engine/blob/master/src/update_engine/action_processor.h
// The whole list is included for the sake of completeness but lots of these
// are not generally applicable and not even used by update_engine any more.
// Also there are clearly duplicate errors for the same condition.
const (
	ExitCodeSuccess                                    ExitCode = 0
	ExitCodeError                                      ExitCode = 1
	ExitCodeOmahaRequestError                          ExitCode = 2
	ExitCodeOmahaResponseHandlerError                  ExitCode = 3
	ExitCodeFilesystemCopierError                      ExitCode = 4
	ExitCodePostinstallRunnerError                     ExitCode = 5
	ExitCodeSetBootableFlagError                       ExitCode = 6
	ExitCodeInstallDeviceOpenError                     ExitCode = 7
	ExitCodeKernelDeviceOpenError                      ExitCode = 8
	ExitCodeDownloadTransferError                      ExitCode = 9
	ExitCodePayloadHashMismatchError                   ExitCode = 10
	ExitCodePayloadSizeMismatchError                   ExitCode = 11
	ExitCodeDownloadPayloadVerificationError           ExitCode = 12
	ExitCodeDownloadNewPartitionInfoError              ExitCode = 13
	ExitCodeDownloadWriteError                         ExitCode = 14
	ExitCodeNewRootfsVerificationError                 ExitCode = 15
	ExitCodeNewKernelVerificationError                 ExitCode = 16
	ExitCodeSignedDeltaPayloadExpectedError            ExitCode = 17
	ExitCodeDownloadPayloadPubKeyVerificationError     ExitCode = 18
	ExitCodePostinstallBootedFromFirmwareB             ExitCode = 19
	ExitCodeDownloadStateInitializationError           ExitCode = 20
	ExitCodeDownloadInvalidMetadataMagicString         ExitCode = 21
	ExitCodeDownloadSignatureMissingInManifest         ExitCode = 22
	ExitCodeDownloadManifestParseError                 ExitCode = 23
	ExitCodeDownloadMetadataSignatureError             ExitCode = 24
	ExitCodeDownloadMetadataSignatureVerificationError ExitCode = 25
	ExitCodeDownloadMetadataSignatureMismatch          ExitCode = 26
	ExitCodeDownloadOperationHashVerificationError     ExitCode = 27
	ExitCodeDownloadOperationExecutionError            ExitCode = 28
	ExitCodeDownloadOperationHashMismatch              ExitCode = 29
	ExitCodeOmahaRequestEmptyResponseError             ExitCode = 30
	ExitCodeOmahaRequestXMLParseError                  ExitCode = 31
	ExitCodeDownloadInvalidMetadataSize                ExitCode = 32
	ExitCodeDownloadInvalidMetadataSignature           ExitCode = 33
	ExitCodeOmahaResponseInvalid                       ExitCode = 34
	ExitCodeOmahaUpdateIgnoredPerPolicy                ExitCode = 35
	ExitCodeOmahaUpdateDeferredPerPolicy               ExitCode = 36
	ExitCodeOmahaErrorInHTTPResponse                   ExitCode = 37
	ExitCodeDownloadOperationHashMissingError          ExitCode = 38
	ExitCodeDownloadMetadataSignatureMissingError      ExitCode = 39
	ExitCodeOmahaUpdateDeferredForBackoff              ExitCode = 40
	ExitCodePostinstallPowerwashError                  ExitCode = 41
	ExitCodeNewPCRPolicyVerificationError              ExitCode = 42
	ExitCodeNewPCRPolicyHTTPError                      ExitCode = 43

	// Use the 2xxx range to encode HTTP errors from the Omaha server.
	// Sometimes aggregated into ExitCodeOmahaErrorInHTTPResponse
	ExitCodeOmahaRequestHTTPResponseBase ExitCode = 2000 // + HTTP response code
)

func (e ExitCode) String() string {
	switch e {
	case ExitCodeSuccess:
		return "success"
	case ExitCodeError:
		return "error"
	case ExitCodeOmahaRequestError:
		return "omaha request error"
	case ExitCodeOmahaResponseHandlerError:
		return "omaha response handler error"
	case ExitCodeFilesystemCopierError:
		return "filesystem copier error"
	case ExitCodePostinstallRunnerError:
		return "postinstall runner error"
	case ExitCodeSetBootableFlagError:
		return "set bootable flag error"
	case ExitCodeInstallDeviceOpenError:
		return "install device open error"
	case ExitCodeKernelDeviceOpenError:
		return "kernel device open error"
	case ExitCodeDownloadTransferError:
		return "download transfer error"
	case ExitCodePayloadHashMismatchError:
		return "payload hash mismatch error"
	case ExitCodePayloadSizeMismatchError:
		return "payload size mismatch error"
	case ExitCodeDownloadPayloadVerificationError:
		return "download payload verification error"
	case ExitCodeDownloadNewPartitionInfoError:
		return "download new partition info error"
	case ExitCodeDownloadWriteError:
		return "download write error"
	case ExitCodeNewRootfsVerificationError:
		return "new rootfs verification error"
	case ExitCodeNewKernelVerificationError:
		return "new kernel verification error"
	case ExitCodeSignedDeltaPayloadExpectedError:
		return "signed delta payload expected error"
	case ExitCodeDownloadPayloadPubKeyVerificationError:
		return "download payload pubkey verification error"
	case ExitCodePostinstallBootedFromFirmwareB:
		return "postinstall booted from firmware B"
	case ExitCodeDownloadStateInitializationError:
		return "download state initialization error"
	case ExitCodeDownloadInvalidMetadataMagicString:
		return "download invalid metadata magic string"
	case ExitCodeDownloadSignatureMissingInManifest:
		return "download signature missing in manifest"
	case ExitCodeDownloadManifestParseError:
		return "download manifest parse error"
	case ExitCodeDownloadMetadataSignatureError:
		return "download metadata signature error"
	case ExitCodeDownloadMetadataSignatureVerificationError:
		return "download metadata signature verification error"
	case ExitCodeDownloadMetadataSignatureMismatch:
		return "download metadata signature mismatch"
	case ExitCodeDownloadOperationHashVerificationError:
		return "download operation hash verification error"
	case ExitCodeDownloadOperationExecutionError:
		return "download operation execution error"
	case ExitCodeDownloadOperationHashMismatch:
		return "download operation hash mismatch"
	case ExitCodeOmahaRequestEmptyResponseError:
		return "omaha request empty response error"
	case ExitCodeOmahaRequestXMLParseError:
		return "omaha request XML parse error"
	case ExitCodeDownloadInvalidMetadataSize:
		return "download invalid metadata size"
	case ExitCodeDownloadInvalidMetadataSignature:
		return "download invalid metadata signature"
	case ExitCodeOmahaResponseInvalid:
		return "omaha response invalid"
	case ExitCodeOmahaUpdateIgnoredPerPolicy:
		return "omaha update ignored per policy"
	case ExitCodeOmahaUpdateDeferredPerPolicy:
		return "omaha update deferred per policy"
	case ExitCodeOmahaErrorInHTTPResponse:
		return "omaha error in HTTP response"
	case ExitCodeDownloadOperationHashMissingError:
		return "download operation hash missing error"
	case ExitCodeDownloadMetadataSignatureMissingError:
		return "download metadata signature missing error"
	case ExitCodeOmahaUpdateDeferredForBackoff:
		return "omaha update deferred for backoff"
	case ExitCodePostinstallPowerwashError:
		return "postinstall powerwash error"
	case ExitCodeNewPCRPolicyVerificationError:
		return "new PCR policy verification error"
	case ExitCodeNewPCRPolicyHTTPError:
		return "new PCR policy HTTP error"
	default:
		if e > ExitCodeOmahaRequestHTTPResponseBase {
			return fmt.Sprintf("omaha response HTTP %d error",
				e-ExitCodeOmahaRequestHTTPResponseBase)
		}
		return fmt.Sprintf("error code %d", e)
	}
}

// NewErrorEvent creates an EventRequest for reporting errors.
func NewErrorEvent(e ExitCode) *omaha.EventRequest {
	return &omaha.EventRequest{
		Type:      omaha.EventTypeUpdateComplete,
		Result:    omaha.EventResultError,
		ErrorCode: int(e),
	}
}

// EventString allows for easily logging events in a readable format.
func EventString(e *omaha.EventRequest) string {
	s := fmt.Sprintf("omaha event: %s: %s", e.Type, e.Result)
	if e.ErrorCode != 0 {
		s = fmt.Sprintf("%s (%d - %s)", s,
			e.ErrorCode, ExitCode(e.ErrorCode))
	}
	return s
}

// ErrorEvent is an error type that can generate EventRequests for reporting.
type ErrorEvent interface {
	error
	ErrorEvent() *omaha.EventRequest
}
