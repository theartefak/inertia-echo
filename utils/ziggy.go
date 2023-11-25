package utils

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// Ziggy is a struct to export compatible Echo routes for https://github.com/tightenco/ziggy.
type Ziggy struct {
	BaseDomain   string                `json:"domain"`
	BasePort     int                   `json:"port"`
	BaseProtocol string                `json:"protocol"`
	BaseUrl      string                `json:"url"`
	Group        string                `json:"group"`
	Routes       map[string]ZiggyRoute `json:"routes"`
}

// ZiggyRoute represents a single route for https://github.com/tightenco/ziggy.
type ZiggyRoute struct {
	Uri     string   `json:"uri"`
	Methods []string `json:"methods"`
	Domain  string   `json:"domain"`
}

// NewZiggy creates a Ziggy struct based on an Echo instance and page information.
func NewZiggy(e *echo.Echo, page map[string]interface{}) Ziggy {
	var z Ziggy

	// Set the default protocol to "http"
	z.BaseProtocol = "http"
	if scheme, ok := page["scheme"]; ok {
		z.BaseProtocol = scheme.(string)
	}

	// Extract host information from page or Echo server address
	var host string
	if h, ok := page["host"]; ok {
		host = h.(string)
	} else {
		host = e.Server.Addr
	}

	// Split host into domain and port
	s := strings.Split(host, ":")
	z.BaseDomain = s[0]
	z.BaseUrl = z.BaseProtocol + "://" + z.BaseDomain

	// Append port to the base URL if available
	if len(s) > 1 {
		port, _ := strconv.Atoi(s[1])
		if port > 0 {
			z.BasePort = port
			z.BaseUrl += ":" + strconv.Itoa(z.BasePort)
		}
	}

	// Initialize the Routes map
	z.Routes = make(map[string]ZiggyRoute, len(e.Routes()))

	// Iterate over Echo routes and populate the ZiggyRoutes map
	for _, route := range e.Routes() {
		// Exclude generated functions and routes
		if matched, _ := regexp.MatchString(`.func\d+$`, route.Name); matched {
			continue
		}

		// Update existing ZiggyRoute or create a new one
		if ziggyRoute, ok := z.Routes[route.Name]; ok {
			ziggyRoute.Methods = append(ziggyRoute.Methods, route.Method)
			z.Routes[route.Name] = ziggyRoute
		} else {
			z.Routes[route.Name] = ZiggyRoute{
				Uri: route.Path,
				Methods: []string{route.Method},
			}
		}
	}

	return z
}
