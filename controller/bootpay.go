package controller

import (
	"log"
	"os"
	"strings"
	"sync"

	"net/http"
)

type BootPay struct {
	mode, pk, appId string
}

func (b BootPay) IsProd() bool {
	return b.mode == "production"
}
func (b BootPay) IsDev() bool {
	return b.mode == "development"
}

func (b BootPay) ApiUrl(uri []string, onlyProd bool) string {
	url := ""
	if onlyProd || b.IsProd() {
		url = "https://api.bootpay.co.kr"
	} else if b.IsDev() {
		url = "https://dev-api.bootpay.co.kr"
	} else {
		log.Panicln("neither IsDev or IsProd, check BootPay Mode Property")
	}

	return url + "/" + strings.Join(uri, "/")

}
func (b BootPay) AccessToken() (string, error) {
	// TODO: Manage Token to Memory
	payload := map[string]interface{}{
		"application_id": b.appId,
		"private_key":    b.pk,
	}
	apiUrl := b.ApiUrl([]string{"request", "token"}, true)
	resp, err := http.Post(apiUrl, "application/json", JsonDataToBuff(payload))
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		var objmap map[string]map[string]interface{}
		GetHttpJson(*resp, &objmap)
		return objmap["data"]["token"].(string), nil
	} else {
		var bytes []byte
		resp.Body.Read(bytes)
		if err != nil {
			log.Fatalln(err)
		}

		log.Printf("Fail BootPay Access Token, Url: %s, status: %v,  Info: %s", apiUrl, resp.Status, string(bytes))
		return "", err
	}

}

func (b BootPay) VerifyReceipt(receiptId string, price int) bool {
	token, err := b.AccessToken()
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("get BootPay Access Token: %v", token)
	apiUrl := b.ApiUrl([]string{"receipt", receiptId}, true)
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Add("Authorization", token)

	// Client객체에서 Request 실행
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	var objmap map[string]map[string]interface{}
	GetHttpJson(*resp, &objmap)
	log.Printf("input receipt: %s receipt: %s status: %v, price: %v, input Price: %v", objmap["data"]["receipt_id"], receiptId, objmap["data"]["status"], objmap["data"]["price"], price)
	status := int(objmap["data"]["status"].(float64))
	return (status == 0 || status == 1 || status == 2 || status == 3) &&
		int(objmap["data"]["price"].(float64)) == price &&
		objmap["data"]["receipt_id"].(string) == receiptId
}

func (b BootPay) Cancel(receiptId, name, reason string, price int) {
	payload := map[string]interface{}{
		"receipt_id": receiptId,
		"price":      price,
		"name":       name,
		"reason":     reason}

	apiUrl := b.ApiUrl([]string{"cancel.json"}, true)
	req, err := http.NewRequest("POST", apiUrl, JsonDataToBuff(payload))
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Add("Content-Type", "application/json")
	// FIXME
	req.Header.Add("Authorization", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	var bytes []byte
	resp.Body.Read(bytes)
	if err == nil {
		str := string(bytes)
		log.Println(str)
	}
}

// >>> Singleton  >>>
var once sync.Once
var instance *BootPay

// nil일 경우 new()로 생성하고 주소값을 반환
func GetBootPay() *BootPay {
	once.Do(func() { // <-- atomic, does not allow repeating
		mode := os.Getenv("MODE")
		instance = &BootPay{mode: mode, pk: os.Getenv("BOOTPAY_PK"), appId: os.Getenv("BOOTPAY_APP_ID")}
		log.Printf("Create new BootPay Instance, instance info: %+v ", instance)
	})
	return instance
}

// <<< Singleton  <<<
