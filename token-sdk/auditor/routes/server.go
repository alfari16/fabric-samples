/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package routes

import (
	"fmt"
	"log"
	"net/http"
	"os"

	oapimiddleware "github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/labstack/echo/v4"
	middleware "github.com/labstack/echo/v4/middleware"
)

type Logger interface {
	Infof(template string, args ...interface{})
	Debugf(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Fatalf(template string, args ...interface{})
}

// Start web server on the main thread. It exits the application if it fails setting up.
func StartWebServer(port string, routesImplementation StrictServerInterface, logger Logger, staticToken string) error {
	e := echo.New()
	baseURL := "/api/v1"

	handler := NewStrictHandler(routesImplementation, nil)
	RegisterHandlersWithBaseURL(e, handler, baseURL)

	// Request validator
	swagger, err := GetSwagger()
	if err != nil {
		log.Fatalf("Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}
	swagger.Servers = nil
	private := e.Group(baseURL)
	private.Use(AuthTokenMiddleWare(staticToken))
	private.Use(oapimiddleware.OapiRequestValidator(swagger))

	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		Skipper: func(c echo.Context) bool {
			return c.Path() == "/api/v1/healthz" || c.Path() == "/api/v1/readyz"
		},
		LogRequestID: true, LogMethod: true, LogURI: true, LogStatus: true, LogLatency: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Status < 400 {
				logger.Infof("%d %s %s %s [%s]", v.Status, v.Method, v.URI, v.Latency.String(), v.RequestID)
			} else if v.Status >= 400 && v.Status < 500 {
				logger.Warnf("%d %s %s %s [%s]", v.Status, v.Method, v.URI, v.Latency.String(), v.RequestID)
			} else {
				logger.Errorf("%d %s %s %s [%s]", v.Status, v.Method, v.URI, v.Latency.String(), v.RequestID)
			}
			return nil
		},
	}))

	// Start REST API server
	return e.Start(fmt.Sprintf("0.0.0.0:%s", port))
}

func AuthTokenMiddleWare(StaticToken string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			xToken := c.Request().Header.Get("Authorization")
			// get bearer token either from header or cookie

			resp := map[string]interface{}{
				"success": false,
				"message": "Unauthorized Request",
				"data":    nil,
			}

			if xToken != StaticToken {
				return c.JSON(http.StatusUnauthorized, resp)
			}

			return next(c)
		}
	}
}
