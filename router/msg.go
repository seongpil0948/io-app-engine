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
	userIds := c.PostFormArray("toUserIds")
	storeClient, _ := app.Inst.Firestore(app.Ctx)
	param_tokens = c.PostFormArray("tokens")
	iter := storeClient.Collection("user").Where("userInfo.userId", "in", userIds).Documents(app.Ctx)
	webToLink := c.PostForm("toWebLink")
	if webToLink == "" {
		webToLink = "https://inout-box.com"
	}
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			log.Fatalln(err)
		}
		userInfo := doc.Data()["userInfo"].(map[string]interface{})
		if acc_tokens, ok := userInfo["fcmTokens"].([]interface{}); ok {
			for _, t := range acc_tokens { // Add User Token If not in Param Token
				exist := false
				token := t.(map[string]interface{})["token"].(string)
				for _, r := range param_tokens {
					if token == r {
						exist = true
					}
				}
				if !exist {
					param_tokens = append(param_tokens, token)
				}
			}
		}

		fmt.Println()
	}
	logo := "https://inout-box.com/logo.png"
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
