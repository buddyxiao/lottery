package main

import (
	"fmt"
	"github.com/goccy/go-json"
	"io"
	"net/http"
	"net/url"
	"sync"
	"testing"
)

var localhostUrl = "http://localhost:8080"

func TestMVC(t *testing.T) {
	client := http.Client{}
	var msg = map[string]string{}
	resp := get(t, client, "/")
	body(t, resp.Body, &msg)
	expect(t, msg["msg"], "当前总共参与抽奖的用户数: 0")
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			reqData := url.Values{
				"users": {fmt.Sprintf("test_u%d", i)},
			}
			response, _ := client.PostForm(localhostUrl+"/import", reqData)
			expect(t, response.StatusCode, http.StatusOK)
		}()
	}
	wg.Wait()
	resp = get(t, client, "/")
	body(t, resp.Body, &msg)
	expect(t, msg["msg"], fmt.Sprintf("当前总共参与抽奖的用户数: 1000"))
	resp = get(t, client, "/lucky")
	resp = get(t, client, "/")
	body(t, resp.Body, &msg)
	expect(t, msg["msg"], fmt.Sprintf("当前总共参与抽奖的用户数: 999"))
}

func get(t *testing.T, client http.Client, path string) *http.Response {
	resp, err := client.Get(localhostUrl + path)
	if err != nil {
		t.Error(err)
	}
	return resp
}
func body(t *testing.T, reader io.Reader, data any) {
	err := json.NewDecoder(reader).Decode(data)
	if err != nil {
		t.Error(err)
	}
}

func expect(t *testing.T, real any, expect any) {
	t.Helper()
	if real != expect {
		t.Fail()
		t.Error(fmt.Sprintf("real: %v, expect: %v", real, expect))
	}
}
