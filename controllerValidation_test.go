package oneweb

import (
	"testing"
)

func TestCheckInputArgumentsWithWrongNumberOfArguments(t *testing.T) {
	/*
	   	method := reflect.ValueOf(&MockController{}).MethodByName("Index")
	   	err := checkInputArguments(method, []reflect.Value{reflect.ValueOf("1234")})
	   	if err == nil {
	   		t.Fatal("Expected to error on invalid arguments")
	   	}
	   }

	   func TestCheckInputArgumentsWrongTypeOfArguments(t *testing.T) {
	   	method := reflect.ValueOf(&MockController{}).MethodByName("GetError")
	   	err := checkInputArguments(method, []reflect.Value{reflect.ValueOf(1234)})
	   	if err == nil || !strings.Contains(err.Error(), "argument(0) type mismatch") {
	   		t.Fatal("Expected to error on input argument type mismatch", err)
	   	}
	   }

	   func TestCheckArgumentsWrongReturnType(t *testing.T) {
	   	method := reflect.ValueOf(&MockController{}).MethodByName("GetTooFewReturns")
	   	err := checkMethodOutputArguments(method.Type())
	   	if err == nil || !strings.Contains(err.Error(), "Expected 2 return variables (string, error), actual was 1") {
	   		t.Fatal("Expected to error on return type mismatch", err)
	   	}
	*/
}
