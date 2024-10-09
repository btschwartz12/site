package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func Proxy(target, prefix string) http.HandlerFunc {
	url, err := url.Parse(target)
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(url)
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)
		proxy.ServeHTTP(w, r)
	}
}
