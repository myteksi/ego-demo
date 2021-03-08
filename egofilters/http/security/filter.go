// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package security

import (
	"encoding/json"
	"io"

	ego "github.com/grab/ego/ego/src/go"
	"github.com/grab/ego/ego/src/go/envoy"
	"github.com/grab/ego/ego/src/go/envoy/datastatus"
	"github.com/grab/ego/ego/src/go/envoy/headersstatus"
	"github.com/grab/ego/ego/src/go/envoy/lifespan"
	"github.com/grab/ego/ego/src/go/envoy/statetype"
	"github.com/grab/ego/ego/src/go/envoy/trailersstatus"
	"github.com/grab/ego/ego/src/go/logger"

	"github.com/grab/ego/egofilters/http/security/context"
	pb "github.com/grab/ego/egofilters/http/security/proto"
	"github.com/grab/ego/egofilters/http/security/verifier"
)

const (
	FilterID = "security"
)

// State of this filter's communication with the verifier.
// The filter has either not started calling the verifier, in the middle of calling it
// or has completed.
type State int

const (
	NotStarted State = iota
	// when the filter is waiting for request body
	WaitingForRequestBody
	// When the filter has received a request & start verifying
	Calling
	// When the filter has sent a local reply to client
	Responded
	// When the filter allows a request coming to next filter
	Complete
	// when the filter is waiting for response body
	WaitingForResponseBody
	// when the filter is signing response
	Signing
)

const (
	authPost uint64 = 0
	signPost uint64 = 1
)

type security struct {
	ego.HttpFilterBase

	config          *securityConfig
	state           State
	requestHeaders  envoy.RequestHeaderMap
	responseHeaders envoy.ResponseHeaderMap

	verifier verifier.Verifier
	signer   verifier.Signer

	// Used to caching response from OnComplete from a goroutine
	authResponse context.AuthResponse

	// Used to cache sign response
	signResponse context.SignResponse

	// Secret key/value pairs that we automagically get injected
	secrets map[string]string
}

func newSecurity(native envoy.GoHttpFilter, config *securityConfig) ego.HttpFilter {
	f := &security{
		state:  NotStarted,
		config: config,
	}
	f.HttpFilterBase.Init(native)

	// Secrets already held by C, no need to copy
	unsafeBytes := []byte(f.Native.GenericSecretProvider().Secret())
	if err := json.Unmarshal(unsafeBytes, &f.secrets); err != nil {
		f.Logger().Error("[newSecurity] unmarshal secret error")
	}
	return f
}

// Noted that we need to handle OnDestroy to stop whatever we're doing in securityFilter
func (f *security) OnDestroy() {
	f.Cancel()
}

func (f *security) DecodeHeaders(headers envoy.RequestHeaderMap, endStream bool) headersstatus.Type {

	f.Logger().Debug("[DecodeHeaders] called")

	config := f.Native.ResolveMostSpecificPerGoFilterConfig(FilterID, f.Native.DecoderCallbacks().Route())
	if config == nil {
		return headersstatus.Continue
	}

	requirement, ok := config.(pb.Requirement)
	if !ok {
		return headersstatus.Continue
	}

	f.verifier, f.signer = f.config.findProvider(&requirement)
	if f.verifier == nil {
		// FIXME: shouldn't we rather block the request?
		return headersstatus.Continue
	}

	f.requestHeaders = headers

	// TODO: check logic for Http::Utility::isWebSocketUpgradeRequest(headers)
	// and Http::Utility::isH2UpgradeRequest(headers)
	if f.verifier.WithBody() && !endStream {
		// Need to wait for body on DecodeData
		f.state = WaitingForRequestBody
		return headersstatus.StopIteration
	}

	// Otherwise startVerify the request without body
	f.Logger().Debug("[DecodeHeaders] start verify without body")
	f.startVerify(nil)

	// Waiting for the authResp from the verifier
	return headersstatus.StopAllIterationAndWatermark
}

func (f *security) DecodeData(data envoy.BufferInstance, endStream bool) datastatus.Type {

	f.Logger().Debug("[DecodeData] called")

	if f.state != WaitingForRequestBody {
		return datastatus.Continue
	}

	// Only purpose of DecodeData is buffer data, if it doesn't need just
	// continue. Note that will only get here if we have a verifier.
	// TODO: check how we can conveniently limit the buffer size.
	if !endStream {
		// wait for all data
		return datastatus.StopIterationAndBuffer
	}

	// We're at the end of stream. Add the last piece of data to the buffer
	dc := f.Native.DecoderCallbacks()
	dc.AddDecodedData(data, true)

	// Start verifying with the body
	f.Logger().Debug("[DecodeData] start verify with body")
	f.startVerify(dc.DecodingBuffer().NewReader(0))

	return datastatus.StopIterationAndWatermark
}

func (f *security) DecodeTrailers(trailes envoy.RequestTrailerMap) trailersstatus.Type {

	f.Logger().Debug("[DecodeTrailers] called")

	if f.state == Calling {
		return trailersstatus.StopIteration
	}
	return trailersstatus.Continue
}

func (f *security) EncodeHeaders(headers envoy.ResponseHeaderMap, endStream bool) headersstatus.Type {

	f.Logger().Debug("[EncodeHeaders] called")

	if f.signer == nil || !f.signer.SigningRequired(headers, f.authResponse) {
		return headersstatus.Continue
	}

	f.responseHeaders = headers
	if !endStream {
		f.state = WaitingForResponseBody
		return headersstatus.StopIteration
	}

	f.startSigning(nil)
	return headersstatus.StopAllIterationAndWatermark
}

func (f *security) EncodeData(data envoy.BufferInstance, endStream bool) datastatus.Type {
	f.Logger().Debug("[EncodeData] called")

	if f.state != WaitingForResponseBody {
		return datastatus.Continue
	}

	if endStream {
		f.Native.EncoderCallbacks().AddEncodedData(data, true)
		f.startSigning(f.Native.EncoderCallbacks().EncodingBuffer().NewReader(0))
		return datastatus.StopIterationAndWatermark
	}
	return datastatus.StopIterationAndBuffer
}

// OnPost will be called from a filter from C side after we Post to get back to the "main-thread"
func (f *security) OnPost(tag uint64) {
	f.Logger().Debug("[OnPost] called")
	switch tag {
	case authPost:
		f.endVerify()
	case signPost:
		f.endSigning()
	}
}

// OnComplete implement for Callbacks interface, will be called by Verifier
func (f *security) OnComplete(response context.AuthResponse) {
	f.Logger().Debug("[OnComplete] called")
	f.authResponse = response
	f.Native.Post(authPost)
	f.Unpin()
}

func (f *security) OnCompleteSigning(signResp context.SignResponse) {
	f.Logger().Debug("[OnCompleteSigning] called")

	f.signResponse = signResp
	f.Native.Post(signPost)
}

func (f *security) startVerify(body io.Reader) {
	f.Logger().Debug("[startVerify] called")
	f.state = Calling
	ctx := context.CreateRequestContext(f, f.Context, f.Native.DecoderCallbacks().ActiveSpan(), f.requestHeaders, f.secrets, body, f.Logger())
	f.Pin()
	go func() {
		f.verifier.Verify(ctx)
	}()
}

func (f *security) endVerify() {

	f.Logger().Debug("[endVerify] called")

	// This stream has been reset, abort the callback.
	if f.state == Responded {
		return
	}

	dc := f.Native.DecoderCallbacks()
	response := f.authResponse

	switch response.Status {
	case context.AuthOK:
		headers := f.requestHeaders
		for k := range response.HeadersToRemove {
			headers.Remove(k)
		}
		for k, v := range response.HeadersToSet {
			headers.SetCopy(k, v)
		}
		for k, v := range response.HeadersToAppend {
			headers.AppendCopy(k, v)
		}
		for k, v := range response.FilterState {
			f.Native.DecoderCallbacks().StreamInfo().FilterState().SetData(context.FilterStatePrefix+k, v, statetype.ReadOnly, lifespan.DownstreamRequest)
		}
		f.config.stats.authOK.Inc()
		f.state = Complete

		// This function has been called only from OnPost, result of OnComplete from a goroutine
		// So we need to continue decoding because the case not allow already handled above
		dc.ContinueDecoding()

	case context.AuthDenied:
		// TODO: support modify headers of the local response to client
		f.Logger().Warn("[endVerify] could not authenticate the request", logger.Data{
			"status_code": response.StatusCode,
		})
		f.state = Responded

		// use only HeadersToSet for now. Consider adding HeaderToAppend and HeaderToRemove if any use cases.
		dc.SendLocalReply(response.StatusCode, response.Body, response.HeadersToSet, "")
		f.config.stats.authDenied.Inc()

	case context.AuthError:
		// TODO: check possibility to support failureModeAllow
		f.Logger().Warn("[endVerify] rejected the request with an error", logger.Data{
			"status_code": response.StatusCode,
		})
		f.state = Responded

		// use only HeadersToSet for now. Consider adding HeaderToAppend and HeaderToRemove if any use cases.
		dc.SendLocalReply(response.StatusCode, response.Body, response.HeadersToSet, "")
		f.config.stats.authError.Inc()

	default:
		f.Logger().Warn("[endVerify] unknown response status from the verifier", logger.Data{
			"status": response.Status,
		})
		f.state = Responded
		f.config.stats.authError.Inc()
	}
}

func (f *security) startSigning(body io.Reader) {

	f.Logger().Debug("[startSigning] called")

	f.state = Signing
	var ctx = context.CreateResponseContext(
		f, f.Context, f.Native.EncoderCallbacks().ActiveSpan(), f.secrets, f.authResponse, f.requestHeaders, f.responseHeaders, body, f.Logger())

	f.Pin()
	go func() {
		defer f.Unpin()
		f.signer.Sign(ctx)
	}()
}

func (f *security) endSigning() {
	f.Logger().Debug("[endSigning] called")
	if f.signResponse.StatusCode != 0 {
		f.responseHeaders.SetStatus(f.signResponse.StatusCode)
	}
	for k, v := range f.signResponse.HeadersToSet {
		f.responseHeaders.SetCopy(k, v)
	}

	f.Native.EncoderCallbacks().ContinueEncoding()
}
