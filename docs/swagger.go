// Package docs provides OpenAPI documentation for AKZA API.
// It embeds swagger-ui and serves the OpenAPI 3.0 specification.
package docs

import (
	"embed"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed openapi.json swagger/*
var content embed.FS

func RegisterSwagger(r *gin.Engine) {
	r.GET("/swagger/*filepath", func(c *gin.Context) {
		fp := c.Param("filepath")

		// default → index.html
		if fp == "" || fp == "/" {
			fp = "/index.html"
		}

		fp = strings.TrimPrefix(fp, "/")
		fullPath := path.Join("swagger", fp)

		// special case: openapi.json
		if fp == "openapi.json" {
			data, _ := content.ReadFile("openapi.json")
			c.Header("Content-Type", "application/json")
			c.Header("Cache-Control", "no-store")
			c.Data(http.StatusOK, "application/json", data)
			return
		}

		data, err := content.ReadFile(fullPath)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		// Content-Type
		switch {
		case strings.HasSuffix(fp, ".html"):
			c.Header("Content-Type", "text/html; charset=utf-8")
		case strings.HasSuffix(fp, ".css"):
			c.Header("Content-Type", "text/css; charset=utf-8")
		case strings.HasSuffix(fp, ".js"):
			c.Header("Content-Type", "application/javascript")
		}

		// cache (важно для production)
		c.Header("Cache-Control", "public, max-age=31536000")

		c.Data(http.StatusOK, http.DetectContentType(data), data)
	})
}