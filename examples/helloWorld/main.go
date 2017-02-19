package main

import (
	"github.com/danapsimer/aws-api-to-lambda-shim/examples/helloWorld/hello"
	"log"
	"net/http"
)

func main() {
	handler, _ := hello.InitHandler()
	log.Fatal(http.ListenAndServe(":8080", handler))
}
