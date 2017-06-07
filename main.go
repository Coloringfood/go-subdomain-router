package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
)

/**
START JSON Import outline
*/
type JsonRoot struct {
	Router Router
	Certs  Certs
}

type Certs struct {
	Cert string
	Key  string
}

type Router struct {
	Port     string
	Ssl_port string
	Handlers []Handler
}

type Handler struct {
	SubDomain string
	Path      string
	Regex     string
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
	fmt.Println(originalDomain, r.URL.Path)

	for _, v := range router.Handlers {
		if v.SubDomain == originalDomain {
			handler = v
		}
	}
	if handler.SubDomain == "" {
		w.WriteHeader(502)
		fmt.Printf("subdomain not defined %s", originalDomain)
		return
	}

	//
	reg, _ := regexp.Compile(handler.Regex)
	routes := []*regexp.Regexp{reg}

	// proxy
	proxy := New(handler.Path)
	proxy.routePatterns = routes

	proxy.handle(w, r)

}

// Skimmed from http://www.kaihag.com/https-and-go/
func redirectToHttps(w http.ResponseWriter, r *http.Request) {
	// remove/add not default ports from req.Host
	target := "https://" + r.Host + r.URL.Path
	if len(r.URL.RawQuery) > 0 {
		target += "?" + r.URL.RawQuery
	}
	fmt.Printf("redirect to: %s", target)
	http.Redirect(w, r, target, http.StatusMovedPermanently)
}

func main() {
	Config := readConfig()
	var err error = nil;

	if Config.Certs.Key != "" {
		// redirect every http request to https
		fmt.Printf("forwarding on port: %v\n", Config.Router.Port)
		go http.ListenAndServe(":" + Config.Router.Ssl_port, http.HandlerFunc(redirectToHttps))

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			HttpHandler(w, r, Config.Router)
		})
		fmt.Printf("listening SSL on port: %v\n", Config.Router.Ssl_port)
		err = http.ListenAndServeTLS(":" + Config.Router.Port, Config.Certs.Cert, Config.Certs.Key, nil)
	} else {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			HttpHandler(w, r, Config.Router)
		})
		fmt.Printf("listening on port: %v\n", Config.Router.Port)
		err = http.ListenAndServe(":" + Config.Router.Port, nil)
	}
	if err == nil {
		fmt.Println("Successfully loaded")
	} else {
		fmt.Println("Unable to start: ", err)
	}
}

// readConfig reads configuration file written in json format, returns the Router struct
func readConfig() JsonRoot {
	file, e := ioutil.ReadFile("./config.json")
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}
	var root JsonRoot
	json.Unmarshal(file, &root)
	return root
}
