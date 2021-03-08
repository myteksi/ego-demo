// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package getheader

import (
	"fmt"
	"net/http"
	"time"

	pb "github.com/grab/ego/egofilters/http/getheader/proto"
	ego "github.com/grab/ego/ego/src/go"
	"github.com/grab/ego/ego/src/go/envoy"
	"github.com/grab/ego/ego/src/go/envoy/headersstatus"
	"github.com/grab/ego/ego/src/go/envoy/loglevel"
)

type getHeaderFilter struct {
	ego.HttpFilterBase
	settings *pb.Settings
	// Just a simple keep requestHeaders & response of 3rdParty httpCall for OnPost callback
	requestHeaders envoy.RequestHeaderMap
	httpResponse   *http.Response
	httpErr        error
}

func newGetHeaderFilter(settings *pb.Settings, native envoy.GoHttpFilter) ego.HttpFilter {
	f := &getHeaderFilter{settings: settings}
	f.HttpFilterBase.Init(native)
	return f
}

func (f *getHeaderFilter) DecodeHeaders(headers envoy.RequestHeaderMap, endStream bool) headersstatus.Type {
	// keep headers for use later in onPost
	f.requestHeaders = headers

	f.Pin() // must do this for every go-routine
	go func() {
		defer f.Unpin() // must do this for every go-routine. Includes f.Recover()

		f.Context.Done()
		request, err := http.NewRequestWithContext(f.Context, "GET", f.settings.Src, nil)
		if err != nil {
			// Send local repsonse
			f.httpErr = err
			f.Native.Post(0)
			return
		}

		client := http.Client{
			Timeout: 2 * time.Second,
		}
		f.httpResponse, f.httpErr = client.Do(request)

		// In this demo we only need one http-call at a time
		// so tag = 0 because we don't need to manage multiple callback
		f.Native.Post(0)
	}()
	// If we turn `headersstatus.Continue` on `DecodeHeaders` from Go side
	// Although we call a http request with a goroutine and defer `defer f.Release()` so there is a chance
	// for onDestroy happened before goroutine DONE and call a post to dispatcher. ()
	return headersstatus.StopAllIterationAndWatermark
}

func (f *getHeaderFilter) OnPost(tag uint64) {
	if f.httpErr != nil {
		// Send local reply and not forward request to upstream
		errMsg := fmt.Sprintf("can not connect to srcs header", f.httpErr)
		f.Native.Log(loglevel.Error, errMsg)
		f.Native.DecoderCallbacks().SendLocalReply(http.StatusFailedDependency, errMsg, nil, "")

		// Don't need to ContinueDecoding here, because we don't want to continue with forward to upstream
		// If you try to do that OnDestoy may happen before and SendLocalReply will not have decodeCallback
	} else {
		respHeader := f.httpResponse.Header.Get(f.settings.Hdr)
		f.requestHeaders.AddCopy(f.settings.Key, respHeader)
		f.Native.DecoderCallbacks().ContinueDecoding()
	}
}
