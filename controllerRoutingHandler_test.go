package oneweb

import (
	"bytes"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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
Method "GetBogus" error: Requires 1 input arg (cr *controllerRequest)
Method "GetTooFewReturns" error: Unsupported return type.  Expected (string, error)
Method "GetWrongReturnType" error: Unsupported return type.  Expected (string, error)
Method "PutBogus" error: Requires 2 input args (cr *controllerRequest, json *YourStruct or []YourStruct)
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
	method := getMethodName("GET", &controllerRequest{ControllerName: "projects", ItemID: "123", Action: "Stuff"})
	if method != "GetStuff" {
		t.Fatal("expected GetStuff method.  Actual: ", method)
	}
}

func TestGetMethodNameIndex(t *testing.T) {
	method := getMethodName("GET", &controllerRequest{})
	if method != "Index" {
		t.Fatal("expected Index method.  Actual: ", method)
	}
}

func TestGetMethodNameUpdate(t *testing.T) {
	method := getMethodName("PUT", &controllerRequest{})
	if method != "Put" {
		t.Fatal("expected Put method.  Actual: ", method)
	}
}

func TestGetMethod(t *testing.T) {
	router := NewControllerRoutingHandler()
	router.RegisterController("Test", &MockController{})
	method := router.getMethod("Test", "Index")
	retVal := method.Call([]reflect.Value{reflect.ValueOf(&controllerRequest{})})
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
	req := &controllerRequest{ItemID: "1234", Action: "Rawmethod"}
	method := router.getMethod("Test", "GetRawmethod")
	callRawMethod(req, method, writer, &http.Request{})
	if writer.Body.String() != "called raw GET method" {
		t.Fatal("expected to call raw method")
	}
}

func TestCallRawPostMethod(t *testing.T) {
	writer := httptest.NewRecorder()
	req := controllerRequest{}
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
	args := getRequestArguments("GET", &controllerRequest{}, "1234")
	if len(args) != 1 {
		t.Fatal("expected 1 arguments.  Actual: ", args)
	}
}

func TestGetArgumentsWithFilter(t *testing.T) {
	cr := &controllerRequest{ItemID: "1234"}
	args := getRequestArguments("GET", cr, nil)
	if len(args) != 1 || (args)[0].Interface() != cr {
		t.Fatal("expected 1 arguments with value cr.  Actual:", args)
	}
}

func TestGetArgumentsWithFilterAndQueryFilter(t *testing.T) {
	cr := &controllerRequest{ItemID: "1234", Action: "Stuff", ActionFilter: "4567"}
	args := getRequestArguments("GET", cr, nil)
	if len(args) != 1 || (args)[0].Interface() != cr {
		t.Fatal("expected 1 argument with value cr.  Actual:", args)
	}
}

func TestGetArgumentForDelete(t *testing.T) {
	cr := &controllerRequest{ItemID: "1234"}
	args := getRequestArguments("DELETE", cr, nil)
	if len(args) != 1 || (args)[0].Interface() != cr {
		t.Fatal("expected 1 argument with value cr.  Actual:", args)
	}
}

func TestGetArgumentForPut(t *testing.T) {
	cr := &controllerRequest{ItemID: "1234"}
	data := &SimpleData{Hello: "there"}
	args := getRequestArguments("PUT", cr, data)
	if len(args) != 2 || (args)[0].Interface() != cr || (args)[1].Interface() != data {
		t.Fatal("expected 2 arguments [cr, data].  Actual:", args)
	}
}

func TestGetArgumentForPost(t *testing.T) {
	args := getRequestArguments("POST", &controllerRequest{}, &SimpleData{Hello: "there"})
	if len(args) != 2 || (args)[1].Interface().(*SimpleData).Hello != "there" {
		t.Fatal("expected 2 argument with a map.  Actual:", args)
	}
}

func TestGetArgumentInvalid(t *testing.T) {
	args := getRequestArguments("BOGUS", &controllerRequest{}, &SimpleData{Hello: "there"})
	if len(args) != 1 {
		t.Fatal("expected 1 arguments.  Actual:", args)
	}
}

func TestCheckUrl(t *testing.T) {
	// success
	cr := &controllerRequest{ControllerName: "Tests", ItemID: "1234", ActionFilter: "Me"}
	err := checkUrl("PUT", "PutMe", cr)
	if err != nil {
		t.Error("Expected to have success")
	}

	// Get without id
	cr = &controllerRequest{ControllerName: "Tests", Action: "Me"}
	err = checkUrl("GET", "GetMe", cr)
	if err == nil || err.Error() != "Malformed URL. Expected: /Tests/{id}/Me/{optional filter}" {
		t.Error("Expected to have success", err)
	}

	// Post with ID
	cr = &controllerRequest{ControllerName: "Tests", ItemID: "1234"}
	err = checkUrl("POST", "Post", cr)
	if err == nil || err.Error() != "Malformed URL. Expected: /Tests" {
		t.Error("Expected to have Invalid Url error: ", err)
	}

	// Post without ID but with action
	cr = &controllerRequest{ControllerName: "Tests", Action: "Me"}
	err = checkUrl("POST", "PostMe", cr)
	if err == nil || err.Error() != "Malformed URL. Expected: /Tests/{id}/Me/{optional filter}" {
		t.Error("Expected to have Invalid Url error: ", err)
	}
}

func TestCallControllerMethod(t *testing.T) {
	method := reflect.ValueOf(&MockController{}).MethodByName("Index")
	result, _ := callControllerMethod(&method, []reflect.Value{reflect.ValueOf(&controllerRequest{})})
	if result != "called Index" {
		t.Fatal("Expected to receive back called message from method")
	}
}

func TestCallControllerMethodWithError(t *testing.T) {
	method := reflect.ValueOf(&MockController{}).MethodByName("GetError")
	_, err := callControllerMethod(&method, []reflect.Value{reflect.ValueOf(&controllerRequest{})})
	if err == nil || err.Error() != "failed" {
		t.Fatal("Expected error return from method", err)
	}
}

func TestHttpHandlerSuccess(t *testing.T) {
	rw := httptest.NewRecorder()
	getMockRouter().controllerRoutingHandler(rw, newHttpRequest("GET", "/projects", nil))
	if len(rw.HeaderMap) != 2 {
		t.Fatal("expected to succeed")
	}
}

func TestHttpHandlerMethodNotFound(t *testing.T) {
	rw := httptest.NewRecorder()
	getMockRouter().controllerRoutingHandler(rw, newHttpRequest("GET", "/projects/123/bogus", nil))
	if rw.Body.String() != "Method \"GetBogus\" not found\n" {
		t.Fatal("expected to be unable to find method", rw.Body.String())
	}
}

func TestHttpHandlerGetMethod(t *testing.T) {
	rw := httptest.NewRecorder()
	getMockRouter().controllerRoutingHandler(rw, newHttpRequest("GET", "/projects/123/method", nil))
	if body := rw.Body.String(); body != "called GetMethod" {
		t.Fatal("expected to be able to call GetMethod method", body)
	}
}

func TestHttpHandlerCallRaw(t *testing.T) {
	rw := httptest.NewRecorder()
	getMockRouter().controllerRoutingHandler(rw, newHttpRequest("GET", "/projects/123/Rawmethod", nil))
	if !strings.Contains(rw.Body.String(), "called raw GET method") {
		t.Fatal("expected called raw method message, Actual:", rw.Body.String())
	}
}

func TestHttpHandlerCantGetJson(t *testing.T) {
	rw := httptest.NewRecorder()
	getMockRouter().controllerRoutingHandler(rw, newHttpRequest("PUT", "/projects/123/valid", &MockErroringReadCloser{}))
	if !strings.Contains(rw.Body.String(), "Failed to read JSON data:") {
		t.Fatal("expected failure getting JSON", rw.Body.String())
	}
}

func TestHttpHandlerInvalidArguments(t *testing.T) {
	rw := httptest.NewRecorder()
	getMockRouter().controllerRoutingHandler(rw, newHttpRequest("PUT", "/projects", ioutil.NopCloser(bytes.NewBufferString(`{ "hello": "there" }`))))
	if body := rw.Body.String(); body != "Malformed URL. Expected: /Projects/{id}\n" {
		t.Fatal("should've gotten bogus arguments: ", body)
	}
}

func TestHttpHandlerErroringMethod(t *testing.T) {
	rw := httptest.NewRecorder()
	getMockRouter().controllerRoutingHandler(rw, newHttpRequest("GET", "/projects/1/error", nil))
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
