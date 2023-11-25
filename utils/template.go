package utils

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"os"
	"strings"
)

// Inertia converts a Go value to an Inertia template HTML string.
func Inertia(v interface{}) template.HTML {
	retVal, err := json.Marshal(v)
	if err != nil {
		// Handle error appropriately, e.g., log or return an error HTML
		return template.HTML("")
	}

	data := template.HTMLEscapeString(string(retVal))
	return template.HTML(fmt.Sprintf("<div id='app' data-page='%s'></div>", data))
}

// JsonEncode encodes a Go value to a JSON string and returns it as a template JS.
func JsonEncode(v interface{}) template.JS {
	retVal, err := json.Marshal(v)
	if err != nil {
		// Handle error appropriately, e.g., log or return an error JS
		return template.JS("")
	}
	return template.JS(string(retVal))
}

// JsonEncodeRaw encodes a Go value to a raw JSON string.
func JsonEncodeRaw(v interface{}) string {
	retVal, err := json.Marshal(v)
	if err != nil {
		// Handle error appropriately, e.g., log or return an error string
		return ""
	}
	return string(retVal)
}

// Vite resolves the asset path using the manifest file and returns it as template HTML.
func Vite(path string, manifestPath ...string) template.HTML {
	// Ensure the path starts with "/"
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Determine the manifest path
	mPath := GetEnvOrDefault("INERTIA_PUBLIC_PATH", "public") + "/manifest.json"
	if len(manifestPath) > 0 {
		mPath = manifestPath[0]
	}

	// Open and read the manifest file
	manifestFile, err := os.Open(mPath)
	if err != nil {
		// Handle error appropriately, e.g., log or return the original path
		return template.HTML(path)
	}
	defer manifestFile.Close()

	manifestData, err := io.ReadAll(manifestFile)
	if err != nil {
		// Handle error appropriately, e.g., log or return the original path
		return template.HTML(path)
	}

	// Unmarshal the manifest data
	var manifest map[string]string
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		// Handle error appropriately, e.g., log or return the original path
		return template.HTML(path)
	}

	// Return the resolved path from the manifest
	return template.HTML(manifest[path])
}
