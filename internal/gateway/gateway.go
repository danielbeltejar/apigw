package gateway

import (
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"apigw/internal/metrics"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
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
	Logger *logrus.Logger
}

func NewGateway(logger *logrus.Logger) *Gateway {
	return &Gateway{Logger: logger}
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metrics.RequestsTotal.Inc()

	for _, route := range g.Routes {
		if strings.HasPrefix(r.URL.Path, route.Pattern) && (len(r.URL.Path) == len(route.Pattern) || r.URL.Path[len(route.Pattern)] == '/') {
			methodAllowed := false
			for _, method := range route.AllowedMethods {
				if r.Method == method {
					methodAllowed = true
					break
				}
			}
			if !methodAllowed {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				metrics.MethodNotAllowedTotal.Inc()
				return
			}

			backendURL := *route.BackendAppURL
			proxy := httputil.NewSingleHostReverseProxy(&backendURL)
			g.Logger.Infof("Proxying request to backend URL: %s", backendURL.String())
			proxy.ServeHTTP(w, r)
			metrics.SuccessfulRequestsTotal.Inc()
			return
		}
	}

	http.NotFound(w, r)
	metrics.NotFoundTotal.Inc()
}

func (g *Gateway) LoadConfig(filename string) {
	configData, err := ioutil.ReadFile(filename)
	if err != nil {
		g.Logger.Fatalf("Error reading configuration file: %v", err)
	}

	var gatewayConfig GatewayConfig
	err = yaml.Unmarshal(configData, &gatewayConfig)
	if err != nil {
		g.Logger.Fatalf("Error parsing YAML: %v", err)
	}

	var routes []Route
	for _, routeConfig := range gatewayConfig.Routes {
		backendURL, err := url.Parse(routeConfig.BackendDNS)
		if err != nil {
			g.Logger.Fatalf("Error parsing backend DNS URL for route %s: %v", routeConfig.Pattern, err)
		}
		routes = append(routes, Route{
			Pattern:        routeConfig.Pattern,
			BackendAppURL:  backendURL,
			AllowedMethods: routeConfig.Method,
		})
	}

	g.Routes = routes
}
