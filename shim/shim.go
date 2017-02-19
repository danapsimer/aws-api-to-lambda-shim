package shim

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/eawsy/aws-lambda-go/service/lambda/runtime"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"encoding/base64"
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

type lambdaResponseWriter struct {
	header         http.Header
	StatusCode     int
	ResponseBuffer bytes.Buffer
}

func NewLambdaResponseWriter() (*lambdaResponseWriter, error) {
	return &lambdaResponseWriter{
		header:     make(map[string][]string),
		StatusCode: 0,
	}, nil
}

func (w *lambdaResponseWriter) Header() http.Header {
	return w.header
}

func (w *lambdaResponseWriter) Write(bytes []byte) (int, error) {
	if w.StatusCode == 0 {
		w.StatusCode = 200
	}
	bytesWritten, err := w.ResponseBuffer.Write(bytes)
	return bytesWritten, err
}

func (w *lambdaResponseWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
}

func queryStringParams2Values(qsp map[string]string) url.Values {
	if qsp == nil {
		return nil
	}
	values := url.Values{}
	for k, v := range qsp {
		values.Add(k, v)
	}
	return values
}

func headers2Header(headers map[string]string) http.Header {
	if headers == nil {
		return nil
	}
	values := http.Header{}
	for k, v := range headers {
		values.Add(k, v)
	}
	return values
}

func header2Headers(header http.Header) map[string]string {
	if header == nil {
		return nil
	}
	values := make(map[string]string)
	for k, v := range header {
		values[k] = v[0]
	}
	return values
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
		urlStr = fmt.Sprintf("%s?%s", msg.Path, queryStringParams2Values(msg.QueryStringParameters).Encode())
	} else {
		urlStr = msg.Path
	}
	url, err := url.ParseRequestURI(urlStr)
	if err != nil {
		log.Printf("ERROR: %s", err.Error())
		return apiGatewayResponse{StatusCode: 500, Body: err.Error()}, nil
	}
	bodyReader := strings.NewReader(msg.Body)
	if msg.IsBase64Encoded {
		bodyReader = base64.NewDecoder(base64.StdEncoding,bodyReader)
	}
	//log.Printf("url parsed: %v", url)
	httpRequest := http.Request{
		Method:        msg.HttpMethod,
		URL:           url,
		Proto:         "HTTP/1.0",
		ProtoMajor:    1,
		ProtoMinor:    0,
		Header:        headers2Header(msg.Headers),
		Body:          ioutil.NopCloser(bodyReader),
		ContentLength: int64(len(msg.Body)),
		Close:         false,
		Host:          msg.Headers["Host"],
		RemoteAddr:    msg.Headers["Host"],
		RequestURI:    url.String(),
	}
	//log.Printf("httpRequest created: %v", &httpRequest)
	responseWriter, err := NewLambdaResponseWriter()
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
		Headers:    header2Headers(responseWriter.Header()),
	}
	//log.Printf("Response: %v", &response)
	return &response, nil
}
