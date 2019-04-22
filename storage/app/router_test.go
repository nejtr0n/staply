package app

import (
	"bytes"
	"encoding/base64"
	json2 "encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func performRequest(r http.Handler, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestNotFounded(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter()

	req, _ := http.NewRequest("GET", "/blah", nil)
	resp := performRequest(router, req)
	assert.Equal(t, resp.Code, http.StatusNotFound)
}

func TestPing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter()

	req, _ := http.NewRequest("GET", "/storage/ping", nil)
	resp := performRequest(router, req)
	assert.Equal(t, resp.Code, http.StatusOK)
	assert.JSONEq(t, resp.Body.String(), `{"message":"pong"}`)
}

func TestUpload(t *testing.T) {
	cases := []struct {
		name string
		ext  string
		data string
		resp string
	}{
		{
			name: "test1.png",
			ext:  "image/png",
			data: "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",
			resp: `[{"name":"test1.png","path":"/images/test1.png","resize":"/images/thumb_test1.png"}]`,
		},
	}

	//gin.SetMode(gin.TestMode)
	router := NewRouter()
	for i, tc := range cases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			// form file
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, _ := writer.CreateFormFile("images[]", tc.name)
			content, _ := base64.StdEncoding.DecodeString(tc.data)
			io.Copy(part, bytes.NewReader(content))
			writer.Close()

			req, _ := http.NewRequest("POST", "/storage/upload", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			resp := performRequest(router, req)
			assert.Equal(t, resp.Code, http.StatusOK)
			assert.Equal(t, tc.resp, resp.Body.String())
		})
	}
}

func TestLink(t *testing.T) {
	cases := []struct {
		url  string
		resp string
	}{
		{
			"https://upload.wikimedia.org/wikipedia/commons/d/d9/Test.png",
			`[{"name":"Test.png","path":"/images/Test.png","resize":"/images/thumb_Test.png"}]`,
		},
	}

	router := NewRouter()
	for i, tc := range cases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			data := url.Values{
				"url": {tc.url},
			}
			req, _ := http.NewRequest("POST", "/storage/upload/link", strings.NewReader(data.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			resp := performRequest(router, req)
			assert.Equal(t, resp.Code, http.StatusOK)
			assert.Equal(t, tc.resp, resp.Body.String())
		})
	}
}

func TestJson(t *testing.T) {
	cases := []struct {
		Name    string `json:"name"`
		Ext     string `json:"type"`
		Size    int    `json:"size"`
		Content string `json:"content"`
		resp    string
	}{
		{
			Name:    "test.png",
			Size:    4862,
			Ext:     "image/png",
			Content: "data:image/png;base64,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",
			resp:    `[{"name":"test.png","path":"/images/test.png","resize":"/images/thumb_test.png"}]`,
		},
	}

	router := NewRouter()
	for i, tc := range cases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			var data []interface{}
			data = append(data, tc)
			jsonData, _ := json2.Marshal(data)

			req, _ := http.NewRequest("POST", "/storage/upload/json", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			resp := performRequest(router, req)
			assert.Equal(t, resp.Code, http.StatusOK)
			assert.Equal(t, tc.resp, resp.Body.String())
		})
	}
}

func TestErrorResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	errorResponse(c, "test")
	c.Request, _ = http.NewRequest("GET", "/fail", nil)
	resp := rec.Result()
	assert.Equal(t, resp.StatusCode, http.StatusBadRequest)
	b, _ := ioutil.ReadAll(resp.Body)
	assert.JSONEq(t, string(b), `{"message":"test"}`)
}
