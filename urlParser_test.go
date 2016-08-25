package oneweb

import (
	"net/http"
	"net/url"
	"testing"
)

func TestParseUrl(t *testing.T) {
	r, _ := http.NewRequest("GET", "/members", nil)
	r.Header.Add("X-User-Id", "1")
	parsedUrl := NewControllerRequest(r)
	if parsedUrl.ControllerName != "Members" || parsedUrl.ControllerFilter != "" || parsedUrl.Action != "" || parsedUrl.UserId != 1 || parsedUrl.HalfAuthId != -1 {
		t.Fatal("expected controller Members with empty filter and Query.  Actual", parsedUrl.ControllerName, parsedUrl.ControllerFilter, parsedUrl.Action, parsedUrl.UserId, parsedUrl.HalfAuthId)
	}
}

func TestParseUrlMoreParts(t *testing.T) {
	parsedUrl := NewControllerRequest(&http.Request{URL: &url.URL{Path: "/members/23/doSomething"}})
	if parsedUrl.ControllerName != "Members" || parsedUrl.ControllerFilter != "23" || parsedUrl.Action != "Dosomething" {
		t.Fatal("expected controller Members filter 23 and Query Dosomething.  Actual", parsedUrl.ControllerName, parsedUrl.ControllerFilter, parsedUrl.Action)
	}
}

func TestParseUrlAllParts(t *testing.T) {
	parsedUrl := NewControllerRequest(&http.Request{URL: &url.URL{Path: "/members/23/doSomething/5"}})
	if parsedUrl.ControllerName != "Members" || parsedUrl.ControllerFilter != "23" || parsedUrl.Action != "Dosomething" || parsedUrl.ActionFilter != "5" {
		t.Fatal("expected controller Members filter 23, Query Dosomething, QueryFilter 5.  Actual", parsedUrl.ControllerName, parsedUrl.ControllerFilter, parsedUrl.Action, parsedUrl.ActionFilter)
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
	halfAuthid, userId := getUserIds(r)
	if halfAuthid != -1 || userId != 12 {
		t.Error("expected valid user ids")
	}

	r.Header.Del("X-User-Id")
	r.Header.Add("X-Half-Auth-User-Id", "123")
	halfAuthid, userId = getUserIds(r)
	if halfAuthid != 123 || userId != -1 {
		t.Error("expected valid user ids")
	}
}
