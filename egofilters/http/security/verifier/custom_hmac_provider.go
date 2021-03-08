// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package verifier

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/grab/ego/ego/src/go/envoy"

	"github.com/grab/ego/egofilters/http/security/context"
	securityhttp "github.com/grab/ego/egofilters/http/security/http"
	pb "github.com/grab/ego/egofilters/http/security/proto"
)

const hmacUserIDSessionKey = "UserID"

type getCurrentTimeOpt func() time.Time

func getCurrentTime() time.Time {
	return time.Now()
}

// CreateCustomHMACProvider ...
func CreateCustomHMACProvider(provider *pb.CustomHMACProvider) (*customHMACProvider, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	return createCustomHMACProvider(provider, securityhttp.NewHttpClientWithCtx(client), getCurrentTime, isValidSignature)
}

func createCustomHMACProvider(
	provider *pb.CustomHMACProvider, client securityhttp.HttpClientWithCtx, getCurrentTime getCurrentTimeOpt, isValidSignature isValidHMACSignatureOpt) (*customHMACProvider, error) {
	return &customHMACProvider{
		provider:         provider,
		client:           client,
		getCurrentTime:   getCurrentTime,
		isValidSignature: isValidSignature,
	}, nil
}

type customHMACProvider struct {
	baseProvider
	provider         *pb.CustomHMACProvider
	client           securityhttp.HttpClientWithCtx
	getCurrentTime   getCurrentTimeOpt
	isValidSignature isValidHMACSignatureOpt
}

func (v *customHMACProvider) Verify(ctx context.RequestContext) {
	// validate headers.
	parts := strings.SplitN(string(ctx.Headers().Authorization().Copy()), ":", 3)
	if 2 != len(parts) {
		// TODO: this should go to request log
		ctx.Logger().Error("[Verify] invalid token length. Expected length of 2.", len(parts))
		v.reportUnauthorizedError(ctx)
		return
	}

	v.checkHMACSignature(ctx, parts)
}

func (v *customHMACProvider) checkHMACSignature(ctx context.RequestContext, parts []string) {

	serviceKey := ctx.GetSecret(v.provider.ServiceKey)
	serviceToken := ctx.GetSecret(v.provider.ServiceToken)

	url, err := url.Parse(v.provider.RequestValidationUrl)
	if err != nil {
		ctx.Logger().Error("[Verify] can't parse url.", url, err)
		v.reportInternalError(ctx)
		return
	}

	request, err := http.NewRequestWithContext(ctx.GoContext(), http.MethodPost, url.String(), ctx.BodyReader())
	if err != nil {
		ctx.Logger().Error("[Verify] can't new http request with context.", err)
		v.reportInternalError(ctx)
		return
	}

	userID, signature := parts[0], parts[1]

	request.Header.Set("Authorization", "Token "+serviceKey+" "+serviceToken)
	request.Header.Set("Cache-Control", "no-cache")
	request.Header.Set("Content-Type", ctx.Headers().ContentType().Copy())

	request.Header.Set("X-Custom-Auth-Date", ctx.Headers().Get("Date").Copy())
	request.Header.Set("X-Custom-Auth-Path", ctx.Headers().Path().Copy())
	request.Header.Set("X-Custom-Auth-Signature", signature)
	request.Header.Set("X-Custom-Auth-Verb", ctx.Headers().Method().Copy())
	request.Header.Set(requestIDHeader, ctx.Headers().Get(requestIDHeader).Copy())

	spanName := ""
	if v.provider.TracingEnabled {
		spanName = "custom_hmac_verify"
	}

	httpResponse, httpErr := v.client.DoWithTracing(ctx, request, spanName)
	v.handleHMACSignatureCheckResponse(ctx, userID, httpResponse, httpErr)
}

func (v *customHMACProvider) handleHMACSignatureCheckResponse(ctx context.RequestContext, userID string, httpResponse *http.Response, httpErr error) {
	if httpErr != nil {
		ctx.Logger().Error("[Verify] error while calling custom auth provider.", httpErr)
		v.reportInternalError(ctx)
		return
	}

	if httpResponse == nil {
		ctx.Logger().Error("[Verify] nil response from custom auth provider.")
		v.reportInternalError(ctx)
		return
	}
	defer httpResponse.Body.Close()

	ctx.Logger().Debug("[Verify] status code from custom auth provider", httpResponse.StatusCode)

	valid, err := v.isValidSignature(httpResponse)
	if err != nil {
		ctx.Logger().Error("[Verify] can't validate signature check response.", err)
		v.reportInternalError(ctx)
		return
	}

	if valid {
		supportedHeaders := map[string]string{
			userIDHeader: userID,
		}

		resp := context.AuthResponseOK()
		resp.HeadersToSet = make(map[string]string)
		resp.HeadersToRemove = make(map[string]struct{})

		for _, h := range v.provider.GeneratedUpstreamHeaders {
			if v, ok := supportedHeaders[h]; ok {
				resp.HeadersToSet[h] = v
				resp.HeadersToRemove[h] = struct{}{}
			}
		}

		resp.FilterState = map[string]string{
			hmacUserIDSessionKey: userID,
		}
		ctx.Callbacks().OnComplete(resp)
		return
	}

	v.reportUnauthorizedError(ctx)
	return
}

func (v *customHMACProvider) reportInternalError(ctx context.RequestContext) {
	ctx.Callbacks().OnComplete(context.AuthResponseError())
}

func (v *customHMACProvider) reportUnauthorizedError(ctx context.RequestContext) {
	ctx.Callbacks().OnComplete(context.AuthResponseUnauthorized())
}

func (v *customHMACProvider) WithBody() bool {
	return true
}

// CustomHMACSignResponse ...
type CustomHMACSignResponse struct {
	Signature string `json:"signature"`
}

func (v *customHMACProvider) Sign(ctx context.ResponseContext) {
	serviceKey := ctx.GetSecret(v.provider.ServiceKey)
	serviceToken := ctx.GetSecret(v.provider.ServiceToken)

	signURL, err := url.Parse(v.provider.ResponseSigningUrl)
	if err != nil {
		ctx.Logger().Error("[Sign] can't parse ResponseSignURL.", err)
		ctx.Callbacks().OnCompleteSigning(context.SignResponse{
			StatusCode: http.StatusInternalServerError,
		})
		return
	}
	signReq, err := http.NewRequestWithContext(ctx.GoContext(), http.MethodPost, signURL.String(), ctx.BodyReader())
	if err != nil {
		ctx.Logger().Error("[Sign] can't create sign request.", err)
		ctx.Callbacks().OnCompleteSigning(context.SignResponse{
			StatusCode: http.StatusInternalServerError,
		})
		return
	}
	// golang will format to UTC by default so here need to force the timezone info to GMT
	// RFC1123 is the preferred time format for RFC7231,
	// eg. Sun, 06 Nov 1994 08:49:37 GMT
	signedTime := v.getCurrentTime().In(time.FixedZone("GMT", 0)).Format(time.RFC1123)
	signReq.Header.Set("X-Custom-Auth-Date", signedTime)
	signReq.Header.Set("Authorization", "Token "+serviceKey+" "+serviceToken)
	signReq.Header.Set("X-Custom-Auth-Status-Code", ctx.Headers().Status().Copy())

	signReq.Header.Set(requestIDHeader, ctx.RequestHeaders().Get(requestIDHeader).Copy())

	spanName := ""
	if v.provider.TracingEnabled {
		spanName = "custom_hmac_sign"
	}

	resp, err := v.client.DoWithTracing(ctx, signReq, spanName)

	if err != nil {
		ctx.Logger().Error("[Sign] can't send sign request.", err)
		ctx.Callbacks().OnCompleteSigning(context.SignResponse{
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	if resp == nil {
		ctx.Logger().Error("[Sign] empty sign response boy.", err)
		ctx.Callbacks().OnCompleteSigning(context.SignResponse{
			StatusCode: http.StatusInternalServerError,
		})
		return
	}
	defer resp.Body.Close()

	signResponse := &CustomHMACSignResponse{}
	if resp.StatusCode != http.StatusOK {
		ctx.Logger().Warn("[Sign] unknown status code", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(signResponse)
	if err != nil {
		ctx.Logger().Error("[Sign] can't unmarshal response.", err)
		ctx.Callbacks().OnCompleteSigning(context.SignResponse{
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	ctx.Callbacks().OnCompleteSigning(context.SignResponse{
		HeadersToSet: map[string]string{
			"X-Custom-Auth-Signature-HMAC-SHA256": signResponse.Signature,
			"Date":                               signedTime,
		},
	})
}

func (v *customHMACProvider) SigningRequired(headers envoy.ResponseHeaderMap, authResp context.AuthResponse) bool {
	if nil == headers || nil == authResp.FilterState || "" == authResp.FilterState[hmacUserIDSessionKey] {
		return false
	}

	status, err := strconv.ParseUint(string(headers.Status()), 10, 64)
	if err != nil {
		return false
	}

	return v.provider.SignResp && status < http.StatusInternalServerError
}
