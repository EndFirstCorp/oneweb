package oneweb

import (
	"fmt"
	"testing"
)

type MockTestRunner struct {
	TestRunner
	Errors   []string
	Messages []string
}

func (r *MockTestRunner) Error(args ...interface{}) {
	r.Errors = append(r.Errors, fmt.Sprint(args...))
}

func (r *MockTestRunner) Logf(format string, args ...interface{}) {
	r.Messages = append(r.Messages, fmt.Sprintf(format, args...))
}

func TestAutoFuzzTestController(t *testing.T) {
	tester := &MockTestRunner{}
	AutoFuzzTestController(tester, &MockController{}) //
	if len(tester.Errors) != 5 || len(tester.Messages) != 13 {
		t.Error("Expected 4 errors and 12 return results")
	}
}

func TestFuzzTestController(t *testing.T) {
	results := FuzzTestController(&MockController{})
	if len(results) != 13 {
		t.Error("expected 13 controller methods tested")
	}
}

func TestFuzzTestControllerMethodIndex(t *testing.T) {
	result := fuzzTestControllerMethod(&MockController{}, "Index")
	if result.MethodName != "Index" || result.ValidationError != nil || result.ReturnData[0] != "called Index" || result.ReturnData[1] != nil {
		t.Error("Problems with Index", result)
	}
}

func TestFuzzTestControllerMethodGet(t *testing.T) {
	result := fuzzTestControllerMethod(&MockController{}, "Get")
	if result.MethodName != "Get" || result.ValidationError != nil || result.ReturnData[0] != "called Get" || result.ReturnData[1] != nil {
		t.Error("Problems with Get", result)
	}
}

func TestFuzzTestControllerMethodGetMethod(t *testing.T) {
	result := fuzzTestControllerMethod(&MockController{}, "GetMethod")
	if result.MethodName != "GetMethod" || result.ValidationError != nil || result.ReturnData[0] != "called GetMethod" || result.ReturnData[1] != nil {
		t.Error("Problems with Get", result)
	}
}

func TestFuzzTestControllerMethodGetError(t *testing.T) {
	result := fuzzTestControllerMethod(&MockController{}, "GetError")
	if result.MethodName != "GetError" || result.ValidationError != nil || result.ReturnData[0] != "called GetError" || result.ReturnData[1].(error).Error() != "failed" {
		t.Error("Problems with GetError", result)
	}
}

func TestFuzzTestControllerMethodGetWrongReturnType(t *testing.T) {
	result := fuzzTestControllerMethod(&MockController{}, "GetWrongReturnType")
	if result.MethodName != "GetWrongReturnType" || result.ValidationError.Error() != "Method \"GetWrongReturnType\" error: Unsupported return type.  Expected (string, error)" || len(result.ReturnData) != 0 {
		t.Error("Problems with GetWrongReturnType", result)
	}
}

func TestFuzzTestControllerMethodGetTooFewReturns(t *testing.T) {
	result := fuzzTestControllerMethod(&MockController{}, "GetTooFewReturns")
	if result.MethodName != "GetTooFewReturns" || result.ValidationError.Error() != "Method \"GetTooFewReturns\" error: Unsupported return type.  Expected (string, error)" || len(result.ReturnData) != 0 {
		t.Error("Problems with GetTooFewReturns", result)
	}
}

func TestFuzzTestControllerMethodPut(t *testing.T) {
	result := fuzzTestControllerMethod(&MockController{}, "Put")
	if result.MethodName != "Put" || result.ValidationError != nil || result.ReturnData[0] != "Called Put with value " {
		t.Error("Problems with Put", result)
	}
}

func TestFuzzTestControllerMethodPutValid(t *testing.T) {
	result := fuzzTestControllerMethod(&MockController{}, "PutValid")
	if result.MethodName != "PutValid" || result.ValidationError != nil || result.ReturnData[0] != "Called PutValid 0" {
		t.Error("Problems with PutValid", result)
	}
}

func TestFuzzTestControllerMethodPutBogus(t *testing.T) {
	result := fuzzTestControllerMethod(&MockController{}, "PutBogus")
	if result.MethodName != "PutBogus" || result.ValidationError.Error() != "Method \"PutBogus\" error: Requires 2 input args (cr *controllerRequest, json *YourStruct or []YourStruct)" || len(result.ReturnData) != 0 {
		t.Error("Problems with PutBogus", result)
	}
}

func TestFuzzTestControllerMethodGetRawmethod(t *testing.T) {
	result := fuzzTestControllerMethod(&MockController{}, "GetRawmethod")
	if result.MethodName != "GetRawmethod" || result.ValidationError != nil || result.ReturnData[0] != "called raw GET method" {
		t.Error("Problems with GetRawmethod", result)
	}
}

func TestFuzzTestControllerMethodPost(t *testing.T) {
	result := fuzzTestControllerMethod(&MockController{}, "Post")
	if result.MethodName != "Post" || result.ValidationError != nil || result.ReturnData[0] != "called raw POST method" {
		t.Error("Problems with Post", result)
	}
}
