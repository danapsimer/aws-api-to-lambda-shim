# AWS API Gateway to Lambda Shim using [http.Handler](https://golang.org/pkg/net/http/#Handler)

This software is based, in part, on the work done by [eawsy](http://github.com/eawsy/aws-lambda-go/) 
 Their terrific idea was to use the Python 2.7 runtime available in [AWS Lambda](https://aws.amazon.com/lambda/)
 to run [go](http://golang.com) programs.  I was in the midst of creating a new API using [goa](http://goa.design)
 and it dawned on me that if I could somehow trick goa into responding to 
 [AWS Api Gateway](https://aws.amazon.com/api-gateway/) requests, I could have a service that could both run standalone 
 and serverless on AWS's API Gateway and Lambda.

So I created this.

Now that AWS Lambda supports GO programs natively there is no reason to use
eawsy but I have maintained support for it.  I have updated the usage instructions
to include examples for both AWS Native and eAWSy.

## Usage

### AWS Native


For complete details checkout the [helloWorld example](examples/helloWorld/aws)

The basic steps for usage are as follows:

**NOTE:** you need to replace the '*${AWS_ACCOUNT_ID}*' place holder in
 the swagger.json and the aws commands above with your account id.

1. separate your configuration of your web service muxer from your call to http.ListenAndServe.

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
3. create your main() for your aws:
     ```go
    package main
     import (
        "github.com/danapsimer/aws-lambda-shim/examples/helloWorld/hello"
        "github.com/danapsimer/aws-lambda-shim/aws"
    )
     func init() {
        shim.NewHttpHandlerShim(hello.InitHandler)
    }
     func main() {
    }
    ```
3. Make your executable:

    ```Makefile
    build:
    	GOOS=linux go build -o handler

    pack:
    	zip handler.zip handler
    ```
4. Create your lambda function in AWS: (in the directory your handler was built)

    ```bash
    aws lambda create-function \
        --function-name hello-world-api \
        --runtime go1.x --handler handler --zip-file fileb://handler.zip \
        --role arn:aws:iam::${AWS_ACCOUNT_ID}:role/lambda_basic_execution
    ```
5. Create your API Gateway API:

    ```bash
    aws apigateway import-rest-api \
      --body file://examples/helloWorld/swagger.json --region us-east-1
    aws lambda add-permission --region us-east-1 \
      --function-name hello-world-api --statement-id 5 \
      --principal apigateway.amazonaws.com --action lambda:InvokeFunction \
      --source-arn 'arn:aws:execute-api:us-east-1:${AWS_ACCOUNT_ID}:3l3za8xwnd/*/*/*'
    ```

### Eawsy

For complete details checkout the [helloWorld example](examples/helloWorld/eawsy)

The basic steps for usage are as follows:

**NOTE:** you need to replace the '*${AWS_ACCOUNT_ID}*' place holder in 
 the swagger.json and the aws commands above with your account id.
 
1. separate your configuration of your web service muxer from your call to http.ListenAndServe. 
  
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
3. create your main() for your eawsy:

    ```go
    package main

    import (
        "github.com/danapsimer/aws-lambda-shim/examples/helloWorld/hello"
        "github.com/danapsimer/aws-lambda-shim/eawsy"
    )

    func init() {
        shim.NewHttpHandlerShim(hello.InitHandler)
    }

    func main() {
    }
    ```
4. Make your eawsy:

    ```Makefile
    build:
    	go build -buildmode=c-shared -ldflags="-w -s" -o handler.so

    pack:
    	zip handler.zip handler.so
    ```
5. Create your eawsy in AWS: (in the directory your eawsy handler was built)
    
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

