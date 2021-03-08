package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	http.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {
		// small pseudo-hmac based on sum length
		// signature = len(METHOD) + len(PATH) + len(BODY)
		reqSignature := r.Header.Get("X-Custom-Auth-Signature")
		reqVerb := r.Header.Get("X-Custom-Auth-Verb")
		reqPath := r.Header.Get("X-Custom-Auth-Path")
		verifedSignature := len(reqVerb) + len(reqPath)
		if body, err := ioutil.ReadAll(r.Body); err == nil {
			verifedSignature = verifedSignature + len(body)
		}
		if reqSignature == fmt.Sprintf("%v", verifedSignature) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"valid": true}`)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"valid": false}`)
		}
	})

	http.HandleFunc("/sign", func(w http.ResponseWriter, r *http.Request) {
		// resp signature = str(status_code) + len(BODY)
		respStatusCode := r.Header.Get("X-Custom-Auth-Status-Code")
		bodyLen := 0
		if body, err := ioutil.ReadAll(r.Body); err == nil {
			bodyLen = len(body)
		}
		signedSignature := fmt.Sprintf("%s_%d", respStatusCode, bodyLen)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"signature": "%s"}`, signedSignature)
	})
	http.ListenAndServe(":8889", nil)
}
