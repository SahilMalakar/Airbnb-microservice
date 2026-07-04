package utils

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func ProxyToService(targetBaseUrl string, pathPrefix string) http.HandlerFunc {
	target, err := url.Parse(targetBaseUrl)
	if err != nil {
		log.Fatalf("error parsing target URL %q: %v", targetBaseUrl, err)
	}

	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			// Strip the prefix FIRST, while pr.Out.URL.Path still holds
			// the full incoming path. This must happen before SetURL,
			// since SetURL joins target.Path with whatever is currently
			// in pr.Out.URL.Path — do it after, and target.Path never
			// survives into the final request.
			pr.Out.URL.Path = strings.TrimPrefix(pr.In.URL.Path, pathPrefix)

			// Now join target's scheme/host/path with our trimmed path,
			// and merge target's query params with the incoming ones.
			// This is the actual "merge" step — Go does it internally.
			pr.SetURL(target)

			pr.SetXForwarded()
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("proxy error forwarding to %s: %v", targetBaseUrl, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte(`{"error":"upstream service unavailable"}`))
		},
	}

	return proxy.ServeHTTP
}