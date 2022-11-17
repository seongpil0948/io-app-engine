package router

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	ctr "github.com/io-boxies/io-app-engine/controller"
)

func SetMailRoutes(c *gin.RouterGroup) {
	c.POST("/sendEmail", sendEmail)
}

func sendEmail(c *gin.Context) {
	userIds := c.PostFormArray("toUserIds")
	subject := c.PostForm("subject")
	body := c.PostForm("body")
	if len(subject) < 2 || len(body) < 2 {
		c.String(http.StatusBadRequest, "메일 제목 또는 본문이 누락되었습니다.")
	}
	err := ctr.IoSendMail(userIds, subject, body)

	// err := sendMail(c.Request, targets, subject, body)
	if err != nil {
		log.Fatalf("Couldn't send email: %v", err.Error())
	} else {
		c.String(http.StatusOK, "전송에 성공하였습니다.")
	}
}
