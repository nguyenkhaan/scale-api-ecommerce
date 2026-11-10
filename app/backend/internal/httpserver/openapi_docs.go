package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const scalarHTML = `<!doctype html>
<html>
  <head>
    <title>SaleScale API Reference</title>
    <meta charset="utf-8" />
  </head>
  <body>
    <script
      id="api-reference"
      data-url="/docs/openapi.yaml">
    </script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
  </body>
</html>`

func RegisterOpenAPIDocs(router *gin.Engine, path string) {
	router.GET("/docs/openapi.yaml", func(ctx *gin.Context) {
		ctx.Header("Content-Type", "application/yaml; charset=utf-8")
		ctx.File(path)
	})
	serveScalar := func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(scalarHTML))
	}
	router.GET("/docs", serveScalar)
	router.GET("/docs/", serveScalar)
	router.GET("/docs/index.html", serveScalar)
}
