package token

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"time"

	b64 "encoding/base64"

	"github.com/gin-gonic/gin"
	cnt "github.com/io-boxies/io-app-engine/controller"
	"github.com/io-boxies/io-app-engine/controller/fire"
)

const cafeClientId = "mnhAX4sDM9UmCchzOwzTAA"
const cafeSecret = "8KAAKdXQLtmgAo0dBt8avC"

func SaveCafeToken(code string, redirectUri string, mallId string, userId string) (int, gin.H) {

	payload := url.Values{}
	payload.Set("grant_type", "authorization_code")
	payload.Set("code", code)
	payload.Set("redirect_uri", redirectUri)
	url := fmt.Sprintf("https://%s.cafe24api.com/api/v2/oauth/token", mallId)
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(payload.Encode()))
	if err != nil {
		return 500, gin.H{"err": err.Error()}
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	sEnc := b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", cafeClientId, cafeSecret)))
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", sEnc))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode >= 400 {
		errObj := make(gin.H)
		cnt.GetHttpJson(*resp, &errObj)
		return resp.StatusCode, errObj
	}
	defer resp.Body.Close()

	var objmap map[string]interface{}
	cnt.GetHttpJson(*resp, &objmap)
	layout := "2006-01-02T15:04:05.000"
	expireAt, _ := time.Parse(layout, objmap["expires_at"].(string))
	refreshExpireAt, _ := time.Parse(layout, objmap["refresh_token_expires_at"].(string))
	issuedAt, _ := time.Parse(layout, objmap["issued_at"].(string))
	var scopes []string
	for _, scope := range objmap["scopes"].([]interface{}) {
		scopes = append(scopes, scope.(string))
	}
	token := &OAuthToken{
		CreatedAt:       issuedAt,
		AccessToken:     objmap["access_token"].(string),
		ExpireAt:        expireAt,
		RefreshToken:    objmap["refresh_token"].(string),
		RefreshExpireAt: refreshExpireAt,
		ClientId:        objmap["client_id"].(string),
		MallId:          objmap["mall_id"].(string),
		UserId:          objmap["user_id"].(string),
		Scopes:          scopes,
	}
	cafeToken := NewCafeToken(*token, objmap["shop_no"].(string))
	inst := fire.GetFireInstance()
	store, _ := inst.Inst.Firestore(inst.Ctx)
	cPath := fmt.Sprintf("user/%s/tokens", userId)
	_, err = store.Collection(cPath).Doc("cafe").Set(inst.Ctx, cafeToken)
	if err != nil {
		return 500, gin.H{"err": err.Error()}
	}
	return 200, nil
}

func GetCafeOrders(mallId string, userId string, startDate string, endDate string) (interface{}, gin.H) {
	inst := fire.GetFireInstance()
	store, _ := inst.Inst.Firestore(inst.Ctx)
	cPath := fmt.Sprintf("user/%s/tokens", userId)
	dsnap, err := store.Collection(cPath).Doc("cafe").Get(inst.Ctx)
	if err != nil {
		return nil, gin.H{"err": err.Error()}
	}
	var token CafeToken
	dsnap.DataTo(&token)
	fmt.Printf("Cafe Token data: %#v\n", token)

	url := fmt.Sprintf("https://%s.cafe24api.com/api/v2/admin/orders?start_date=%s&end_date=%s&embed=items,cancellation,return,exchange", mallId, startDate, endDate)
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Cafe24-Api-Version", "2022-06-01")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode >= 400 {
		errObj := make(gin.H)
		cnt.GetHttpJson(*resp, &errObj)
		return nil, errObj
	}

	var objmap map[string]interface{}
	cnt.GetHttpJson(*resp, &objmap)
	return objmap, nil

}
