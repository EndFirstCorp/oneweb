package oneweb

import (
	"net/http"
	"strconv"
	"strings"
)

type ControllerRequest struct {
	ControllerName   string
	ControllerFilter string
	Action           string
	ActionFilter     string
	UserId           int
	HalfAuthId       int
}

func NewControllerRequest(r *http.Request) *ControllerRequest {
	halfAuthId, userId := getUserIds(r)
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
	return &ControllerRequest{controllerName, controllerFilter, action, actionFilter, userId, halfAuthId}
}

func removeTrailingSlash(urlPath string) string {
	if urlPath != "/" && urlPath[len(urlPath)-1:len(urlPath)] == "/" {
		urlPath = urlPath[0 : len(urlPath)-1]
	}
	return urlPath

}

func getUserIds(r *http.Request) (int, int) {
	var halfAuthId, userId int
	var err error
	if halfAuthId, err = strconv.Atoi(r.Header.Get("X-Half-Auth-User-Id")); err != nil {
		halfAuthId = -1
	}
	if userId, err = strconv.Atoi(r.Header.Get("X-User-Id")); err != nil {
		userId = -1
	}
	return halfAuthId, userId
}
