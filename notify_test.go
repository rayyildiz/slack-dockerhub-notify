package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler(t *testing.T) {
	testCases := []struct {
		method     string
		url        string
		statusCode int
		expected   string
	}{
		{http.MethodGet, "/", http.StatusOK, "DOCTYPE html"},
		{http.MethodDelete, "/test", http.StatusInternalServerError, ""},
		{http.MethodPut, "/test", http.StatusInternalServerError, ""},
	}

	for _, tc := range testCases {

		req, err := http.NewRequest(tc.method, tc.url, nil)

		if err != nil {
			t.Errorf("error occured creating request, %v", err)
		}

		httpRecorder := httptest.NewRecorder()
		hh := http.HandlerFunc(handler)

		hh.ServeHTTP(httpRecorder, req)

		if status := httpRecorder.Code; status != tc.statusCode {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		if strings.Contains(httpRecorder.Body.String(), tc.expected) == false {
			t.Errorf("handler returned unexpected body: got %v want %v",
				httpRecorder.Body.String(), tc.expected)
		}
	}
}

func BenchmarkHomePageHandler(b *testing.B) {

	for n := 0; n < b.N; n++ {
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		httpRecorder := httptest.NewRecorder()
		hh := http.HandlerFunc(handler)

		hh.ServeHTTP(httpRecorder, req)
	}
}
