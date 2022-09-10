package router

import (
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/io-boxies/io-app-engine/controller/fire"
	"google.golang.org/api/iterator"
)

func SetMailRoutes(c *gin.RouterGroup) {
	c.POST("/sendEmail", sendEmail)
}

const smtpAddress = "smtp.gmail.com"
const smtpId = "inoutboxofficial@gmail.com"
const smtpPw = "enxhdimmhxziphsg"

func sendEmail(c *gin.Context) {
	auth := smtp.PlainAuth("", smtpId, smtpPw, smtpAddress)

	from := smtpId
	userIds := c.PostFormArray("toUserIds")
	subject := c.PostForm("subject")
	body := c.PostForm("body")
	if len(subject) < 2 || len(body) < 2 {
		c.String(http.StatusBadRequest, "메일 제목 또는 본문이 누락되었습니다.")
	}

	var to []string
	app := fire.GetFireInstance()
	storeClient, _ := app.Inst.Firestore(app.Ctx)
	iter := storeClient.Collection("user").Where("userInfo.userId", "in", userIds).Documents(app.Ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			log.Fatalln(err)
		}
		userInfo := doc.Data()["userInfo"].(map[string]interface{})
		email := userInfo["email"].(string)
		to = append(to, email)
	}
	frm := fmt.Sprintf("From: %s\r\n", from)
	toto := fmt.Sprintf("To: %s\r\n", strings.Join(to, ", "))
	sbj := fmt.Sprintf("Subject: %s\n", subject)
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	bd := fmt.Sprintf("<html><body><div>%s</div></body></html>", body)
	msg := []byte(frm + toto + sbj + mime + bd)
	// 메일 보내기
	err := smtp.SendMail(fmt.Sprintf("%s:587", smtpAddress), auth, from, to, msg)
	if err != nil {
		panic(err)
	}

	// err := sendMail(c.Request, targets, subject, body)
	// if err != nil {
	// 	log.Errorf(c.Request.Context(), "Couldn't send email: %v", err)
	// } else {
	// 	c.String(http.StatusOK, "전송에 성공하였습니다.")
	// }
}
