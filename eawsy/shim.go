package eawsy

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/danapsimer/aws-api-to-lambda-shim/utils"
	"github.com/eawsy/aws-lambda-go/service/lambda/runtime"
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
	runtime.HandleFunc(func(evt json.RawMessage, ctx *runtime.Context) (interface{}, error) {
		return shim.handle(evt, ctx)
	})
	return shim, nil
}

type identity struct {
	ApiKey                        string
	UserArn                       string
	AccessKey                     string
	Caller                        string
	UserAgent                     string
	User                          string
	CognitoIdentityPoolId         string
	CognitoIdentityId             string
	CognitoAuthenticationProvider string
	SourceId                      string
	AccountId                     string
}

type requestContext struct {
	ResourceId   string
	ApiId        string
	ResourcePath string
	HttpMethod   string
	RequestId    string
	AccountId    string
	Identity     identity
	Stage        string
}

type apiGatewayMessage struct {
	Body                  string
	Resource              string
	RequestContext        requestContext
	QueryStringParameters map[string]string
	HttpMethod            string
	PathParameters        map[string]string
	Headers               map[string]string
	StageVariable         map[string]string
	Path                  string
	IsBase64Encoded       bool
}

type apiGatewayResponse struct {
	StatusCode int32             `json:"statusCode"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`
}

func (shim *HttpHandlerShim) handle(evt json.RawMessage, ctx *runtime.Context) (interface{}, error) {
	if !shim.initCalled {
		handler, err := shim.init()
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			return nil, err
		}
		shim.handler = handler
		shim.initCalled = true
	}

	log.Printf("payload: %s", string(evt))
	var msg apiGatewayMessage
	err := json.Unmarshal(evt, &msg)
	if err != nil {
		log.Printf("ERROR: %s", err.Error())
		return apiGatewayResponse{StatusCode: 500, Body: err.Error()}, nil
	}

	var urlStr string
	if len(msg.QueryStringParameters) != 0 {
		urlStr = fmt.Sprintf("%s?%s", msg.Path, utils.QueryStringParams2Values(msg.QueryStringParameters).Encode())
	} else {
		urlStr = msg.Path
	}
	url, err := url.ParseRequestURI(urlStr)
	if err != nil {
		log.Printf("ERROR: %s", err.Error())
		return apiGatewayResponse{StatusCode: 500, Body: err.Error()}, nil
	}
	var bodyReader io.Reader
	bodyReader = strings.NewReader(msg.Body)
	if msg.IsBase64Encoded {
		bodyReader = base64.NewDecoder(base64.StdEncoding, bodyReader)
	}
	//log.Printf("url parsed: %v", url)
	httpRequest := http.Request{
		Method:        msg.HttpMethod,
		URL:           url,
		Proto:         "HTTP/1.0",
		ProtoMajor:    1,
		ProtoMinor:    0,
		Header:        utils.Headers2Header(msg.Headers),
		Body:          ioutil.NopCloser(bodyReader),
		ContentLength: int64(len(msg.Body)),
		Close:         false,
		Host:          msg.Headers["Host"],
		RemoteAddr:    msg.Headers["Host"],
		RequestURI:    url.String(),
	}
	//log.Printf("httpRequest created: %v", &httpRequest)
	responseWriter, err := utils.NewLambdaResponseWriter()
	if err != nil {
		log.Printf("ERROR: %s", err.Error())
		return apiGatewayResponse{StatusCode: 500, Body: err.Error()}, nil
	}
	//log.Printf("calling service.Mux(%v,%v)", responseWriter, &httpRequest)
	shim.handler.ServeHTTP(responseWriter, &httpRequest)
	if responseWriter.ResponseBuffer.Len() > 0 {
		if _, hasContentType := responseWriter.Header()["Content-Type"]; !hasContentType {
			responseWriter.Header().Add("Content-Type", http.DetectContentType(responseWriter.ResponseBuffer.Bytes()))
		}
	}
	responseBody := responseWriter.ResponseBuffer.String()
	response := apiGatewayResponse{
		StatusCode: int32(responseWriter.StatusCode),
		Body:       responseBody,
		Headers:    utils.Header2Headers(responseWriter.Header()),
	}
	//log.Printf("Response: %v", &response)
	return &response, nil
}
