package inertia

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/theartefak/inertia-echo/utils"
)

// Response represents an Inertia response.
type Response struct {
	component string
	props     map[string]interface{}
	viewData  map[string]interface{}
	rootView  string
	version   string
	status    int
}

// NewResponse creates a new Inertia response.
func NewResponse(component string, props map[string]interface{}, rootView string, version string) Response {
	var r Response
	r.component = component
	r.props = props
	r.viewData = make(map[string]interface{})
	r.rootView = rootView
	r.version = version
	r.status = http.StatusOK
	return r
}

// With adds a key-value pair to the response's props.
func (r Response) With(key interface{}, value interface{}) Response {
	switch key.(type) {
	case string:
		r.props[key.(string)] = value
		break
	case map[string]interface{}:
		for k, v := range key.(map[string]interface{}) {
			r.props[k] = v
		}
	}
	return r
}

// WithViewData adds a key-value pair to the response's view data.
func (r Response) WithViewData(key interface{}, value interface{}) Response {
	switch key.(type) {
	case string:
		r.viewData[key.(string)] = value
		break
	case map[string]interface{}:
		for k, v := range key.(map[string]interface{}) {
			r.props[k] = v
		}

		for k, v := range key.(map[string]interface{}) {
			r.viewData[k] = v
		}
	}
	return r
}

// ToResponse sends the Inertia response to the client.
func (r Response) ToResponse(c echo.Context) error {
	req := c.Request()

	var only []string
	if data := req.Header.Get(HeaderPartialData); data != "" {
		only = strings.Split(data, ",")
	}

	var props map[string]interface{}
	if only != nil && req.Header.Get(HeaderPartialData) == r.component {
		props = make(map[string]interface{})
		for _, v := range only {
			props[v] = r.props[v]
		}
	} else {
		props = r.props
	}

	// Iterate over props and handle special cases
	utils.WalkRecursive(props, func(prop interface{}) {
		type HandlerType func() interface{}
		switch prop.(type) {
		case func() interface{}:
			if f, ok := prop.(func() interface{}); ok {
				prop = HandlerType(f)
			}
		case Response:
			prop = prop.(Response).ToResponse(c)
		}
	})

	// Determine the scheme based on request details
	scheme := utils.GetEnvOrDefault("SCHEME", "http")
	if req.TLS != nil || req.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}

	// Prepare data for the Inertia page
	page := map[string]interface{}{
		"component": r.component,
		"props":     props,
		"url":       req.URL.String(),
		"version":   r.version,
		// Inertia-Echo-specifics
		"host":   req.Host,
		"path":   req.URL.Path,
		"scheme": scheme,
		"method": req.Method,
		"status": r.status,
	}

	// Check if Inertia header is present
	if req.Header.Get(HeaderPrefix) == "true" {
		c.Response().Header().Set("Vary", "Accept")
		c.Response().Header().Set("X-Inertia", "true")
		return c.JSON(r.status, page)
	}

	// Attach page data to view data and render the view
	r.viewData["page"] = page
	return c.Render(r.status, r.rootView, r.viewData)
}

// Status sets a custom status for the Inertia response.
func (r Response) Status(code int) Response {
	r.status = code
	return r
}
