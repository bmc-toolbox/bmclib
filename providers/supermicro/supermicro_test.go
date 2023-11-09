package supermicro

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
)

func TestParseToken(t *testing.T) {
	testcases := []struct {
		name        string
		body        []byte
		expectToken string
	}{
		{
			"token with key type 1 found",
			[]byte(`<script>SmcCsrfInsert ("CSRF-TOKEN", "A0v9gild518yF36XZ6jqNZNsOUrHiEpkvM+QHKKVTFw");/*SmcCsrfInsert ("CSRF_TOKEN", "A0v9gild518yF36XZ6jqNZNsOUrHiEpkvM+QHKKVTFw");*/</script></body>`),
			"A0v9gild518yF36XZ6jqNZNsOUrHiEpkvM+QHKKVTFw",
		},
		{
			"token with key type 2 found",
			[]byte(`                </td>
            </tr>
        </table>
        </div>
    <script>SmcCsrfInsert ("CSRF_TOKEN", "Te6xHPx3NDhDmL4T21cZ/tXWbzatZQ3JHbQUCF5Hkns");</script></body>
</html>
`),
			"Te6xHPx3NDhDmL4T21cZ/tXWbzatZQ3JHbQUCF5Hkns",
		},
		{
			"token with key type 3 found",
			[]byte(`</script>
			<CSRF_TOKEN>
			<input type="hidden" name="initName" id="initName"/>
			<div id="refreshTag"></div>
			<script>SmcCsrfInsert ("CSRF_TOKEN", "fYQ/Xhd1AvA+kP/bM/tO5mhOzv4eM5evCOH/YSuBN70");</script></body>
			</html>`),
			"fYQ/Xhd1AvA+kP/bM/tO5mhOzv4eM5evCOH/YSuBN70",
		},
		{
			"token with key type 4 found",
			[]byte(`<script>SmcCsrfInsert ("CSRF_TOKEN", "RYjdEjWIhU+PCRFMBP2ZRPPePcQ4n3dM3s+rCgTnBBU");</script></body>`),
			"RYjdEjWIhU+PCRFMBP2ZRPPePcQ4n3dM3s+rCgTnBBU",
		},
		{
			"token with key type 5 found",
			[]byte(`<script>SmcCsrfInsert ("CSRF-TOKEN", "RYjdEjWIhU+PCRFMBP2ZRPPePcQ4n3dM3s+rCgTnBBU");</script></body>`),
			"RYjdEjWIhU+PCRFMBP2ZRPPePcQ4n3dM3s+rCgTnBBU",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gotToken := parseToken(tc.body)
			assert.Equal(t, tc.expectToken, gotToken)

		})
	}
}

func TestOpen(t *testing.T) {
	type handlerFuncMap map[string]func(http.ResponseWriter, *http.Request)
	testcases := []struct {
		name           string
		errorContains  string
		user           string
		pass           string
		handlerFuncMap handlerFuncMap
	}{
		{
			"happy path",
			"",
			"foo",
			"bar",
			handlerFuncMap{
				"/": func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				},
				// first request to login
				"/cgi/login.cgi": func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, r.Method, http.MethodPost)
					assert.Equal(t, r.Header.Get("Content-Type"), "application/x-www-form-urlencoded")

					b, err := io.ReadAll(r.Body)
					if err != nil {
						t.Fatal(err)
					}

					assert.Equal(t, `name=Zm9v&pwd=YmFy&check=00`, string(b))

					response := []byte(`<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
				<html xmlns="http://www.w3.org/1999/xhtml">
				<head>
					<META http-equiv="Content-Type" content="text/html; charset=utf-8">
					<META HTTP-EQUIV="Pragma" CONTENT="no_cache">
					<META NAME="ATEN International Co Ltd." CONTENT="(c) ATEN International Co Ltd. 2010">
					<title></title>
					<script language="JavaScript" type="text/javascript">
				<!--
					self.location = "../cgi/url_redirect.cgi?url_name=mainmenu";
				-->
					</script>
				</head>
				<body>
				</body>
				</html>`)
					_, _ = w.Write(response)
				},

				// second request for the csrf token
				"/cgi/url_redirect.cgi": func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, r.Method, http.MethodGet)
					assert.Equal(t, "url_name=topmenu", r.URL.RawQuery)
					_, err := io.ReadAll(r.Body)
					if err != nil {
						t.Fatal(err)
					}

					response := []byte(`<script>SmcCsrfInsert ("CSRF-TOKEN", "A0v9gild518yF36XZ6jqNZNsOUrHiEpkvM+QHKKVTFw");/*SmcCsrfInsert ("CSRF_TOKEN", "A0v9gild518yF36XZ6jqNZNsOUrHiEpkvM+QHKKVTFw");*/</script></body>`)
					_, _ = w.Write(response)
				},
			},
		},
		{
			"login error",
			"401: failed to login",
			"foo",
			"bar",
			handlerFuncMap{
				"/cgi/login.cgi": func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, r.Method, http.MethodPost)
					assert.Equal(t, r.Header.Get("Content-Type"), "application/x-www-form-urlencoded")

					b, err := io.ReadAll(r.Body)
					if err != nil {
						t.Fatal(err)
					}

					assert.Equal(t, `name=Zm9v&pwd=YmFy&check=00`, string(b))

					response := []byte(`barf`)
					w.WriteHeader(401)
					_, _ = w.Write(response)
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mux := http.NewServeMux()
			for endpoint, handler := range tc.handlerFuncMap {
				mux.HandleFunc(endpoint, handler)
			}

			server := httptest.NewTLSServer(mux)
			defer server.Close()

			server.Config.ErrorLog = log.Default()
			parsedURL, err := url.Parse(server.URL)
			if err != nil {
				t.Fatal(err)
			}

			client := NewClient(parsedURL.Hostname(), tc.user, tc.pass, logr.Discard(), WithPort(parsedURL.Port()))
			if err := client.Open(context.Background()); err != nil {
				if tc.errorContains != "" {
					assert.ErrorContains(t, err, tc.errorContains)

					return
				}

				assert.Nil(t, err)
			}
		})
	}

}

func Test_Close(t *testing.T) {
	testcases := []struct {
		name          string
		errorContains string
		user          string
		pass          string
		endpoint      string
		handler       func(http.ResponseWriter, *http.Request)
	}{
		{
			"happy path",
			"",
			"foo",
			"bar",
			"/cgi/logout.cgi",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, http.MethodGet)

				_, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}

				w.WriteHeader(200)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc(tc.endpoint, tc.handler)

			server := httptest.NewTLSServer(mux)
			defer server.Close()

			parsedURL, err := url.Parse(server.URL)
			if err != nil {
				t.Fatal(err)
			}

			client := NewClient(parsedURL.Hostname(), tc.user, tc.pass, logr.Discard(), WithPort(parsedURL.Port()))
			if err := client.Close(context.Background()); err != nil {
				if tc.errorContains != "" {
					assert.ErrorContains(t, err, tc.errorContains)

					return
				}

				assert.Nil(t, err)
			}
		})
	}

}

func Test_initScreenPreview(t *testing.T) {
	testcases := []struct {
		name          string
		errorContains string
		endpoint      string
		handler       func(http.ResponseWriter, *http.Request)
	}{
		{
			"happy path",
			"",
			"/cgi/op.cgi",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, http.MethodPost)
				assert.Equal(t, r.Header.Get("Content-Type"), "application/x-www-form-urlencoded")

				b, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, `op=sys_preview&_=`, string(b))

				_, _ = w.Write([]byte(`<?xml version="1.0" ?>
				<IPMI>
				</IPMI>`))
			},
		},
		{
			"error returned",
			"400",
			"/cgi/op.cgi",
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(400)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc(tc.endpoint, tc.handler)

			server := httptest.NewTLSServer(mux)
			defer server.Close()

			parsedURL, err := url.Parse(server.URL)
			if err != nil {
				t.Fatal(err)
			}

			client := NewClient(parsedURL.Hostname(), "foo", "bar", logr.Discard(), WithPort(parsedURL.Port()))
			if err := client.initScreenPreview(context.Background()); err != nil {
				if tc.errorContains != "" {
					assert.ErrorContains(t, err, tc.errorContains)

					return
				}

				assert.Nil(t, err)
			}
		})
	}
}

func Test_fetchScreenPreview(t *testing.T) {
	testcases := []struct {
		name          string
		expectImage   []byte
		errorContains string
		endpoint      string
		handler       func(http.ResponseWriter, *http.Request)
	}{
		{
			"happy path",
			[]byte(`fake image is fake`),
			"",
			"/cgi/url_redirect.cgi",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, http.MethodGet)
				assert.Equal(t, r.Header.Get("Content-Type"), "application/x-www-form-urlencoded")
				assert.Equal(t, "url_name=Snapshot&url_type=img", r.URL.RawQuery)

				_, _ = w.Write([]byte(`fake image is fake`))
			},
		},
		{
			"error returned",
			nil,
			"400",
			"/cgi/url_redirect.cgi",
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(400)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc(tc.endpoint, tc.handler)

			server := httptest.NewTLSServer(mux)
			defer server.Close()

			parsedURL, err := url.Parse(server.URL)
			if err != nil {
				t.Fatal(err)
			}

			client := NewClient(parsedURL.Hostname(), "foo", "bar", logr.Discard(), WithPort(parsedURL.Port()))

			image, err := client.fetchScreenPreview(context.Background())
			if err != nil {
				if tc.errorContains != "" {
					assert.ErrorContains(t, err, tc.errorContains)
					return
				}
			}

			assert.Nil(t, err)
			assert.Equal(t, tc.expectImage, image)
		})
	}
}
