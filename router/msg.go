package router

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	ctr "github.com/io-boxies/io-app-engine/controller"
)

func SetMsgRoutes(c *gin.RouterGroup) {
	c.POST("/sendPush", sendPush)
}

func sendPush(c *gin.Context) {
	var param_tokens []string
	webToLink := c.PostForm("toWebLink")
	userIds := c.PostFormArray("toUserIds")
	param_tokens = c.PostFormArray("tokens")
	title := c.PostForm("title")
	body := c.PostForm("body")
	err := ctr.IoSendPush(userIds, webToLink, param_tokens, map[string]string{}, title, body)
	if err != nil {
		log.Fatalf("error in sendPush: %s", err.Error())
	}

	c.String(http.StatusOK, "Successfully sent message ")
}
