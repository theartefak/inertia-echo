package inertia

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"sync"

	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/theartefak/inertia-echo/utils"
)

// The Base "X-Inertia" header prefixes
const (
	HeaderPrefix       = "X-Inertia"
	HeaderVersion      = HeaderPrefix + "-Version"
	HeaderLocation     = HeaderPrefix + "-Location"
	HeaderPartialData  = HeaderPrefix + "-Partial-Data"
)

// Inertia is a struct representing the Inertia handler.
type Inertia struct {
	config           InertiaConfig
	templates        *template.Template
	sharedProps      map[string]map[string]interface{}
	sharedPropsMutex *sync.Mutex
	version          interface{}
}

// InertiaConfig holds the configuration for the Inertia handler.
type InertiaConfig struct {
	Echo             *echo.Echo
	PublicPath       string
	TemplatesPath    string
	RootView         string
	TemplateFuncMap  template.FuncMap
	HTTPErrorHandler echo.HTTPErrorHandler
	RequestIDConfig  middleware.RequestIDConfig
}

// DefaultHTTPErrorHandler is the default error handler for Inertia.
var DefaultHTTPErrorHandler = func(err error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}
	_ = c.Render(code, "Error", map[string]interface{}{
		"status": code,
	})
}

// Create a default Inertia Config to use without the hassle of setting up everything
func NewDefaultInertiaConfig(e *echo.Echo) (i InertiaConfig) {
	i = InertiaConfig{}
	i.Echo = e

	// Get the configured root view from the environment
	i.RootView = utils.GetEnvOrDefault("INERTIA_ROOT_VIEW", "app.html")
	// Get the configured public path from the environment
	i.PublicPath = utils.GetEnvOrDefault("INERTIA_PUBLIC_PATH", "public")
	// Get the configured templates path from the environment
	i.TemplatesPath = utils.GetEnvOrDefault("INERTIA_RESOURCES_PATH", "resources") + "/views/*.html"

	// Set a default error handler to render a default error page
	i.HTTPErrorHandler = DefaultHTTPErrorHandler

	// Register convenient template functions
	i.TemplateFuncMap = template.FuncMap{
		"inertia":         utils.Inertia,
		"json_encode":     utils.JsonEncode,
		"json_encode_raw": utils.JsonEncodeRaw,
		"vite":             utils.Vite,
		"routes": func() template.JS {
			retVal, _ := json.Marshal(e.Routes())
			return template.JS(string(retVal))
		},
		"routes_ziggy": func(v interface{}) template.HTML {
			ziggy := utils.NewZiggy(e, v.(map[string]interface{}))
			retVal, _ := json.Marshal(ziggy)
			return template.HTML(fmt.Sprintf("<script>const Ziggy = %s;</script>", string(retVal)))
		},
	}

	i.RequestIDConfig = middleware.DefaultRequestIDConfig

	return i
}

// Instance a new Inertia Handler with a default config
func NewInertia(echo *echo.Echo) (i *Inertia) {
	return NewInertiaWithConfig(NewDefaultInertiaConfig(echo))
}

// Instance a new Inertia Handler with a configuration
func NewInertiaWithConfig(config InertiaConfig) (i *Inertia) {
	if config.Echo == nil {
		log.Fatal("[Inertia] echo.Echo reference required in the given config!")
	}

	i = new(Inertia)
	i.config = config
	i.sharedProps = make(map[string]map[string]interface{})
	i.sharedPropsMutex = &sync.Mutex{}
	i.config.Echo.Renderer = i
	i.config.Echo.HTTPErrorHandler = i.config.HTTPErrorHandler
	log.Printf("[Inertia] Loading templates out of %s", i.config.TemplatesPath)
	i.templates = template.Must(template.New("").Funcs(i.config.TemplateFuncMap).ParseGlob(i.config.TemplatesPath))
	// Try to set a version off of the manifest, if any
	i.SetViteVersion()
	// Register a unique id generator to identify requests
	i.config.Echo.Use(middleware.RequestIDWithConfig(i.config.RequestIDConfig))

	return i
}

// Render renders a template document
func (i *Inertia) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// Always empty the shared props for i request
	sharedProps := i.Shared(c)
	isMap := reflect.TypeOf(data).Kind() == reflect.Map
	if i.templates.Lookup(name) != nil {
		if isMap {
			viewContext := data.(map[string]interface{})
			viewContext["reverse"] = c.Echo().Reverse
			viewContext["shared"] = sharedProps
		}
		return i.templates.ExecuteTemplate(w, name, data)
	}

	if isMap {
		if sharedProps != nil {
			return NewResponse(name, utils.MergeMaps(sharedProps, data.(map[string]interface{})), i.config.RootView, i.GetVersion()).Status(c.Response().Status).ToResponse(c)
		} else {
			return NewResponse(name, data.(map[string]interface{}), i.config.RootView, i.GetVersion()).Status(c.Response().Status).ToResponse(c)
		}
	}
	return NewResponse(name, sharedProps, i.config.RootView, i.GetVersion()).Status(c.Response().Status).ToResponse(c)
}

// Share a key/value pairs with every response
func (i *Inertia) Share(c echo.Context, key string, value interface{}) {
	rid := c.Request().Header.Get(echo.HeaderXRequestID)
	i.sharedPropsMutex.Lock()
	if reqSharedProps, ok := i.sharedProps[rid]; ok {
		reqSharedProps[key] = value
	} else {
		i.sharedProps[rid] = map[string]interface{}{
			key: value,
		}
	}
	i.sharedPropsMutex.Unlock()
}

// Share multiple key/values with every response
func (i *Inertia) Shares(c echo.Context, values map[string]interface{}) {
	rid := c.Request().Header.Get(echo.HeaderXRequestID)
	i.sharedPropsMutex.Lock()
	if _, ok := i.sharedProps[rid]; !ok {
		i.sharedProps[rid] = make(map[string]interface{})
	}
	i.sharedPropsMutex.Unlock()

	for key, value := range values {
		i.sharedPropsMutex.Lock()
		i.sharedProps[rid][key] = value
		i.sharedPropsMutex.Unlock()
	}
}

// Get a specific key-value from the shared information
func (i *Inertia) GetShared(c echo.Context, key string) (interface{}, bool) {
	rid := c.Request().Header.Get(echo.HeaderXRequestID)
	i.sharedPropsMutex.Lock()
	if reqSharedProps, ok := i.sharedProps[rid]; ok {
		value, ok := reqSharedProps[key]
		i.sharedPropsMutex.Unlock()
		return value, ok
	}
	return nil, false
}

// Returns the shared props (if any) and deletes them
func (i *Inertia) Shared(c echo.Context) map[string]interface{} {
	rid := c.Request().Header.Get(echo.HeaderXRequestID)
	i.sharedPropsMutex.Lock()
	sharedProps := i.sharedProps[rid]
	delete(i.sharedProps, rid)
	i.sharedPropsMutex.Unlock()
	return sharedProps
}

// Set a version callback "func() string"
func (i *Inertia) Version(version func() string) {
	i.version = version
}

// Set a version string
func (i *Inertia) SetVersion(version string) {
	i.version = version
}

// Create a version hash off of the manifest.json file md5
func (i *Inertia) SetViteVersion(viteManifestPath ...string) bool {
	filePath := i.config.PublicPath + "/manifest.json"
	if len(viteManifestPath) > 0 {
		filePath = viteManifestPath[0]
	}
	fileData, err := os.ReadFile(filePath)
	if err == nil {
		hash := md5.New()
		hash.Write(fileData)
		i.version = hex.EncodeToString(hash.Sum(nil))
		return true
	}
	return false
}

//
func (i *Inertia) GetVersion() string {
	if i.version != nil {
		type HandlerType func() string
		switch i.version.(type) {
		case func() string:
			if f, ok := i.version.(func() string); ok {
				i.version = HandlerType(f)
			}
			break
		}

		return i.version.(string)
	}

	return ""
}
