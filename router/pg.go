package router

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"

	"cloud.google.com/go/firestore"
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
	// objmap["createdAt"] = time.Now()
	inst := fire.GetFireInstance()
	store, _ := inst.Inst.Firestore(inst.Ctx)
	doc := store.Collection("bootpayFeedBack").NewDoc()
	doc.Set(inst.Ctx, objmap)

	method, exists := objmap["method_origin_symbol"].(string)
	if !exists {
		method = ""
	}
	status, exists := objmap["status"].(float64)
	if !exists {
		status = -1
	}
	userId, exists := objmap["metadata"].(map[string]interface{})["uid"].(string)
	if !exists {
		userId = ""
	}

	if method == "vbank" && status == 1 {
		// 가상계좌 - 입금완료 처리
		price := objmap["price"].(float64)
		dsnap, err := store.Collection("ioPay").Doc(userId).Get(inst.Ctx)
		if !dsnap.Exists() {
			log.Fatalf("userId(%s) ioPay doc not exist", userId)
		} else if err != nil {
			log.Fatalln(err.Error())
		}
		data := dsnap.Data()
		budget := uint16(data["budget"].(int64))
		newCoin := moneyToCoin(price)
		newBudget := newCoin + budget
		log.Printf("userId: %s before budget: %d price: %d required coin to: %d", userId, uint16(budget), uint16(price), uint16(newBudget))

		_, err = store.Collection("ioPay").Doc(userId).Update(inst.Ctx, []firestore.Update{{Path: "budget", Value: newBudget}})
		if err != nil {
			log.Fatalf("ioPay update error: %s", err.Error())
		}
		err = ctr.IoSendMail([]string{userId}, title, chargeHtmlBody(newCoin))
		if err != nil {
			log.Printf("Couldn't send email: %v", err.Error())
		}
		err = ctr.IoSendPush([]string{userId}, "", []string{}, map[string]string{}, title, fmt.Sprintf("%d개 코인이 충전 되었습니다.", newCoin))
		if err != nil {
			log.Printf("Couldn't send push: %v", err.Error())
		}
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}

const COIN_PAY_RATIO = 10
const title = "인아웃코인 충전완료 알림"

func moneyToCoin(money float64) uint16 {
	return uint16(math.Round(money / COIN_PAY_RATIO))
}

func chargeHtmlBody(coin uint16) string {
	return fmt.Sprintf(`
		<p>%d개 코인이 충전 되었습니다.</p>
		<br />
		처리된 내용에 문의가 있으실 경우 고객센터로 문의 주시면 감사하겠습니다.
		<br />
		해당 메일은 회신이 불가한 메일입니다.
	`, coin)
}
