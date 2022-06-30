// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by unknown module path version unknown version DO NOT EDIT.
package api

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {

	// (GET /metrics)
	GetMetrics(ctx echo.Context) error
	// Cloud Control Request
	// (POST /v1)
	PostV1(ctx echo.Context) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// GetMetrics converts echo context to params.
func (w *ServerInterfaceWrapper) GetMetrics(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetMetrics(ctx)
	return err
}

// PostV1 converts echo context to params.
func (w *ServerInterfaceWrapper) PostV1(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.PostV1(ctx)
	return err
}

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface) {
	RegisterHandlersWithBaseURL(router, si, "")
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.GET(baseURL+"/metrics", wrapper.GetMetrics)
	router.POST(baseURL+"/v1", wrapper.PostV1)

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/7SV0W/6NhDH/5XotkdDoGhSlbeOVRPS2qJW2x4qHkxygLvEds8XBEL53yfbBArJr1Kl",
	"/p6I8fl7n5y/dzlAbiprNGp2kB2A0FmjHYbFC9IW6Z7IkF/mRjNq9o/S2lLlkpXR6Zsz2v/n8g1W0j/9",
	"SriCDH5Jz9pp3HVpVGuaRkCBLidlvQhkMNOMpGWZxKxJGyiOwoHoxGLJWCRWETQ3BfrfS8UQnIQ9AStD",
	"lWTIQGme3IAA3luMS1wjQSOgQufk+odC7fbpqGNSeh0QCd9rRVhA9grHhG34ohHwIHUh2dC+i36XxyQH",
	"QF1X/vxfyjEImBJK9kLPKAsQ8Lct4voPLDE8PNb/xX1nyi3C4hpMwG5gpFUDT7RGPcAdkxywXIfMW1mq",
	"IJmd8f27TE1VSV34kG/Rm0uSVYhQjPHhqoAncEkk952CHkt0Jjtpnl/ZLN8wZy/1gEwq/1I+AU/hql33",
	"eu53eVlHc32BvgP1jO9d7QtXfNYz58BL1M/OtGHXxTxr9RXvpc5zdD11KCSH1j610XLPPa0gAIl6a/Rm",
	"lqrHU9d4MSzKiJi1y+kPKb0y3U59upvPkpWhZFqaukimRjOZ0nMqLr3E478vl3vJ3XwGArZILkpsR8PR",
	"cOyRjUUtrYIMJsPRcAICrORNKEdanV22xjASfbXCQJwVkMGfyK0RxeVQvRmNvjRMT7b71CHHXF0zdubs",
	"kTxpoSBErGRd8s8f8rXGncWcsUiwjWkEpNtxsJxxPaWcG8f/jCH6BB3/bor9t5H6xuzhnE6TY7Zk6dN9",
	"NClTjU3nUsffhtT2YD9WmLIXl/dbNFSf5Ikx/fgpD9/Uuqqknzxw2Q3P8a19x4Sx/gqVVBoWAcYFEf/v",
	"AWoqIYMNs83StDS5LDfGcXY7up1As2j+DwAA//86o4/yWAgAAA==",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	var res = make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	var resolvePath = PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		var pathToFile = url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
