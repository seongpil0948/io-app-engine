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
	config.AllowOriginFunc = allowOriginFunc
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
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		fmt.Println(pair[0], ":", pair[1])
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

var allowOrigins = []string{
	"https://io-box.firebaseapp.com",
	"https://io-box.web.app",
	"https://inout-box.com",
	"io-box-admin.web.app",
	"5173",
	"5174",
	"8080",
	"io-box--dev-pcug7p0p.web.app",
	"io-box--pr1-dev-tr8yrr1h.web.app",
}

func allowOriginFunc(origin string) bool {
	fmt.Printf("origin: %s in allowOriginFunc", origin)
	for i := 0; i < len(allowOrigins); i++ {
		o := allowOrigins[i]
		if strings.Contains(o, origin) == true {
			return true
		} else if strings.Contains(origin, o) == true {
			return true
		}
	}
	return false
}
