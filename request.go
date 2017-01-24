package oneweb

import (
	"encoding/json"
	"github.com/pkg/errors"
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
}

func newControllerRequest(r *http.Request) (*ControllerRequest, error) {
	userJSON := r.Header.Get("X-User")
	user := &User{}
	err := json.Unmarshal([]byte(userJSON), user)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to retrieve user information. Unauthorized")
	}
	user.JSON = userJSON
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
	return &ControllerRequest{controllerName, controllerFilter, action, actionFilter, user}, nil
}

func removeTrailingSlash(urlPath string) string {
	if urlPath != "/" && urlPath[len(urlPath)-1:len(urlPath)] == "/" {
		urlPath = urlPath[0 : len(urlPath)-1]
	}
	return urlPath

}
