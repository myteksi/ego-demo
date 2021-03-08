package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
)

// A simple echo service that dumps request headers and body into the
// response body and adds a custom header needed by one of the examples
func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("x-powered-by", runtime.Version())
		w.WriteHeader(http.StatusOK)

		fmt.Fprintf(w, "%v %v%v\n", r.Method, r.Host, r.URL)

		for name, headers := range r.Header {
			for _, h := range headers {
				fmt.Fprintf(w, "%v: %v\n", name, h)
			}
		}
		fmt.Fprint(w, "\n")

		if body, err := ioutil.ReadAll(r.Body); err != nil {
			fmt.Fprint(w, err.Error())
		} else {
			_, _ = w.Write(body)
		}
	})
	http.ListenAndServe(":8888", nil)
}
