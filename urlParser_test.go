package oneweb

import (
	"net/http"
	"net/url"
	"testing"
)

func TestParseUrl(t *testing.T) {
	r, _ := http.NewRequest("GET", "/members", nil)
	r.Header.Add("X-User-Id", "1")
	req := newControllerRequest(r)
	if req.ControllerName != "Members" || req.ItemID != "" || req.Action != "" || req.UserID != 1 || req.HalfAuthID != -1 {
		t.Fatal("expected controller Members with empty filter and Query.  Actual", req.ControllerName, req.ItemID, req.Action, req.UserID, req.HalfAuthID)
	}
}

func TestParseUrlMoreParts(t *testing.T) {
	req := newControllerRequest(&http.Request{URL: &url.URL{Path: "/members/23/doSomething"}})
	if req.ControllerName != "Members" || req.ItemID != "23" || req.Action != "Dosomething" {
		t.Fatal("expected controller Members filter 23 and Query Dosomething.  Actual", req.ControllerName, req.ItemID, req.Action)
	}
}

func TestParseUrlAllParts(t *testing.T) {
	req := newControllerRequest(&http.Request{URL: &url.URL{Path: "/members/23/doSomething/5"}})
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

func TestGetUserIds(t *testing.T) {
	r, _ := http.NewRequest("GET", "/url", nil)
	r.Header.Add("X-User-Id", "12")
	halfAuthID, userID := getUserIDs(r)
	if halfAuthID != -1 || userID != 12 {
		t.Error("expected valid user ids")
	}

	r.Header.Del("X-User-Id")
	r.Header.Add("X-Half-Auth-User-Id", "123")
	halfAuthID, userID = getUserIDs(r)
	if halfAuthID != 123 || userID != -1 {
		t.Error("expected valid user ids")
	}
}
