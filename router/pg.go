package router

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	ctr "github.com/io-boxies/io-app-engine/controller"
	"github.com/io-boxies/io-app-engine/controller/fire"
)

func SetPGRoutes(g *gin.RouterGroup) {
	g.GET("/verifyReceipt", verifyReceipt)
	g.POST("/onFeedBack", onFeedBack)
}

func verifyReceipt(c *gin.Context) {
	pg := ctr.GetBootPay()
	log.Printf("verifyReceipt queries: %s", c.Request.URL.Query())
	p, ok := c.GetQuery("price")
	if !ok {
		c.String(http.StatusBadRequest, "Fail at GetQuery(price)")
		return
	}
	price, err := strconv.Atoi(p)
	if err != nil {
		log.Fatal(err)
	}
	receipt, ok := c.GetQuery("receiptId")
	if !ok {
		c.String(http.StatusBadRequest, "Fail at GetQuery(receipt)")
		return
	}
	result := pg.VerifyReceipt(receipt, price)
	if result {
		c.String(http.StatusOK, "sp")
	} else {
		c.String(http.StatusOK, "jh")
	}

}

func onFeedBack(c *gin.Context) {
	//https://docs.bootpay.co.kr/rest/feedback
	data, _ := c.GetRawData()
	var objmap map[string]interface{}
	json.Unmarshal(data, &objmap)
	objmap["createdAt"] = time.Now()
	log.Printf("Received onFeedBack Data %v", objmap)
	inst := fire.GetFireInstance()
	store, _ := inst.Inst.Firestore(inst.Ctx)
	doc := store.Collection("bootpayFeedBack").NewDoc()
	// receipt_id := objmap["receipt_id"].(string)
	// doc := store.Collection("bootpayFeedBack").Doc(receipt_id)
	doc.Set(inst.Ctx, objmap)

	c.JSON(200, gin.H{
		"success": true,
	})
}
