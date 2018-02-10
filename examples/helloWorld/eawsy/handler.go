package main

import (
	"github.com/danapsimer/aws-api-to-lambda-shim/examples/helloWorld/hello"
	"github.com/danapsimer/aws-api-to-lambda-shim/eawsy"
)

func init() {
	eawsy.NewHttpHandlerShim(hello.InitHandler)
}

func main() {
}
