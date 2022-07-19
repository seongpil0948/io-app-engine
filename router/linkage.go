package router

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	tk "github.com/io-boxies/io-app-engine/controller/token"
)

func SetLinkRoutes(g *gin.RouterGroup) {
	g.POST("/saveCafeToken", saveCafeToken)
	g.GET("/getCafeOrders", getCafeOrders)
}

func saveCafeToken(c *gin.Context) {
	code := c.PostForm("code")
	if len(code) < 3 {
		c.AbortWithError(400, errors.New("code Field is required"))
		return
	}
	redirectUri := c.PostForm("redirectUri")
	if len(redirectUri) < 3 {
		c.AbortWithError(400, errors.New("redirectUri Field is required"))
		return
	}
	mallId := c.PostForm("mallId")
	if len(mallId) < 3 {
		c.AbortWithError(400, errors.New("mallId Field is required"))
		return
	}
	userId := c.PostForm("userId")
	if len(userId) < 3 {
		c.AbortWithError(400, errors.New("userId Field is required"))
		return
	}
	status, err := tk.SaveCafeToken(code, redirectUri, mallId, userId)
	if err != nil {
		c.AbortWithStatusJSON(status, err)
		return
	}

	c.String(http.StatusOK, "OK")
}

func getCafeOrders(c *gin.Context) {
	mallId := c.Query("mallId")
	if len(mallId) < 3 {
		c.AbortWithError(400, errors.New("mallId Field is required"))
		return
	}
	userId := c.Query("userId")
	if len(userId) < 3 {
		c.AbortWithError(400, errors.New("userId Field is required"))
		return
	}
	startDate := c.Query("startDate")
	if len(startDate) < 3 {
		c.AbortWithError(400, errors.New("startDate Field is required"))
		return
	}
	endDate := c.Query("endDate")
	if len(endDate) < 3 {
		c.AbortWithError(400, errors.New("endDate Field is required"))
		return
	}
	orders, err := tk.GetCafeOrders(mallId, userId, startDate, endDate)
	if err != nil {
		c.AbortWithStatusJSON(500, err)
	} else {
		c.JSON(200, gin.H{
			"orders": orders,
		})
	}

}
