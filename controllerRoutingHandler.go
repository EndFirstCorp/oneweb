package oneweb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
)

type ControllerRoutingHandler struct {
	Controllers       map[string]interface{}
	controllerMethods map[string]*reflect.Value
}

func NewControllerRoutingHandler() *ControllerRoutingHandler {
	return &ControllerRoutingHandler{Controllers: make(map[string]interface{}), controllerMethods: make(map[string]*reflect.Value)}
}

func (c *ControllerRoutingHandler) RegisterController(name string, controller interface{}) error {
	c.Controllers[name] = controller
	return c.addValidControllerMethods(controller, name)
}

func (c *ControllerRoutingHandler) Handler() http.Handler {
	return http.HandlerFunc(c.controllerRoutingHandler)
}

func (c *ControllerRoutingHandler) controllerRoutingHandler(rw http.ResponseWriter, r *http.Request) {
	cr := NewControllerRequest(r)

	methodName := getMethodName(r.Method, cr)
	method := c.getMethod(cr.ControllerName, methodName)
	if method == nil {
		http.Error(rw, "Method \""+methodName+"\" not found", http.StatusInternalServerError)
		return
	}

	if isRawMethod(method.Type()) {
		callRawMethod(cr, method, rw, r)
		return
	}

	json, err := getJsonBody(r, method)
	if err != nil {
		http.Error(rw, "Failed to read JSON data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	arguments := getRequestArguments(r.Method, cr, json)
	err = checkRuntimeArguments(method, arguments, methodName, r.Method, cr)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	retVal, err := callControllerMethod(method, arguments)
	if err != nil {
		http.Error(rw, "Internal error calling controller method: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeResponse(rw, retVal)
}

func (c *ControllerRoutingHandler) addValidControllerMethods(controller interface{}, controllerName string) error {
	controllerValue := reflect.ValueOf(controller)
	numMethod := controllerValue.NumMethod()
	var errMsg string
	for i := 0; i < numMethod; i++ {
		method := controllerValue.Method(i)
		methodName := controllerValue.Type().Method(i).Name
		httpVerb, action, err := validateMethod(method, methodName)
		if err != nil {
			errMsg += err.Error() + "\n"
		} else {
			c.controllerMethods[strings.Title(strings.ToLower(controllerName))+strings.Title(strings.ToLower(httpVerb))+strings.Title(strings.ToLower(action))] = &method
		}
	}
	return errors.New(errMsg)
}

func writeResponse(rw http.ResponseWriter, json string) {
	rw.Header().Add("Access-Control-Allow-Origin", "*")
	rw.Header().Add("Content-Type", "application/json")
	fmt.Fprintf(rw, json)
}

func getMethodName(httpVerb string, cr *ControllerRequest) string {
	methodName := strings.Title(strings.ToLower(httpVerb))
	if methodName == "Get" && cr.ControllerFilter == "" {
		methodName = "Index"
	}

	if methodName != "Index" && cr.Action != "" {
		methodName = methodName + cr.Action
	}
	return methodName
}

func (c *ControllerRoutingHandler) getMethod(controllerName string, methodName string) *reflect.Value {
	return c.controllerMethods[controllerName+methodName]
}

func getJsonBody(r *http.Request, method *reflect.Value) (interface{}, error) {
	if r.Method == "POST" || r.Method == "PUT" {
		outType := method.Type().In(method.Type().NumIn() - 1)
		pointer := isPointer(outType)
		var data reflect.Value
		if pointer {
			data = reflect.New(outType.Elem()) // pointer of pointer doesn't work
		} else {
			data = reflect.New(outType)
		}
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return data.Interface(), err
		}
		err = json.Unmarshal(body, data.Interface())
		if pointer {
			return data.Interface(), err
		}
		return data.Elem().Interface(), err
	}
	return nil, nil
}

func getRequestArguments(httpVerb string, cr *ControllerRequest, json interface{}) []reflect.Value {
	var args []reflect.Value = nil
	if cr.ControllerFilter == "" {
		args = []reflect.Value{reflect.ValueOf(cr)}
	} else if cr.ActionFilter == "" {
		args = []reflect.Value{reflect.ValueOf(cr), reflect.ValueOf(cr.ControllerFilter)}
	} else {
		args = []reflect.Value{reflect.ValueOf(cr), reflect.ValueOf(cr.ControllerFilter), reflect.ValueOf(cr.ActionFilter)}
	}
	if httpVerb == "PUT" || httpVerb == "POST" {
		args = append(args, reflect.ValueOf(json))
	}
	return args
}

func callRawMethod(cr *ControllerRequest, method *reflect.Value, rw http.ResponseWriter, r *http.Request) {
	method.Call([]reflect.Value{reflect.ValueOf(cr), reflect.ValueOf(rw), reflect.ValueOf(r)})
}

func checkRuntimeArguments(method *reflect.Value, arguments []reflect.Value, methodName string, httpVerb string, cr *ControllerRequest) error {
	methodType := method.Type()
	numIn := methodType.NumIn()

	if len(arguments) != numIn {
		return errors.New("Invalid Url to call method: \"" + methodName + "\"")
	}
	return nil
}

/*func getValidUrl(numIn int, httpVerb string, cr *ControllerRequest) string {
	if numIn == 1 && httpVerb == "POST" {
		return cr.ControllerName
	} else if numIn == 1 && httpVerb != "POST" || numIn == 2 && httpVerb == "PUT" {
		return cr.ControllerName + "/{id}"
	} else {
		return cr.ControllerName + "/{id}/" + cr.Action + "/{filter}"
	}
}*/

func callControllerMethod(method *reflect.Value, arguments []reflect.Value) (string, error) {
	ret := method.Call(arguments)
	retVal := ret[0].String()
	retErr := ret[1].Interface()
	if retErr == nil {
		return retVal, nil
	}
	return retVal, retErr.(error)
}
