package oneweb

import (
	"errors"
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
		return httpVerb, action, errors.New("Method \"" + methodName + "\" error: Unsupported http verb: \"" + httpVerb + "\"")
	}

	if !method.IsValid() {
		return httpVerb, action, errors.New("Method \"" + methodName + "\" error: Internal error validating method")
	}

	methodType := method.Type()
	if isRawMethod(methodType) {
		return httpVerb, action, nil
	}

	if !isJSONReturnArgs(methodType) {
		return httpVerb, action, errors.New("Method \"" + methodName + "\" error: Unsupported return type.  Expected (string, error)")
	}

	numIn := methodType.NumIn()
	if numIn >= 5 {
		return httpVerb, action, errors.New("Method \"" + methodName + "\" error: Invalid number of input arguments.  Expected 4 or fewer")
	}

	validID := isValidID(methodType, numIn, httpVerb)
	validAction := isValidAction(methodType, numIn, httpVerb)

	if action != "" && !validAction {
		if httpVerb == "Post" || httpVerb == "Put" {
			return httpVerb, action, errors.New("Method \"" + methodName + "\" error: Requires either 3 or 4 input args (cr *ControllerRequest, id string, actionFilter string [optional], json *YourStruct or []YourStruct)")
		} else if httpVerb == "Get" || httpVerb == "Delete" {
			return httpVerb, action, errors.New("Method \"" + methodName + "\" error: Requires either 2 or 3 input args (cr *ControllerRequest, id string, actionFilter string [optional])")
		}
	} else if action == "" {
		if httpVerb == "Post" && numIn == 2 && !isJSONReceiverArg(methodType.In(0)) {
			return httpVerb, action, errors.New("Method \"" + methodName + "\" error: Requires 2 input arg (cr *ControllerRequest, json *YourStruct or []YourStruct)")
		} else if httpVerb == "Post" && numIn == 3 && !validID {
			return httpVerb, action, errors.New("Method \"" + methodName + "\" error: Requires 3 input args (cr *ControllerRequest, id string, json *YourStruct or []YourStruct)")
		} else if (httpVerb == "Get" || httpVerb == "Delete") && !validID {
			return httpVerb, action, errors.New("Method \"" + methodName + "\" error: Requires 2 input arg (cr *ControllerRequest, id string)")
		} else if httpVerb == "Put" && !validID {
			return httpVerb, action, errors.New("Method \"" + methodName + "\" error: Requires 3 input args (cr *ControllerRequest, id string, json *YourStruct or []YourStruct))")
		} else if httpVerb == "Index" && numIn != 1 {
			return httpVerb, action, errors.New("Method \"Index\" requires 0 input args")
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

func isValidID(methodType reflect.Type, numIn int, httpVerb string) bool {
	// if there is only an id then you need 2 arg + json
	return numIn > 1 && isStringArg(methodType.In(1)) &&
		(httpVerb == "Get" || httpVerb == "Delete") ||
		numIn == 3 && httpVerb == "Put" && isJSONReceiverArg(methodType.In(2))
}

func isValidAction(methodType reflect.Type, numIn int, httpVerb string) bool {
	// if there is a controller action then you need either 2 or 3 args + json
	return numIn > 1 && isStringArg(methodType.In(1)) &&
		(numIn == 2 && (httpVerb == "Get" || httpVerb == "Delete") ||
			numIn == 3 && (httpVerb == "Get" || httpVerb == "Delete") && isStringArg(methodType.In(2)) ||
			numIn == 3 && (httpVerb == "Post" || httpVerb == "Put") && isJSONReceiverArg(methodType.In(2)) ||
			numIn == 4 && (httpVerb == "Post" || httpVerb == "Put") && isStringArg(methodType.In(2)) && isJSONReceiverArg(methodType.In(3)))
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

func isJSONReceiverArg(argType reflect.Type) bool {
	return isPointer(argType) || isSlice(argType)
}
