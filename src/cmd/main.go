package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type RouteConfig struct {
	Pattern    string   `yaml:"pattern"`
	Method     []string `yaml:"method"`
	BackendDNS string   `yaml:"backend_dns"`
}

type GatewayConfig struct {
	Routes []RouteConfig `yaml:"routes"`
}

type Route struct {
	Pattern        string
	BackendAppURL  *url.URL
	AllowedMethods []string
}

type Gateway struct {
	Routes []Route
}

func NewGateway() *Gateway {
	return &Gateway{}
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clientAddr := r.RemoteAddr
	requestPath := r.URL.Path
	requestURL := r.URL.String()
	requestMethod := r.Method

	log.Printf("Incoming request from client %s: %s %s", clientAddr, requestMethod, requestURL)

	for _, route := range g.Routes {
		if strings.HasPrefix(requestPath, route.Pattern) {
			trimmedPath := strings.TrimPrefix(requestPath, "/api")
			if trimmedPath == "" {
				trimmedPath = "/"
			}

			methodAllowed := false
			for _, method := range route.AllowedMethods {
				if requestMethod == method {
					methodAllowed = true
					break
				}
			}
			if !methodAllowed {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				return
			}

			backendURL := *route.BackendAppURL
			backendURL.Path = trimmedPath

			proxy := httputil.NewSingleHostReverseProxy(&backendURL)
			proxy.Director = func(req *http.Request) {
				req.Host = backendURL.Host
				req.URL.Host = backendURL.Host
				req.URL.Scheme = backendURL.Scheme
				req.URL.Path = trimmedPath

				log.Printf("Proxying request to backend: %s %s", req.Method, req.URL.String())
			}

			proxy.ServeHTTP(w, r)
			return
		}
	}
	http.NotFound(w, r)
}

func (g *Gateway) LoadConfig(filename string) {
	configData, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Error reading configuration file:", err)
	}

	var gatewayConfig GatewayConfig
	err = yaml.Unmarshal(configData, &gatewayConfig)
	if err != nil {
		log.Fatal("Error parsing YAML:", err)
	}

	var routes []Route
	for _, routeConfig := range gatewayConfig.Routes {
		backendURL, err := url.Parse(routeConfig.BackendDNS)
		if err != nil {
			log.Fatalf("Error parsing backend DNS URL for route %s: %v", routeConfig.Pattern, err)
		}
		routes = append(routes, Route{
			Pattern:        routeConfig.Pattern,
			BackendAppURL:  backendURL,
			AllowedMethods: routeConfig.Method,
		})
	}

	g.Routes = routes
}

func main() {

	go func() {
		log.Println("Health check server listening on port 8081...")
		if err := http.ListenAndServe(":8081", nil); err != nil {
			log.Fatalf("Failed to start health check server: %v", err)
		}
	}()

	g := NewGateway()

	g.LoadConfig("config/config.yaml")

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})
	
	log.Println("API Gateway listening on port 8080...")

	for _, route := range g.Routes {
		log.Printf("Registered route: %s\n", route.Pattern)
	}

	log.Fatal(http.ListenAndServe(":8080", g))
}
