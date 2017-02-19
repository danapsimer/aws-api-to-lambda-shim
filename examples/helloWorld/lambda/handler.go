package main

import (
	"github.com/danapsimer/aws-api-to-lambda-shim/examples/helloWorld/hello"
	"github.com/danapsimer/aws-api-to-lambda-shim/shim"
)

func init() {
	shim.NewHttpHandlerShim(hello.InitHandler)
}

func main() {
}
