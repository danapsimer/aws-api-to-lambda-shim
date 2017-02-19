# AWS Lambda HTTP Handler Shim

This software is based, in part, on the work done by [eawsy](http://github.com/eawsy/aws-lambda-go/) 
 Their terrific idea was to use the Python 2.7 runtime available in [AWS Lamnda](https://aws.amazon.com/lambda/)
 to run [go](http://golang.com) programs.  I was in the midst of creating a new API using [goa](http://goa.design)
 and it dawned on me that if I could somehow trick goa into responding to [AWS Api Gateway](https://aws.amazon.com/api-gateway/)
 I could have a service that could both run standalone and serverless on API Gateway and Lambnda.

So I created this.

## Usage

For complete details checkout the [helloWorld example](examples/helloWorld)

The basic steps for usage are as follows:

**NOTE:** you need to replace the '*${AWS_ACCOUNT_ID}*' place holder in 
 the swagger.json and the aws commands above with your account id.
 
1. separate your configuration of your web service muxer from your call to
  http.ListenAndServe. 
    
    ```go
    package hello

    import (
      "net/http"
      ...
    )
 
    func InitHandler() (http.Handler, error) {
      mux := http.NewServeMux()
      mux.HandleFunc("/hello/", func(w http.ResponseWriter, req *http.Request) {
        ...
      })
      return mux, nil
    }
    ```
2. Create your main() for your web service: 

    ```go
    package main

    import (
    	"github.com/danapsimer/aws-lambda-shim/examples/helloWorld/hello"
	"log"
	"net/http"
    )

    func main() {
	handler, _ := hello.InitHandler()
	log.Fatal(http.ListenAndServe(":8080", handler))
    }
    ```
3. create your main() for your lambda: 

    ```go
    package main

    import (
        "github.com/danapsimer/aws-lambda-shim/examples/helloWorld/hello"
        "github.com/danapsimer/aws-lambda-shim/shim"
    )

    func init() {
        shim.NewHttpHandlerShim(hello.InitHandler)
    }

    func main() {
    }
    ```
4. Make your lambda: 

    ```Makefile
    build:
    	go build -buildmode=c-shared -ldflags="-w -s" -o handler.so
    	chown `stat -c "%u:%g" .` handler.so

    pack:
    	zip handler.zip handler.so
    	chown `stat -c "%u:%g" .` handler.zip
    ```
5. Create your lambda in AWS: (in the directory your lambda handler was built) 
    
    ```bash
    aws lambda create-function \
        --function-name hello-world-api \
        --runtime python2.7 --handler handler.handle --zip-file fileb://handler.zip \
        --role arn:aws:iam::${AWS_ACCOUNT_ID}:role/lambda_basic_execution
    ```
6. Create your API Gateway API: 
    
    ```bash
    aws apigateway import-rest-api \
      --body file://examples/helloWorld/swagger.json --region us-east-1
    aws lambda add-permission --region us-east-1 \
      --function-name hello-world-api --statement-id 5 \
      --principal apigateway.amazonaws.com --action lambda:InvokeFunction \
      --source-arn 'arn:aws:execute-api:us-east-1:${AWS_ACCOUNT_ID}:3l3za8xwnd/*/*/*'
    ```

