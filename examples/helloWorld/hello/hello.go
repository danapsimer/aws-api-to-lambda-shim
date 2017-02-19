package hello

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
)

var (
	helloMatcher = regexp.MustCompile("/hello/(.*)")
)

type greeting struct {
	Greeting string `json:"greeting,omitempty"`
}

func InitHandler() (http.Handler, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello/", func(w http.ResponseWriter, req *http.Request) {
		matches := helloMatcher.FindAllStringSubmatch(req.URL.Path, -1)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		name := matches[0][1]
		if name == "" {
			name = "World"
		}
		greeting := greeting{fmt.Sprintf("Hello, %s!", name)}
		jsonEncoder := json.NewEncoder(w)
		jsonEncoder.Encode(greeting)
	})
	return mux, nil
}
