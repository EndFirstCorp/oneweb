package oneweb

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func getMockRouter() *ControllerRoutingHandler {
	router := NewControllerRoutingHandler()
	router.RegisterController("projects", &MockController{})
	return router
}

func TestRegisterController(t *testing.T) {
	router := NewControllerRoutingHandler()
	err := router.RegisterController("projects", &MockController{})
	expectedErr := `Method "Bogus" error: Unsupported http verb: ""
Method "GetTooFewReturns" error: Unsupported return type.  Expected (string, error)
Method "GetWrongReturnType" error: Unsupported return type.  Expected (string, error)
Method "PutBogus" error: Requires either 3 or 4 input args (cr *ControllerRequest, id string, actionFilter string [optional], json *YourStruct or []YourStruct)
`
	if len(router.controllerMethods) != 8 || expectedErr != err.Error() {
		t.Fatal("expected 8 valid controller methods with errors for other 5: ", err)
	}
}

func TestWriteResponse(t *testing.T) {
	rw := httptest.NewRecorder()
	writeResponse(rw, "hello")
	if len(rw.HeaderMap) != 2 || rw.Body.String() != "hello" {
		t.Fatal("expected 2 calls to get header and 1 call to write")
	}
}

func TestGetMethodName(t *testing.T) {
	method := getMethodName("GET", &ControllerRequest{ControllerName: "projects", ControllerFilter: "123", Action: "Stuff"})
	if method != "GetStuff" {
		t.Fatal("expected GetStuff method.  Actual: ", method)
	}
}

func TestGetMethodNameIndex(t *testing.T) {
	method := getMethodName("GET", &ControllerRequest{})
	if method != "Index" {
		t.Fatal("expected Index method.  Actual: ", method)
	}
}

func TestGetMethodNameUpdate(t *testing.T) {
	method := getMethodName("PUT", &ControllerRequest{})
	if method != "Put" {
		t.Fatal("expected Put method.  Actual: ", method)
	}
}

func TestGetMethod(t *testing.T) {
	router := NewControllerRoutingHandler()
	router.RegisterController("Test", &MockController{})
	method := router.getMethod("Test", "Index")
	retVal := method.Call([]reflect.Value{reflect.ValueOf(&ControllerRequest{})})
	if retVal[0].String() != "called Index" {
		t.Fatal("expected to be able to call Index method")
	}
}

func TestGetMethodFailed(t *testing.T) {
	router := NewControllerRoutingHandler()
	router.RegisterController("Test", &MockController{})
	method := router.getMethod("Test", "GetStuff")
	if method != nil {
		t.Fatal("expected to receive empty method")
	}
}

func TestIsRawMethod(t *testing.T) {
	router := NewControllerRoutingHandler()
	router.RegisterController("Test", &MockController{})
	method := router.getMethod("Test", "GetRawmethod")
	isRaw := isRawMethod(method.Type())
	if !isRaw {
		t.Fatal("expected to be raw method")
	}
}

func TestIsRawPostMethod(t *testing.T) {
	router := NewControllerRoutingHandler()
	router.RegisterController("Test", &MockController{})
	method := router.getMethod("Test", "Post")
	isRaw := isRawMethod(method.Type())
	if !isRaw {
		t.Fatal("expected to be raw method")
	}
}

func TestCallRawGetMethod(t *testing.T) {
	router := NewControllerRoutingHandler()
	router.RegisterController("Test", &MockController{})
	writer := httptest.NewRecorder()
	req := &ControllerRequest{ControllerFilter: "1234", Action: "Rawmethod"}
	method := router.getMethod("Test", "GetRawmethod")
	callRawMethod(req, method, writer, &http.Request{})
	if writer.Body.String() != "called raw GET method" {
		t.Fatal("expected to call raw method")
	}
}

func TestCallRawPostMethod(t *testing.T) {
	writer := httptest.NewRecorder()
	req := ControllerRequest{}
	router := NewControllerRoutingHandler()
	router.RegisterController("Test", &MockController{})
	method := router.getMethod("Test", "Post")
	callRawMethod(&req, method, writer, &http.Request{})
	if writer.Body.String() != "called raw POST method" {
		t.Fatal("expected to call raw method")
	}
}

type SimpleData struct {
	Hello string
}

func TestGetJsonBody(t *testing.T) {
	router := getMockRouter()
	method := router.getMethod("Projects", "Put")
	data, err := getJSONBody(&http.Request{Method: "PUT", Body: ioutil.NopCloser(bytes.NewBufferString(`{ "hello": "there" }`))}, method)
	if err != nil || data.(*SimpleData).Hello != "there" {
		t.Fatal("expected json object with property hello and value there")
	}
}

func TestGetJsonBodyNotPostOrPut(t *testing.T) {
	data, err := getJSONBody(&http.Request{Method: "GET", Body: ioutil.NopCloser(bytes.NewBufferString(`{ "hello": "there" }`))}, nil)
	if data != nil || err != nil {
		t.Fatal("expected empty return values")
	}
}

func TestGetJsonErrors(t *testing.T) {
	router := getMockRouter()
	method := router.getMethod("Projects", "Post")
	_, err := getJSONBody(&http.Request{Method: "POST", Body: &MockErroringReadCloser{}}, method)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetArgumentsForIndexPage(t *testing.T) {
	args := getRequestArguments("GET", &ControllerRequest{}, "1234")
	if len(args) != 1 {
		t.Fatal("expected 1 arguments.  Actual: ", args)
	}
}

func TestGetArgumentsWithFilter(t *testing.T) {
	args := getRequestArguments("GET", &ControllerRequest{ControllerFilter: "1234"}, "1234")
	if len(args) != 2 || (args)[1].String() != "1234" {
		t.Fatal("expected 1 arguments with value 1234.  Actual:", args)
	}
}

func TestGetArgumentsWithFilterAndQueryFilter(t *testing.T) {
	args := getRequestArguments("GET", &ControllerRequest{ControllerFilter: "1234", Action: "Stuff", ActionFilter: "4567"}, "1234")
	if len(args) != 3 || (args)[1].String() != "1234" || args[2].String() != "4567" {
		t.Fatal("expected 2 arguments with value 1234 and 4567.  Actual:", args)
	}
}

func TestGetArgumentForDelete(t *testing.T) {
	args := getRequestArguments("DELETE", &ControllerRequest{ControllerFilter: "1234"}, "1234")
	if len(args) != 2 || (args)[1].String() != "1234" {
		t.Fatal("expected 2 argument with first equal 1234.  Actual:", args)
	}
}

func TestGetArgumentForPut(t *testing.T) {
	args := getRequestArguments("PUT", &ControllerRequest{ControllerFilter: "1234"}, &SimpleData{Hello: "there"})
	if len(args) != 3 || (args)[1].String() != "1234" || (args)[2].Interface().(*SimpleData).Hello != "there" {
		t.Fatal("expected 2 arguments with first equal 1234 and second a map.  Actual:", args)
	}
}

func TestGetArgumentForPost(t *testing.T) {
	args := getRequestArguments("POST", &ControllerRequest{}, &SimpleData{Hello: "there"})
	if len(args) != 2 || (args)[1].Interface().(*SimpleData).Hello != "there" {
		t.Fatal("expected 2 argument with a map.  Actual:", args)
	}
}

func TestGetArgumentInvalid(t *testing.T) {
	args := getRequestArguments("BOGUS", &ControllerRequest{}, &SimpleData{Hello: "there"})
	if len(args) != 1 {
		t.Fatal("expected 1 arguments.  Actual:", args)
	}
}

func TestCheckRuntimeArgumentsSuccess(t *testing.T) {
	router := getMockRouter()
	cr := &ControllerRequest{ControllerName: "Tests", ControllerFilter: "1234"}
	args := []reflect.Value{reflect.ValueOf(cr), reflect.ValueOf("1234"), reflect.ValueOf(&SimpleData{})}
	err := checkRuntimeArguments(router.getMethod("Projects", "Put"), args, "Put", "PUT", cr)
	if err != nil {
		t.Fatal("Expected to have success")
	}
}

func TestCheckRuntimeArgumentsFail(t *testing.T) {
	router := getMockRouter()
	req := &ControllerRequest{ControllerName: "Tests", ControllerFilter: "1234"}
	args := []reflect.Value{reflect.ValueOf(req), reflect.ValueOf("1234")}
	err := checkRuntimeArguments(router.getMethod("Projects", "Put"), args, "Put", "PUT", req)
	if err == nil || err.Error() != "Invalid Url to call method: \"Put\"" {
		t.Fatal("Expected to have Invalid Url error: ", err)
	}
}

func TestCallControllerMethod(t *testing.T) {
	method := reflect.ValueOf(&MockController{}).MethodByName("Index")
	result, _ := callControllerMethod(&method, []reflect.Value{reflect.ValueOf(&ControllerRequest{})})
	if result != "called Index" {
		t.Fatal("Expected to receive back called message from method")
	}
}

func TestCallControllerMethodWithError(t *testing.T) {
	method := reflect.ValueOf(&MockController{}).MethodByName("GetError")
	_, err := callControllerMethod(&method, []reflect.Value{reflect.ValueOf(&ControllerRequest{}), reflect.ValueOf("1234")})
	if err == nil || err.Error() != "failed" {
		t.Fatal("Expected error return from method", err)
	}
}

func TestHttpHandlerSuccess(t *testing.T) {
	rw := httptest.NewRecorder()
	getMockRouter().controllerRoutingHandler(rw, &http.Request{URL: &url.URL{Path: "/projects"}, Method: "GET"})
	if len(rw.HeaderMap) != 2 {
		t.Fatal("expected to succeed")
	}
}

func TestHttpHandlerMethodNotFound(t *testing.T) {
	rw := httptest.NewRecorder()
	getMockRouter().controllerRoutingHandler(rw, &http.Request{URL: &url.URL{Path: "/projects/123/bogus"}, Method: "GET"})
	if rw.Body.String() != "Method \"GetBogus\" not found\n" {
		t.Fatal("expected to be unable to find method", rw.Body.String())
	}
}

func TestHttpHandlerGetMethod(t *testing.T) {
	rw := httptest.NewRecorder()
	getMockRouter().controllerRoutingHandler(rw, &http.Request{URL: &url.URL{Path: "/projects/123/method"}, Method: "GET"})
	if rw.Body.String() != "called GetMethod" {
		t.Fatal("expected to be able to call GetMethod method")
	}
}

func TestHttpHandlerCallRaw(t *testing.T) {
	rw := httptest.NewRecorder()
	getMockRouter().controllerRoutingHandler(rw, &http.Request{URL: &url.URL{Path: "/projects/123/Rawmethod"}, Method: "GET"})
	if !strings.Contains(rw.Body.String(), "called raw GET method") {
		t.Fatal("expected called raw method message, Actual:", rw.Body.String())
	}
}

func TestHttpHandlerCantGetJson(t *testing.T) {
	rw := httptest.NewRecorder()
	getMockRouter().controllerRoutingHandler(rw, &http.Request{URL: &url.URL{Path: "/projects/123/valid"}, Method: "PUT", Body: &MockErroringReadCloser{}})
	if !strings.Contains(rw.Body.String(), "Failed to read JSON data:") {
		t.Fatal("expected failure getting JSON", rw.Body.String())
	}
}

func TestHttpHandlerInvalidArguments(t *testing.T) {
	rw := httptest.NewRecorder()
	getMockRouter().controllerRoutingHandler(rw, &http.Request{URL: &url.URL{Path: "/projects"}, Method: "PUT", Body: ioutil.NopCloser(bytes.NewBufferString(`{ "hello": "there" }`))})
	if rw.Body.String() != "Invalid Url to call method: \"Put\"\n" {
		t.Fatal("should've gotten bogus arguments: ", rw.Body.String())
	}
}

func TestHttpHandlerErroringMethod(t *testing.T) {
	rw := httptest.NewRecorder()
	getMockRouter().controllerRoutingHandler(rw, &http.Request{URL: &url.URL{Path: "/projects/1/error"}, Method: "GET"})
	if rw.Body.String() != "Internal error calling controller method: failed\n" {
		t.Fatal("should've had an error: ", rw.Body.String())
	}
}

func TestHandler(t *testing.T) {
	handler := NewControllerRoutingHandler().Handler()
	if reflect.TypeOf(handler) != reflect.TypeOf((*http.HandlerFunc)(nil)).Elem() {
		t.Fatal("expected a http handler function")
	}
}

type MockErroringReadCloser struct {
	io.ReadCloser
}

func (c *MockErroringReadCloser) Read(p []byte) (n int, err error) {
	return 0, errors.New("failed")
}

func (c *MockErroringReadCloser) Close() error {
	return nil
}
