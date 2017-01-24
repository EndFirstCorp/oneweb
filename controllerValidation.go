package oneweb

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

func isRawMethod(methodType reflect.Type) bool {
	numOut := methodType.NumOut()
	numIn := methodType.NumIn()
	return numIn == 3 && numOut == 0 &&
		methodType.In(0) == reflect.TypeOf((*ControllerRequest)(nil)) &&
		methodType.In(1) == reflect.TypeOf((*http.ResponseWriter)(nil)).Elem() &&
		methodType.In(2) == reflect.TypeOf((*http.Request)(nil))
}

func validateMethod(method reflect.Value, methodName string) (httpVerb string, action string, err error) {
	httpVerb, action = parseMethod(methodName)
	if httpVerb == "" {
		return httpVerb, action, fmt.Errorf("Method \"%s\" error: Unsupported http verb: \"%s\"", methodName, httpVerb)
	}

	if !method.IsValid() {
		return httpVerb, action, fmt.Errorf("Method \"%s\" error: Internal error validating method", methodName)
	}

	methodType := method.Type()
	if isRawMethod(methodType) {
		return httpVerb, action, nil
	}

	if !isJSONReturnArgs(methodType) {
		return httpVerb, action, fmt.Errorf("Method \"%s\" error: Unsupported return type.  Expected (string, error)", methodName)
	}

	numIn := methodType.NumIn()
	switch httpVerb {
	case "Index", "Get", "Delete":
		if numIn != 1 || (numIn == 1 && !isControllerRequestArg(methodType.In(0))) { // only ControllerRequest
			return httpVerb, action, fmt.Errorf("Method \"%s\" error: Requires 1 input arg (cr *ControllerRequest)", methodName)
		}
	case "Post", "Put":
		if numIn != 2 || (numIn == 2 && (!isControllerRequestArg(methodType.In(0)) || !isJSONReceiverArg(methodType.In(1)))) {
			return httpVerb, action, fmt.Errorf("Method \"%s\" error: Requires 2 input args (cr *ControllerRequest, json *YourStruct or []YourStruct)", methodName)
		}
	}

	return httpVerb, action, nil
}

func parseMethod(methodName string) (string, string) {
	methodName = strings.Title(strings.ToLower(methodName))
	for _, prefix := range []string{"Index", "Get", "Put", "Post", "Delete"} {
		if strings.Index(methodName, prefix) == 0 {
			return prefix, strings.Title(strings.ToLower(methodName[len(prefix):len(methodName)]))
		}
	}
	return "", ""
}

func isJSONReturnArgs(methodType reflect.Type) bool {
	return methodType.NumOut() == 2 && isStringArg(methodType.Out(0)) && isErrorArg(methodType.Out(1))
}

func isErrorArg(argType reflect.Type) bool {
	return argType == reflect.TypeOf((*error)(nil)).Elem()
}

func isStringArg(argType reflect.Type) bool {
	return argType.Kind() == reflect.String
}

func isPointer(item reflect.Type) bool {
	return item.Kind() == reflect.Ptr
}

func isSlice(item reflect.Type) bool {
	return item.Kind() == reflect.Slice
}

func isControllerRequestArg(argType reflect.Type) bool {
	return argType == reflect.TypeOf(&ControllerRequest{})
}

func isJSONReceiverArg(argType reflect.Type) bool {
	return isPointer(argType) || isSlice(argType)
}
