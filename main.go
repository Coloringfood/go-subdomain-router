package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"net/url"
	"net/http/httputil"
)

/**
START JSON Import outline
 */
type JsonRoot struct {
	Router Router
}

type Router struct {
	Port     string
	Handlers []Handler
}

type Handler struct {
	SubDomain string
	Path      string
	Regex	  string
}
/**
END JSON Import outline
 */

/**
Start Proxy Objects
 */
type Prox struct {
	target        *url.URL
	proxy         *httputil.ReverseProxy
	routePatterns []*regexp.Regexp
}

func New(target string) *Prox {
	target_url, _ := url.Parse(target)

	return &Prox{target: target_url, proxy: httputil.NewSingleHostReverseProxy(target_url)}
}

func (p *Prox) handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-GoProxy", "GoProxy")

	if p.routePatterns == nil || p.parseWhiteList(r) {
		p.proxy.ServeHTTP(w, r)
	}
}

func (p *Prox) parseWhiteList(r *http.Request) bool {
	for _, target_regexp := range p.routePatterns {
		fmt.Println(r.URL.Path)
		if target_regexp.MatchString(r.URL.Path) {
			return true
		}
	}
	fmt.Println("Not accepted routes ", r.URL.Path)
	return false
}
/**
End Proxy Objects
 */

func HttpHandler(w http.ResponseWriter, r *http.Request, router Router) {
	originalDomain := r.Host
	var handler Handler
	fmt.Println(originalDomain)

	for _, v := range router.Handlers {
		fmt.Println(v)
		if v.SubDomain == originalDomain {
			handler = v
		}
	}
	if handler.SubDomain == "" {
		w.WriteHeader(502)
		fmt.Printf("subdomain not defined %s", originalDomain)
		return
	}
	fmt.Println(handler)

	//
	reg, _ := regexp.Compile(handler.Regex)
	routes := []*regexp.Regexp{reg}

	// proxy
	proxy := New(handler.Path)
	proxy.routePatterns = routes

	proxy.handle(w, r)

}

func main() {
	Router := readConfig()
	fmt.Printf("listening on port: %v\n", Router.Port)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		HttpHandler(w, r, Router)
	})
	err := http.ListenAndServe(":" + Router.Port, nil)
	if err == nil {
		fmt.Println("Successfully loaded")
	} else {
		fmt.Println("Unable to start: ", err)
	}
}

// readConfig reads configuration file written in json format, returns the Router struct
func readConfig() Router {
	file, e := ioutil.ReadFile("./config.json")
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}
	var root JsonRoot
	json.Unmarshal(file, &root)
	return root.Router
}
