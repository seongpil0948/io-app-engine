package router

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func SetCommonRoutes(g *gin.RouterGroup) {
	g.GET("/ping", Ping)
	g.GET("/checkClient", checkClient)
}

func Ping(g *gin.Context) {
	g.JSON(http.StatusOK, gin.H{
		"msg": "pong",
	})
}

func checkClient(c *gin.Context) {
	// FIXME: 미들웨어로 모든요청에 놓을수 있도록하자
	id := uuid.New()
	getPath := c.Request.URL.String()
	remoteIp, port, _ := net.SplitHostPort(c.Request.RemoteAddr)
	// let's get the request HTTP header "X-Forwarded-For (XFF)"
	// if the value returned is not null, then this is the real IP address of the user.
	c.JSON(200, gin.H{
		"uuid":            id.String(),
		"pathInfo":        getPath,
		"remoteIp":        remoteIp,
		"remotePort":      port,
		"ip":              c.ClientIP(),
		"X-Forwarded-For": c.GetHeader("X-FORWARDED-FOR"),
	})
}
