package router

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/io-boxies/io-app-engine/controller/linkage"
	tk "github.com/io-boxies/io-app-engine/controller/token"
)

func SetLinkRoutes(g *gin.RouterGroup) {
	g.POST("/getZigZagOrders", getZigZagOrders)
	g.POST("/saveCafeToken", saveCafeToken)
	g.POST("/getCafeOrders", getCafeOrders)
	g.GET("/refreshTokens", refreshTokens)
}

func saveCafeToken(c *gin.Context) {
	code := c.PostForm("code")
	if len(code) < 3 {
		c.AbortWithStatusJSON(400, gin.H{"err": "code Field is required"})
		return
	}
	redirectUri := c.PostForm("redirectUri")
	if len(redirectUri) < 3 {
		c.AbortWithStatusJSON(400, gin.H{"err": "redirectUri Field is required"})
		return
	}
	mallId := c.PostForm("mallId")
	if len(mallId) < 3 {
		c.AbortWithStatusJSON(400, gin.H{"err": "mallId Field is required"})
		return
	}
	userId := c.PostForm("userId")
	if len(userId) < 3 {
		c.AbortWithStatusJSON(400, gin.H{"err": "userId Field is required"})
		return
	}
	status, err := tk.SaveCafeToken(code, redirectUri, mallId, userId)
	if err != nil {
		c.AbortWithStatusJSON(status, err)
		return
	}

	c.String(http.StatusOK, "OK")
}

func refreshTokens(c *gin.Context) {
	cronH := c.Request.Header["X-Appengine-Cron"]
	isCron := len(cronH) > 0 && cronH[0] == "true"
	if !isCron {
		log.Printf("요청 헤더에서 크론명세를 발견하지 못했습니다. %#v", cronH)
	}
	err := tk.RefreshTokens()
	if err != nil {
		log.Fatalln(err.Error())
		c.AbortWithStatusJSON(500, gin.H{"err": err})
	}
	c.String(200, "refresh tokens is done")
}
func getCafeOrders(c *gin.Context) {
	mallId := c.PostForm("mallId")
	if len(mallId) < 3 {
		c.AbortWithStatusJSON(400, gin.H{"err": "mallId Field is required"})
		return
	}
	userId := c.PostForm("userId")
	if len(userId) < 3 {
		c.AbortWithStatusJSON(400, gin.H{"err": "userId Field is required"})
		return
	}
	startDate := c.PostForm("startDate")
	if len(startDate) < 3 {
		c.AbortWithStatusJSON(400, gin.H{"err": "startDate Field is required"})
		return
	}
	endDate := c.PostForm("endDate")
	if len(endDate) < 3 {
		c.AbortWithStatusJSON(400, gin.H{"err": "endDate Field is required"})
		return
	}
	tokenId := c.PostForm("tokenId")
	if len(tokenId) < 3 {
		c.AbortWithStatusJSON(400, gin.H{"err": "tokenId Field is required"})
		return
	}
	orders, err := tk.GetCafeOrders(mallId, userId, startDate, endDate, tokenId)
	if err["err"] == "doc not exist" {
		c.AbortWithStatusJSON(401, err)
	} else if err != nil {
		c.AbortWithStatusJSON(500, err)
	} else {
		c.JSON(200, gin.H{
			"orders": orders,
		})
	}
}

func getZigZagOrders(c *gin.Context) {
	tokenDbId := c.PostForm("tokenDbId")
	if len(tokenDbId) < 3 {
		c.AbortWithStatusJSON(400, gin.H{"err": "tokenDbId is required"})
		return
	}
	userId := c.PostForm("userId")
	if len(userId) < 3 {
		c.AbortWithStatusJSON(400, gin.H{"err": "userId Field is required"})
		return
	}
	startDate := c.PostForm("startDate")
	if len(startDate) < 3 {
		c.AbortWithStatusJSON(400, gin.H{"err": "startDate Field is required"})
		return
	}
	endDate := c.PostForm("endDate")
	if len(endDate) < 3 {
		c.AbortWithStatusJSON(400, gin.H{"err": "endDate Field is required"})
		return
	}

	var token tk.IoAuthToken
	dsnap, errObj := tk.GetToken(userId, tokenDbId)
	if errObj != nil {
		c.AbortWithStatusJSON(400, errObj)
		return
	}
	dataMap := dsnap.Data()
	dbByte, _ := json.Marshal(dataMap)
	_ = json.Unmarshal(dbByte, &token)
	if token.Service != "ZIGZAG" {
		c.AbortWithStatusJSON(400, gin.H{"err": fmt.Sprintf("%s는 지그재그 토큰이 아닙니다. %s", tokenDbId, token.Service)})
		return
	}
	startInt, err := strconv.Atoi(strings.Replace(startDate, "-", "", -1))
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"err": err.Error()})
		return
	}
	endInt, err := strconv.Atoi(strings.Replace(endDate, "-", "", -1))
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"err": err.Error()})
		return
	}
	resp, errObj := linkage.GetZigzagOrders(token.AccessKey, token.SecretKey, token.CreatedAt.Format("2006-01-02"), startInt, endInt)
	if errObj != nil {
		c.AbortWithStatusJSON(500, errObj)
		return
	}
	c.JSON(200, resp)
}
