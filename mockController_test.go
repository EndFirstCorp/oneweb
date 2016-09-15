package oneweb

import (
	"errors"
	"fmt"
	"net/http"
)

type MockController struct {
}

func (c *MockController) Index(cr *ControllerRequest) (string, error) {
	return "called Index", nil
}

func (c *MockController) Get(cr *ControllerRequest) (string, error) {
	return "called Get", nil
}

func (c *MockController) GetMethod(cr *ControllerRequest) (string, error) {
	return "called GetMethod", nil
}

func (c *MockController) GetError(cr *ControllerRequest) (string, error) {
	return "called GetError", errors.New("failed")
}

func (c *MockController) GetWrongReturnType(cr *ControllerRequest) (int, error) {
	return 1, nil
}

func (c *MockController) GetTooFewReturns(cr *ControllerRequest) int {
	return 1
}

func (c *MockController) Put(cr *ControllerRequest, data *SimpleData) (string, error) {
	return "Called Put with value " + data.Hello, nil
}

func (c *MockController) PutValid(cr *ControllerRequest, data []SimpleData) (string, error) {
	return fmt.Sprintf("Called PutValid %v", len(data)), nil
}

func (c *MockController) PutBogus(cr *ControllerRequest) (string, error) {
	return "", nil
}

func (c *MockController) GetBogus(id string) (string, error) {
	return "", nil
}

func (c *MockController) GetRawmethod(cr *ControllerRequest, rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte("called raw GET method"))
}

func (c *MockController) Post(cr *ControllerRequest, rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte("called raw POST method"))
}

func (c *MockController) Bogus(cr *ControllerRequest, id string) (string, error) {
	return "", nil
}
