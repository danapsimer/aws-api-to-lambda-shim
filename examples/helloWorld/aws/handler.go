package main

import (
	"github.com/danapsimer/aws-api-to-lambda-shim/examples/helloWorld/hello"
	"github.com/danapsimer/aws-api-to-lambda-shim/aws"
)

func init() {
	aws.NewHttpHandlerShim(hello.InitHandler)
}

func main() {
}
