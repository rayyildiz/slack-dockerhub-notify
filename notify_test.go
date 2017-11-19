package notify

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
		{http.MethodGet, "/abcde", http.StatusOK, "UA-49404964-3"},
		{http.MethodDelete, "/test", http.StatusInternalServerError, ""},
		{http.MethodPut, "/test", http.StatusInternalServerError, ""},
	}

	for _, tc := range testCases {
		//log.Printf("%v", tc)

		req, err := http.NewRequest(tc.method, tc.url, nil)

		if err != nil {
			t.Errorf("error occured creating request, %v", err)
		}

		httpRecoreder := httptest.NewRecorder()
		hh := http.HandlerFunc(handler)

		hh.ServeHTTP(httpRecoreder, req)

		if status := httpRecoreder.Code; status != tc.statusCode {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		if strings.Contains(httpRecoreder.Body.String(), tc.expected) == false {
			t.Errorf("handler returned unexpected body: got %v want %v",
				httpRecoreder.Body.String(), tc.expected)
		}
	}
}

func BenchmarkHomePageHandler(b *testing.B) {

	for n := 0; n < b.N; n++ {
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		httpRecoreder := httptest.NewRecorder()
		hh := http.HandlerFunc(handler)

		hh.ServeHTTP(httpRecoreder, req)
	}
}
