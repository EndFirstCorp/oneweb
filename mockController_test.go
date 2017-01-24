package oneweb

import (
	"fmt"
	"github.com/pkg/errors"
	"net/http"
)

type MockController struct {
}

func (c *MockController) Index(cr *controllerRequest) (string, error) {
	return "called Index", nil
}

func (c *MockController) Get(cr *controllerRequest) (string, error) {
	return "called Get", nil
}

func (c *MockController) GetMethod(cr *controllerRequest) (string, error) {
	return "called GetMethod", nil
}

func (c *MockController) GetError(cr *controllerRequest) (string, error) {
	return "called GetError", errors.New("failed")
}

func (c *MockController) GetWrongReturnType(cr *controllerRequest) (int, error) {
	return 1, nil
}

func (c *MockController) GetTooFewReturns(cr *controllerRequest) int {
	return 1
}

func (c *MockController) Put(cr *controllerRequest, data *SimpleData) (string, error) {
	return "Called Put with value " + data.Hello, nil
}

func (c *MockController) PutValid(cr *controllerRequest, data []SimpleData) (string, error) {
	return fmt.Sprintf("Called PutValid %v", len(data)), nil
}

func (c *MockController) PutBogus(cr *controllerRequest) (string, error) {
	return "", nil
}

func (c *MockController) GetBogus(id string) (string, error) {
	return "", nil
}

func (c *MockController) GetRawmethod(cr *controllerRequest, rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte("called raw GET method"))
}

func (c *MockController) Post(cr *controllerRequest, rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte("called raw POST method"))
}

func (c *MockController) Bogus(cr *controllerRequest, id string) (string, error) {
	return "", nil
}

func (c *MockController) privateMethod(cr *controllerRequest) (string, error) {
	return "called privateMethod", nil
}
