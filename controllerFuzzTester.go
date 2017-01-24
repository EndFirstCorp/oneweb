package oneweb

import (
	"net/http"
	"net/http/httptest"
	"reflect"
)

type TestRunner interface {
	Error(args ...interface{})
	Logf(format string, args ...interface{})
}

type MethodTestResult struct {
	MethodName               string
	ValidationError          error
	HasInvalidSQLQueryParams bool
	ReturnData               []interface{}
}

func fuzzTestControllerMethod(controller interface{}, methodName string) MethodTestResult {
	return testControllerMethod(reflect.ValueOf(controller), methodName)
}

func AutoFuzzTestController(t TestRunner, controller interface{}) {
	results := FuzzTestController(controller)
	for _, method := range results {
		if method.ValidationError != nil {
			t.Error(method.ValidationError)
		}
		t.Logf("Method \"%v\" returned: %v", method.MethodName, method.ReturnData)
	}
}

func FuzzTestController(controller interface{}) []MethodTestResult {
	controllerValue := reflect.ValueOf(controller)
	numMethod := controllerValue.NumMethod()
	testResults := make([]MethodTestResult, numMethod, numMethod)
	for i := 0; i < numMethod; i++ {
		testResults[i] = testControllerMethod(controllerValue, controllerValue.Type().Method(i).Name)
	}
	return testResults
}

func testControllerMethod(controllerValue reflect.Value, methodName string) MethodTestResult {
	var retVal []reflect.Value
	_, _, err := validateMethod(controllerValue.MethodByName(methodName), methodName)
	if err == nil {
		retVal = callMethod(controllerValue.MethodByName(methodName))
	}
	return MethodTestResult{
		MethodName:      methodName,
		ValidationError: err,
		ReturnData:      getReturnValues(retVal),
	}
}

func callMethod(method reflect.Value) []reflect.Value {
	if isRawMethod(method.Type()) {
		writer := httptest.NewRecorder()
		args := []reflect.Value{reflect.ValueOf(&controllerRequest{}), reflect.ValueOf(writer), reflect.ValueOf(&http.Request{})}
		method.Call(args)
		return []reflect.Value{reflect.ValueOf(writer.Body.String())}
	}
	args := getArgs(method)
	return method.Call(args)
}

func getArgs(method reflect.Value) []reflect.Value {
	methodType := method.Type()
	numArgs := methodType.NumIn()
	args := make([]reflect.Value, numArgs, numArgs)
	for i := 0; i < numArgs; i++ {
		switch methodType.In(i).Kind() {
		case reflect.Ptr:
			myType := methodType.In(i)
			item := reflect.New(myType.Elem())
			args[i] = item
		default:
			myType := methodType.In(i)
			item := reflect.New(myType)
			args[i] = item.Elem()
		}
	}
	return args
}

func getReturnValues(retVals []reflect.Value) []interface{} {
	output := make([]interface{}, len(retVals), len(retVals))
	for i, item := range retVals {
		output[i] = item.Interface()
	}
	return output
}
