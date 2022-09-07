package router

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func initMiddle(r *gin.Engine) {
	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.Print("=== CustomRecovery ===")
		if err, ok := recovered.(string); ok {
			c.String(http.StatusInternalServerError, fmt.Sprintf("error: %s", err))
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	}))
	config := cors.DefaultConfig()
	// port not working 인듯..
	config.AllowOrigins = []string{"http://localhost", "http://localhost:8090", "https://io-box.firebaseapp.com", "https://io-box.web.app", "https://io-box--dev-wplgfcvy.web.app", "https://inout-box.com", "https://io-box--dev-pcug7p0p.web.app", "https://io-box--pr1-dev-tr8yrr1h.web.app"}
	r.Use(cors.New(config))
}

func InitRoutes() gin.Engine {
	if port := os.Getenv("PORT"); port == "" {
		os.Setenv("PORT", "8000")
	}
	mode := os.Getenv("GAE_ENV")
	if strings.HasPrefix(mode, "standard") {
		log.Printf("deploy App Engine for production, Port: %s", os.Getenv("PORT"))
		gin.SetMode(gin.ReleaseMode)
	} else {
		log.Printf("deploy App Engine for development, Port: %s", os.Getenv("PORT"))
	}
	r := gin.Default()
	initMiddle(r)

	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		log.Printf("===> Endpoint %v %v %v %v\n", httpMethod, absolutePath, handlerName, nuHandlers)
	}
	gin.ForceConsoleColor()
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		log.Printf("endpoint %v %v %v %v\n", httpMethod, absolutePath, handlerName, nuHandlers)
	}

	api := r.Group("/api")
	SetAuthRoutes(api.Group("auth"))
	SetPGRoutes(api.Group("payment"))
	SetMsgRoutes(api.Group("msg"))
	SetCommonRoutes(api.Group("common"))
	SetLogRoutes(api.Group("log"))
	SetMailRoutes(api.Group("mail"))
	SetLinkRoutes(api.Group("linkage"))
	return *r
}
