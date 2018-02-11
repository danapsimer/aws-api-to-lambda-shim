package main

import (
	"github.com/danapsimer/aws-api-to-lambda-shim/aws"
	"github.com/danapsimer/aws-api-to-lambda-shim/examples/helloWorld/hello"
)

func init() {
	aws.NewHttpHandlerShim(hello.InitHandler)
}

func main() {
}
