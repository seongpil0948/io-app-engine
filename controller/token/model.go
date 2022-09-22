package token

import "time"

type OAuthTokenFace interface {
	refresh()
	isExpired()
}

// https://stackoverflow.com/questions/38809137/golang-multiple-json-tag-names-for-one-field
type OAuthToken struct {
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
}

type CafeToken struct {
	OAuthToken
	ShopNo string `json:"shopNo" firestore:"shopNo,omitempty"`
}

func NewCafeToken(token OAuthToken, shopNo string) *CafeToken {
	return &CafeToken{
		OAuthToken: token,
		ShopNo:     shopNo,
	}
}
