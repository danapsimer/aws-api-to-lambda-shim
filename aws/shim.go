package aws

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/danapsimer/aws-api-to-lambda-shim/utils"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type ShimInitFunc func() (http.Handler, error)

type HttpHandlerShim struct {
	init       ShimInitFunc
	initCalled bool
	handler    http.Handler
}

func NewHttpHandlerShim(init ShimInitFunc) (*HttpHandlerShim, error) {
	shim := &HttpHandlerShim{
		init:       init,
		initCalled: false,
		handler:    nil,
	}
	lambda.Start(func(ctx context.Context, evt events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		return shim.handle(ctx, evt)
	})
	return shim, nil
}

func (shim *HttpHandlerShim) handle(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if !shim.initCalled {
		handler, err := shim.init()
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			return events.APIGatewayProxyResponse{StatusCode: 500, Body: err.Error()}, err
		}
		shim.handler = handler
		shim.initCalled = true
	}

	var urlStr string
	if len(request.QueryStringParameters) != 0 {
		urlStr = fmt.Sprintf("%s?%s", request.Path, utils.QueryStringParams2Values(request.QueryStringParameters).Encode())
	} else {
		urlStr = request.Path
	}
	url, err := url.ParseRequestURI(urlStr)
	if err != nil {
		log.Printf("ERROR: %s", err.Error())
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: err.Error()}, err
	}
	var bodyReader io.Reader
	bodyReader = strings.NewReader(request.Body)
	if request.IsBase64Encoded {
		bodyReader = base64.NewDecoder(base64.StdEncoding, bodyReader)
	}
	httpRequest := http.Request{
		Method:        request.HTTPMethod,
		URL:           url,
		Proto:         "HTTP/1.0",
		ProtoMajor:    1,
		ProtoMinor:    0,
		Header:        utils.Headers2Header(request.Headers),
		Body:          ioutil.NopCloser(bodyReader),
		ContentLength: int64(len(request.Body)),
		Close:         false,
		Host:          request.Headers["Host"],
		RemoteAddr:    request.Headers["Host"],
		RequestURI:    url.String(),
	}
	responseWriter, err := utils.NewLambdaResponseWriter()
	if err != nil {
		log.Printf("ERROR: %s", err.Error())
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: err.Error()}, nil
	}
	shim.handler.ServeHTTP(responseWriter, httpRequest.WithContext(ctx))
	if responseWriter.ResponseBuffer.Len() > 0 {
		if _, hasContentType := responseWriter.Header()["Content-Type"]; !hasContentType {
			responseWriter.Header().Add("Content-Type", http.DetectContentType(responseWriter.ResponseBuffer.Bytes()))
		}
	}
	responseBody := responseWriter.ResponseBuffer.String()
	response := events.APIGatewayProxyResponse{
		StatusCode: int(responseWriter.StatusCode),
		Body:       responseBody,
		Headers:    utils.Header2Headers(responseWriter.Header()),
	}
	return response, nil
}
