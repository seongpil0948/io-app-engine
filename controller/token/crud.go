package token

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	b64 "encoding/base64"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	cnt "github.com/io-boxies/io-app-engine/controller"
	"github.com/io-boxies/io-app-engine/controller/fire"
	"google.golang.org/api/iterator"
)

const cafeClientId = "mnhAX4sDM9UmCchzOwzTAA"
const cafeSecret = "8KAAKdXQLtmgAo0dBt8avC"

func saveFirestoreCafeToken(resp *http.Response, userId string, docId string) error {
	defer resp.Body.Close()
	var objmap map[string]interface{}
	cnt.GetHttpJson(*resp, &objmap)
	if objmap == nil {
		return errors.New("objmap is nil")
	}
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
		UserId:          userId,
		Scopes:          scopes,
		Service:         "CAFE",
		ServiceId:       objmap["user_id"].(string),
	}
	cafeToken := NewCafeToken(*token, objmap["shop_no"].(string))
	inst := fire.GetFireInstance()
	store, _ := inst.Inst.Firestore(inst.Ctx)
	cPath := fmt.Sprintf("user/%s/tokens", userId)
	_, err := store.Collection(cPath).Doc(docId).Set(inst.Ctx, cafeToken)
	return err
}

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

	err = saveFirestoreCafeToken(resp, userId, uuid.New().String())
	if err != nil {
		return 500, gin.H{"err": err.Error()}
	}
	return 200, nil
}

func RefreshTokens() error {
	inst := fire.GetFireInstance()
	store, _ := inst.Inst.Firestore(inst.Ctx)
	iter := store.CollectionGroup("tokens").Documents(inst.Ctx)
	client := &http.Client{}
	cafePayload := url.Values{}
	cafePayload.Set("grant_type", "refresh_token")
	sEnc := b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", cafeClientId, cafeSecret)))

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return fmt.Errorf("error in get token List: %#v", err.Error())
		}
		var token CafeToken
		doc.DataTo(&token)
		log.Printf("Try refresh Doc: %s ", doc.Ref.ID)
		if token.Service == "CAFE" {
			log.Println("In refresh cafe token Scope")
			url := fmt.Sprintf("https://%s.cafe24api.com/api/v2/oauth/token", token.MallId)
			cafePayload.Set("refresh_token", token.RefreshToken)
			req, _ := http.NewRequest("POST", url, bytes.NewBufferString(cafePayload.Encode()))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Add("Authorization", fmt.Sprintf("Basic %s", sEnc))
			resp, err := client.Do(req)
			if err != nil || resp.StatusCode >= 400 {
				errObj := make(gin.H)
				cnt.GetHttpJson(*resp, &errObj)
				if errObj != nil {

					return fmt.Errorf("error in cafe refresh API: %#v", errObj)
				}
				return fmt.Errorf("error in cafe refresh API: %#v", err.Error())
			}
			defer resp.Body.Close()
			err = saveFirestoreCafeToken(resp, token.UserId, doc.Ref.ID)
			log.Println("refresh done")
			if err != nil {
				return fmt.Errorf("error in save refresh token: %#v", err.Error())
			}
		}

	}
	return nil
}

func GetCafeOrders(mallId string, userId string, startDate string, endDate string, tokenId string) (interface{}, gin.H) {
	inst := fire.GetFireInstance()
	store, _ := inst.Inst.Firestore(inst.Ctx)
	cPath := fmt.Sprintf("user/%s/tokens", userId)
	dsnap, err := store.Collection(cPath).Doc(tokenId).Get(inst.Ctx)
	if !dsnap.Exists() {
		return nil, gin.H{"err": "doc not exist"}
	} else if err != nil {
		return nil, gin.H{"err": err.Error()}
	}
	var token CafeToken
	dsnap.DataTo(&token)
	// fmt.Printf("Cafe token data: %#v\n", token)

	url := fmt.Sprintf("https://%s.cafe24api.com/api/v2/admin/orders?start_date=%s&end_date=%s&embed=items,cancellation,return,exchange", mallId, startDate, endDate)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Cafe24-Api-Version", "2022-06-01")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode >= 400 {
		errObj := make(gin.H)
		cnt.GetHttpJson(*resp, &errObj)
		log.Printf("#%v", errObj)
		return nil, errObj
	}

	var objmap map[string]interface{}
	cnt.GetHttpJson(*resp, &objmap)
	return objmap, nil

}
