package oneweb

import (
	"io"
	"net/http"
	"testing"
)

func TestParseUrl(t *testing.T) {
	req, _ := newControllerRequest(newHttpRequest("GET", "/members", nil))
	if req.ControllerName != "Members" || req.ItemID != "" || req.Action != "" || req.User == nil || req.User.Email != "test@test.com" {
		t.Fatal("expected controller Members with empty filter and Query.  Actual", req.ControllerName, req.ItemID, req.Action, req.User)
	}
}

func TestParseUrlMoreParts(t *testing.T) {
	req, _ := newControllerRequest(newHttpRequest("GET", "/members/23/doSomething", nil))
	if req.ControllerName != "Members" || req.ItemID != "23" || req.Action != "Dosomething" {
		t.Fatal("expected controller Members filter 23 and Query Dosomething.  Actual", req.ControllerName, req.ItemID, req.Action)
	}
}

func TestParseUrlAllParts(t *testing.T) {
	req, _ := newControllerRequest(newHttpRequest("GET", "/members/23/doSomething/5", nil))
	if req.ControllerName != "Members" || req.ItemID != "23" || req.Action != "Dosomething" || req.ActionFilter != "5" {
		t.Fatal("expected controller Members filter 23, Query Dosomething, QueryFilter 5.  Actual", req.ControllerName, req.ItemID, req.Action, req.ActionFilter)
	}
}

func TestRemoveTrailingSlashNothingToRemove(t *testing.T) {
	url := removeTrailingSlash("/members/23/doSomething")
	if url != "/members/23/doSomething" {
		t.Fatal("expected /members/23/doSomething")
	}
}

func TestRemoveTrailingSlash(t *testing.T) {
	url := removeTrailingSlash("/members/23/doSomething/")
	if url != "/members/23/doSomething" {
		t.Fatal("expected /members/23/doSomething")
	}
}

func newHttpRequest(method, url string, body io.ReadCloser) *http.Request {
	r, _ := http.NewRequest(method, url, nil)
	r.Header.Add("X-User", `{"Email":"test@test.com"}`)
	r.Body = body
	return r
}
