package oneweb

import (
	"encoding/json"
	"net/http"
	"strings"
)

type User struct {
	UserID   int
	Email    string
	FullName string
	JSON     string
}

type ControllerRequest struct {
	ControllerName string
	ItemID         string
	Action         string
	ActionFilter   string
	User           *User
	Headers        map[string]string
}

func newControllerRequest(r *http.Request) *ControllerRequest {
	headers := make(map[string]string)
	for key, value := range r.Header {
		if len(value) != 0 {
			headers[key] = value[0]
		}
	}

	urlPath := removeTrailingSlash(r.URL.Path)
	urlParams := strings.Split(urlPath, "/")
	controllerName := strings.Title(strings.ToLower(urlParams[1]))

	var action string
	var controllerFilter string
	var actionFilter string
	if len(urlParams) >= 3 {
		controllerFilter = urlParams[2]
	}
	if len(urlParams) >= 4 {
		action = urlParams[3]
		action = strings.Title(strings.ToLower(action))
	}
	if len(urlParams) >= 5 {
		actionFilter = urlParams[4]
	}

	userJSON := r.Header.Get("X-User")
	user := &User{}
	json.Unmarshal([]byte(userJSON), user)
	user.JSON = userJSON
	return &ControllerRequest{controllerName, controllerFilter, action, actionFilter, user, headers}
}

func removeTrailingSlash(urlPath string) string {
	if urlPath != "/" && urlPath[len(urlPath)-1:len(urlPath)] == "/" {
		urlPath = urlPath[0 : len(urlPath)-1]
	}
	return urlPath

}
