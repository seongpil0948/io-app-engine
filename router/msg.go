package router

import (
	"fmt"
	"log"
	"net/http"

	"firebase.google.com/go/v4/messaging"
	"github.com/gin-gonic/gin"
	"github.com/io-boxies/io-app-engine/controller/fire"
	"google.golang.org/api/iterator"
)

func SetMsgRoutes(c *gin.RouterGroup) {
	c.POST("/sendPush", sendPush)
}

func sendPush(c *gin.Context) {
	var param_tokens []string
	app := fire.GetFireInstance()
	msgClient, _ := app.Inst.Messaging(app.Ctx)
	uIds := c.PostFormArray("userIds")
	storeClient, _ := app.Inst.Firestore(app.Ctx)
	param_tokens = c.PostFormArray("tokens")
	iter := storeClient.Collection("user").Where("userInfo.userId", "in", uIds).Documents(app.Ctx)
	webToLink := c.PostForm("toWebLink")
	if webToLink == "" {
		webToLink = "https://io-box.web.app"
	}
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			log.Fatalln(err)
		}
		userInfo := doc.Data()["userInfo"].(map[string]interface{})
		acc_tokens := userInfo["fcmTokens"].([]interface{})
		for _, t := range acc_tokens { // Add User Token If not in Param Token
			exist := false
			for _, r := range param_tokens {
				if t == r {
					exist = true
				}
			}
			if !exist {
				param_tokens = append(param_tokens, t.(string))
			}
		}

		fmt.Println()
	}
	logo := "https://io-box.web.app/logo.png"
	message := &messaging.MulticastMessage{
		Data: map[string]string{
			"data1": c.PostForm("data1"),
			"data2": c.PostForm("data2"),
			"data3": c.PostForm("data3"),
		},
		Notification: &messaging.Notification{
			Title:    c.PostForm("title"),
			Body:     c.PostForm("body"),
			ImageURL: logo,
		},
		Tokens: param_tokens,
		Webpush: &messaging.WebpushConfig{
			Headers: map[string]string{"Urgency": "high"},
			Notification: &messaging.WebpushNotification{
				RequireInteraction: true,
				Badge:              logo,
				Icon:               logo,
			},
			FCMOptions: &messaging.WebpushFCMOptions{Link: webToLink},
		},
	}
	response, err := msgClient.SendMulticast(app.Ctx, message)
	if err != nil {
		log.Fatalln(err)
	}
	// Response is a message ID string.
	c.JSON(http.StatusOK, fmt.Sprintf("Successfully sent message %v", *response))
}
