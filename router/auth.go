package router

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/io-boxies/io-app-engine/controller/fire"
)

func SetAuthRoutes(g *gin.RouterGroup) {
	g.GET("/customToken/:userId", getCustomToken)
}

func getCustomToken(c *gin.Context) {
	app := fire.GetFireInstance()
	userId := c.Param("userId")
	client, err := app.Inst.Auth(app.Ctx)
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}
	claims := map[string]interface{}{
		"premiumAccount": true,
	}
	token, err := client.CustomTokenWithClaims(app.Ctx, userId, claims)
	if err != nil {
		log.Fatalf("error minting custom token: %v\n", err)
	}

	log.Printf("Got custom token: %v\n", token)
	c.JSON(200, gin.H{
		"userId": userId,
		"token":  token,
	})
}
