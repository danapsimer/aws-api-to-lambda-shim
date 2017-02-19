# System build

build:
	go build -buildmode=c-shared -ldflags="-w -s" -o handler.so
	chown `stat -c "%u:%g" .` handler.so

pack:
	zip handler.zip handler.so
	chown `stat -c "%u:%g" .` handler.zip

# Docker build for OS X or Windoze

dbuild:
	@docker run --rm \
		-v $(GOPATH):/go \
		-w /go/src/github.com/danapsimer/aws-api-to-lambda-shim/examples/helloWorld/lambda \
		eawsy/aws-lambda-go make build

dpack:
	@docker run --rm \
		-v $(GOPATH):/go \
		-w /go/src/github.com/danapsimer/aws-api-to-lambda-shim/examples/helloWorld/lambda \
		eawsy/aws-lambda-go make pack

clean:
	@rm -rf handler.zip handler.so

deploy:
	@aws lambda update-function-code --function-name hello-world-api --zip-file fileb://handler.zip
