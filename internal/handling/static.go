package handling

import (
	"net/http"
	"strings"
)

func StaticHandler(fs http.Handler, strip string) http.Handler {
	return http.StripPrefix(strip, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".js") {
			w.Header().Set("Content-Type", "application/javascript")
		}
		if strings.HasSuffix(r.URL.Path, ".css") {
			w.Header().Set("Content-Type", "text/css")
		}
		fs.ServeHTTP(w, r)
	}))
}
