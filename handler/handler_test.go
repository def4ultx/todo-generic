package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(GetHandler))
	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Errorf("got err %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("want %v, got %v", http.StatusOK, resp.StatusCode)
	}
}

func TestHandlerRecorder(t *testing.T) {
	rr := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Errorf("got err %v", err)
	}

	GetHandler(rr, req)
	if rr.Result().StatusCode != http.StatusOK {
		t.Errorf("want %v, got %v", http.StatusOK, rr.Result().StatusCode)
	}
}
