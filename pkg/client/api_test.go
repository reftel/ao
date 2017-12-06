package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func init() {
	logrus.SetLevel(logrus.FatalLevel)
}

func ReadTestFile(name string) []byte {
	filePath := fmt.Sprintf("./test_files/%s.json", name)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	return data
}

const affiliation = "paas"

func TestApiClient_Do(t *testing.T) {

	t.Run("Should include correct headers and path", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			assert.Len(t, req.Header, 5)

			agent := req.Header.Get("User-Agent")
			assert.Equal(t, "Go-http-client/1.1 ao/", agent)

			auth := req.Header.Get("Authorization")
			assert.Equal(t, "Bearer test", auth)

			accept := req.Header.Get("Accept")
			assert.Equal(t, "application/json", accept)

			contentType := req.Header.Get("Content-Type")
			assert.Equal(t, "application/json", contentType)

			assert.Equal(t, "/v1/hello", req.URL.Path)
		}))
		defer ts.Close()

		api := NewApiClient(ts.URL, "test", affiliation)
		api.Do(http.MethodGet, "/hello", nil)
	})

	t.Run("Should parse success Response struct correct", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			response := `{"success": true, "message": "OK", "items": [], "count": 0}`

			data := []byte(response)
			w.Write(data)
		}))
		defer ts.Close()

		api := NewApiClient(ts.URL, "test", affiliation)
		res, err := api.Do(http.MethodGet, "/", nil)

		assert.NoError(t, err)

		assert.Equal(t, true, res.Success)
		assert.Equal(t, "OK", res.Message)
		assert.Equal(t, json.RawMessage("[]"), res.Items)
		assert.Equal(t, 0, res.Count)
	})

	t.Run("Should fail when trying to connect to non existing host", func(t *testing.T) {
		api := NewApiClient("http://notvalid:8080", "", "")
		_, err := api.Do(http.MethodGet, "/", nil)
		assert.Error(t, err)
	})

	t.Run("Should send payload and retrieve response correctly", func(t *testing.T) {

		ac := NewAuroraConfig()
		assert.NotNil(t, ac.Files)
		assert.NotNil(t, ac.Versions)

		response := Response{
			Items: json.RawMessage(`[]`),
		}

		responseBody, err := json.Marshal(response)
		payload, err := json.Marshal(ac)
		assert.NoError(t, err)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			defer req.Body.Close()
			body, err := ioutil.ReadAll(req.Body)
			assert.NoError(t, err)

			assert.Equal(t, payload, body)
			w.Write(responseBody)
		}))
		defer ts.Close()

		api := NewApiClient(ts.URL, "", "")
		_, err = api.Do(http.MethodPut, "/", payload)
		assert.NoError(t, err)
	})

	t.Run("Should return error when status code is 403, 404, 500, 503", func(t *testing.T) {
		testCases := []struct {
			StatusCode int
			Message    string
			Path       string
		}{
			{http.StatusForbidden, `{"message": "Access denied", "path": "/"}`, "/"},
			{http.StatusNotFound, `{"message": "Not Found", "path": "/"}`, "/"},
			{http.StatusInternalServerError, `{"message": "Server error", "path": "/"}`, "/"},
			{http.StatusServiceUnavailable, `{"message": "Service unavailable", "path": "/"}`, "/"},
		}

		var testServers []*httptest.Server
		for _, test := range testCases {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(test.StatusCode)

				data := []byte(test.Message)
				w.Write(data)
			}))
			testServers = append(testServers, ts)

			api := NewApiClient(ts.URL, "test", affiliation)
			_, err := api.Do(http.MethodGet, test.Path, nil)

			assert.Error(t, err)
		}

		for _, ts := range testServers {
			ts.Close()
		}
	})
}