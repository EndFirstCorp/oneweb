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
	cr := newControllerRequest(r)
	methodName := getMethodName(r.Method, cr)
	err := checkUrl(r.Method, methodName, cr)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	method := c.getMethod(cr.ControllerName, methodName)
	if method == nil {
		http.Error(rw, "Method \""+methodName+"\" not found", http.StatusInternalServerError)
		return
	}

	if isRawMethod(method.Type()) {
		callRawMethod(cr, method, rw, r)
		return
	}

	json, err := getJSONBody(r, method)
	if err != nil {
		http.Error(rw, "Failed to read JSON data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	arguments := getRequestArguments(r.Method, cr, json)
	retVal, err := callControllerMethod(method, arguments)
	if err != nil {
		http.Error(rw, "Internal error calling controller method: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeResponse(rw, retVal)
}

func (c *ControllerRoutingHandler) addValidControllerMethods(controller interface{}, controllerName string) error {
	controllerValue := reflect.ValueOf(controller)
	controllerType := controllerValue.Type()
	numMethod := controllerValue.NumMethod()
	var errMsg string
	for i := 0; i < numMethod; i++ {
		methodName := controllerType.Method(i).Name
		if strings.ToLower(methodName[:1]) == methodName[:1] { // private method (lowercase first letter), so skip
			continue
		}
		method := controllerValue.Method(i)
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

func checkUrl(httpVerb, methodName string, cr *ControllerRequest) error {
	if methodName == "Index" {
		return nil
	}
	switch httpVerb {
	case "GET", "DELETE", "PUT": // always expect the id (controllerFilter) to be present
		if cr.ItemID == "" && cr.Action == "" {
			return fmt.Errorf("Malformed URL. Expected: /%s/{id}", cr.ControllerName)
		} else if cr.ItemID == "" {
			return fmt.Errorf("Malformed URL. Expected: /%s/{id}/%s/{optional filter}", cr.ControllerName, cr.Action)
		}
	case "POST":
		if cr.ItemID != "" && cr.Action == "" {
			return fmt.Errorf("Malformed URL. Expected: /%s", cr.ControllerName)
		} else if cr.ItemID == "" && cr.Action != "" {
			return fmt.Errorf("Malformed URL. Expected: /%s/{id}/%s/{optional filter}", cr.ControllerName, cr.Action)
		}
	}
	return nil
}

func getMethodName(httpVerb string, cr *ControllerRequest) string {
	methodName := strings.Title(strings.ToLower(httpVerb))
	if methodName == "Get" && cr.ItemID == "" && cr.Action == "" && cr.ActionFilter == "" {
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

func getJSONBody(r *http.Request, method *reflect.Value) (interface{}, error) {
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
	args := []reflect.Value{reflect.ValueOf(cr)}
	if httpVerb == "PUT" || httpVerb == "POST" {
		args = append(args, reflect.ValueOf(json))
	}
	return args
}

func callRawMethod(cr *ControllerRequest, method *reflect.Value, rw http.ResponseWriter, r *http.Request) {
	method.Call([]reflect.Value{reflect.ValueOf(cr), reflect.ValueOf(rw), reflect.ValueOf(r)})
}

func callControllerMethod(method *reflect.Value, arguments []reflect.Value) (string, error) {
	ret := method.Call(arguments)
	retVal := ret[0].String()
	retErr := ret[1].Interface()
	if retErr == nil {
		return retVal, nil
	}
	return retVal, retErr.(error)
}
