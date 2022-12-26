package controller

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"

	"firebase.google.com/go/v4/messaging"
	"github.com/io-boxies/io-app-engine/controller/fire"
	"google.golang.org/api/iterator"
)

const smtpAddress = "smtp.gmail.com"
const smtpId = "inoutboxofficial@gmail.com"
const smtpPw = "enxhdimmhxziphsg"

func IoSendMail(userIds []string, subject string, mailBody string) error {
	auth := smtp.PlainAuth("", smtpId, smtpPw, smtpAddress)
	from := smtpId
	var to []string
	app := fire.GetFireInstance()
	storeClient, _ := app.Inst.Firestore(app.Ctx)
	if len(userIds) > 0 {
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
		bd := fmt.Sprintf("<html><body><div>%s</div></body></html>", mailBody)
		msg := []byte(frm + toto + sbj + mime + bd)
		// 메일 보내기
		err := smtp.SendMail(fmt.Sprintf("%s:587", smtpAddress), auth, from, to, msg)
		if err != nil {
			return err
		}
	}

	return nil
}

func IoSendPush(userIds []string, webToLink string,
	param_tokens []string, data map[string]string, pushTitle string, pushBody string) error {

	if len(userIds) > 0 {
		app := fire.GetFireInstance()
		msgClient, _ := app.Inst.Messaging(app.Ctx)
		storeClient, _ := app.Inst.Firestore(app.Ctx)
		iter := storeClient.Collection("user").Where("userInfo.userId", "in", userIds).Documents(app.Ctx)

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
		if len(param_tokens) < 1 {
			log.Print("none param_tokens")
			return nil
		}
		logo := "https://inout-box.com/logo.png"
		message := &messaging.MulticastMessage{
			Data: data,
			Notification: &messaging.Notification{
				Title:    pushTitle,
				Body:     pushBody,
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
		_, err := msgClient.SendMulticast(app.Ctx, message)
		if err != nil {
			return err
		}
	}
	return nil
}
