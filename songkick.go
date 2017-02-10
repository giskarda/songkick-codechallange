package main

import (
	"bufio"
	"bytes"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	cache "github.com/patrickmn/go-cache"
)

var c = cache.New(12*time.Hour, 30*time.Second)

var allowed_host_header = "songkick-api-proxy"

type transport struct {
	http.RoundTripper
}

func (t *transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	request := req.URL.RawPath
	c_resp, found := c.Get(request)
	if found {
		log.Println("Found request in cache, avoid Roundtrip")
		r := bufio.NewReader(bytes.NewReader(c_resp.([]byte)))
		resp, err := http.ReadResponse(r, nil)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		return resp, nil
	}

	resp, err = t.RoundTripper.RoundTrip(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Println("cacca")
		log.Fatal(err)
		return nil, err
	}
	c.Set(request, dump, cache.DefaultExpiration)
	log.Println("Add request to cache, next time it will be faster")

	return resp, nil
}

func checkAndProxy(w http.ResponseWriter, r *http.Request, revProxy *httputil.ReverseProxy) {
	if r.Host != "" && r.Host == allowed_host_header {
		log.Println("Request can be accepted")
		revProxy.ServeHTTP(w, r)
	} else {
		log.Printf("Request couldn't be fullfilled with header Host: <%s>", r.Host)
	}
}

// taken from httputil
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func newReverseProxy(target *url.URL) *httputil.ReverseProxy {
	// Taken from httputil.NewSingleHostReverseProxy()
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		// This is where we differentiate from  NewSingleHostReverseProxy
		req.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}
	return &httputil.ReverseProxy{Director: director}
}
func main() {
	u, err := url.Parse("http://api.songkick.com/")
	if err != nil {
		log.Fatal(err)
	}

	revProxy := newReverseProxy(u)
	revProxy.Transport = &transport{http.DefaultTransport}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		checkAndProxy(w, r, revProxy)
	})

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Println("Starting Songkick proxy server. Listening on port 8080 and Proxying request to http://api.songkick.com")
	log.Fatal(srv.ListenAndServe())
}
