package oneweb

import (
	"net/http"
	"strconv"
	"strings"
)

type ControllerRequest struct {
	ControllerName string
	ItemID         string
	Action         string
	ActionFilter   string
	UserID         int
	HalfAuthID     int
}

func newControllerRequest(r *http.Request) *ControllerRequest {
	halfAuthID, userID := getUserIDs(r)
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
	return &ControllerRequest{controllerName, controllerFilter, action, actionFilter, userID, halfAuthID}
}

func removeTrailingSlash(urlPath string) string {
	if urlPath != "/" && urlPath[len(urlPath)-1:len(urlPath)] == "/" {
		urlPath = urlPath[0 : len(urlPath)-1]
	}
	return urlPath

}

func getUserIDs(r *http.Request) (int, int) {
	var halfAuthID, userID int
	var err error
	if halfAuthID, err = strconv.Atoi(r.Header.Get("X-Half-Auth-User-Id")); err != nil {
		halfAuthID = -1
	}
	if userID, err = strconv.Atoi(r.Header.Get("X-User-Id")); err != nil {
		userID = -1
	}
	return halfAuthID, userID
}
