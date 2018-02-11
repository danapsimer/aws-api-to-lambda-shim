package main

import (
	"github.com/danapsimer/aws-api-to-lambda-shim/eawsy"
	"github.com/danapsimer/aws-api-to-lambda-shim/examples/helloWorld/hello"
)

func init() {
	eawsy.NewHttpHandlerShim(hello.InitHandler)
}

func main() {
}
