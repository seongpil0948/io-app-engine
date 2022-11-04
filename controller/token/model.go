package token

import (
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"github.com/io-boxies/io-app-engine/controller/fire"
)

type IoAuthTokenFace interface {
	refresh()
	isExpired()
}

// https://stackoverflow.com/questions/38809137/golang-multiple-json-tag-names-for-one-field
type IoAuthToken struct {
	CreatedAt       time.Time `json:"createdAt" firestore:"createdAt,omitempty"`
	AccessToken     string    `json:"access_token" firestore:"accessToken,omitempty"`
	ExpireAt        time.Time `json:"expireAt" firestore:"expireAt,omitempty"`
	RefreshToken    string    `json:"refresh_token" firestore:"refreshToken,omitempty"`
	RefreshExpireAt time.Time `json:"refreshExpireAt" firestore:"refreshExpireAt,omitempty"`
	ClientId        string    `json:"client_id" firestore:"clientId,omitempty"`
	MallId          string    `json:"mall_id" firestore:"mallId,omitempty"`
	UserId          string    `json:"user_id" firestore:"userId,omitempty"`
	Scopes          []string  `json:"scopes" firestore:"scopes,omitempty"`
	Service         string    `json:"service" firestore:"service,omitempty"`
	ServiceId       string    `json:"service_id" firestore:"serviceId,omitempty"`
	Alias           string    `json:"alias" firestore:"alias,omitempty"`
	AccessKey       string    `json:"accessKey" firestore:"accessKey,omitempty"`
	SecretKey       string    `json:"secretKey" firestore:"secretKey,omitempty"`
}

type CafeToken struct {
	IoAuthToken
	ShopNo string `json:"shopNo" firestore:"shopNo,omitempty"`
}

func NewCafeToken(token IoAuthToken, shopNo string) *CafeToken {
	return &CafeToken{
		IoAuthToken: token,
		ShopNo:      shopNo,
	}
}

func GetToken(userId, tokenId string) (*firestore.DocumentSnapshot, gin.H) {
	inst := fire.GetFireInstance()
	store, _ := inst.Inst.Firestore(inst.Ctx)
	cPath := fmt.Sprintf("user/%s/tokens", userId)
	dsnap, err := store.Collection(cPath).Doc(tokenId).Get(inst.Ctx)
	if !dsnap.Exists() {
		return nil, gin.H{"err": "doc not exist"}
	} else if err != nil {
		return nil, gin.H{"err": err.Error()}
	}
	return dsnap, nil
}
