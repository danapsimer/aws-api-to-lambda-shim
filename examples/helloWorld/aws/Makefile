AWS_ACCOUNT_ID := $(shell aws iam get-user | jq -r '.User.Arn' | cut -d : -f 5)
$(info AWS_ACCOUNT_ID = '$(AWS_ACCOUNT_ID)')

AWS_ROLE := arn:aws:iam::$(AWS_ACCOUNT_ID):role/lambda_basic_execution

build:
	GOOS=linux go build -o handler

pack:
	zip handler.zip handler

clean:
	@rm -rf handler.zip handler

create:
	@aws lambda create-function \
		--region us-east-1 \
		--function-name hello-world-api-go \
		--zip-file fileb://handler.zip \
		--runtime go1.x \
		--role $(AWS_ROLE) \
		--handler handler

update:
	@aws lambda update-function-code --function-name hello-world-api-go --zip-file fileb://handler.zip