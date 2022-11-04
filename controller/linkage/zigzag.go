package linkage

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// const signedDate = "1667413215384"

type ZigOrderResp struct {
	ItemList   []interface{} `json:"itemList" firestore:"itemList,omitempty"`
	TotalCount float64       `json:"totalCount" firestore:"totalCount,omitempty"`
}

const zigApiUrl = "https://openapi.zigzag.kr/1/graphql"

func GetZigzagOrders(accessKey, secretKey, signedDate string, dateFrom, dateTo int) (*ZigOrderResp, gin.H) {
	// date format: 20221029
	authHeader := getAuthHeader(orderQuery, signedDate, accessKey, secretKey)
	body, err := postZigzag(zigApiUrl,
		map[string]interface{}{
			"query": orderQuery,
			"variables": fmt.Sprintf(`{
				"status": "NEW_ORDER",
				"date_ymd_from": %d,
				"date_ymd_to": %d
			}`, dateFrom, dateTo)}, authHeader)
	if err != nil {
		return nil, gin.H{"err": err.Error()}
	}
	resp, err := parseOrderList(body)
	if err != nil {
		return nil, gin.H{"err": err.Error(), "body": string(body)}
	}
	return resp, nil
}

func parseOrderList(body []byte) (*ZigOrderResp, error) {
	var resData map[string]map[string]map[string]interface{}
	if err := json.Unmarshal(body, &resData); err != nil {
		return nil, err
	}
	if data, ok := resData["data"]; ok {
		var data map[string]interface{} = data["order_item_list"]
		item_list, ok := data["item_list"].([]interface{})
		if !ok {
			return nil, errors.New("zigzag order_list field item_list is nil")
		}

		total_count, ok := data["total_count"].(float64)
		if !ok {
			return nil, errors.New("zigzag order_list field total_count is nil")
		}

		return &ZigOrderResp{ItemList: item_list, TotalCount: total_count}, nil
	}
	return nil, nil
}

func postZigzag(url string, payload map[string]interface{}, authHeader string) ([]byte, error) {
	payloadBytes, _ := json.Marshal(payload)
	buff := bytes.NewBuffer(payloadBytes)

	req, err := http.NewRequest("POST", url, buff)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-solution", "io-box")

	req.Header.Add("Authorization", authHeader)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	} else if strings.Contains(string(body), "authenticated_failed") {
		return nil, errors.New("authenticated_failed")
	}

	return body, nil

}

func getAuthHeader(query string, signedDate string, accessKey string, secretKey string) string {
	m := regexp.MustCompile(`\s+`)
	headQuery := m.ReplaceAllString(query, " ")
	message := fmt.Sprintf("%s.%s", signedDate, string(headQuery))
	mac := hmac.New(sha1.New, []byte(secretKey))
	mac.Write([]byte(message))
	signature := hex.EncodeToString(mac.Sum(nil)) // 결과 가져오기 및 문자열로 전환
	log.Printf("signature: %s", signature)
	return fmt.Sprintf("CEA algorithm=HmacSHA256, access-key=%s, signed-date=%s, signature=%s", accessKey, signedDate, signature)
}

const orderQuery = `query (
  $status: OrderItemStatus
  $date_ymd_from: Int
  $date_ymd_to: Int
) {
  order_item_list(
    status: $status
    date_ymd_from: $date_ymd_from
    date_ymd_to: $date_ymd_to
  ) {
    total_count # int
    item_list {
      quantity # 수량 # int
      total_amount # 총 금액 # int
      unit_price # 단가 # int
      order_item_number # 주문 품목 번호 # string
      product_id # 상품번호 string
      product_item_id # 옵션번호 string
      product_info {
        # 상품 정보
        name # 상품명 # string
        options # 상품옵션 # string
        product_no # 상품번호 # string
        product_item_code # 옵션코드 # string # 옵션이 없는 상품이면 무시해도 됩니다
        option_detail_list {
          # 선택된 옵션 상세 정보 목록 # Array
          name # 옵션명 # string
          value # 옵션값 # string
        }
        product_code # 기본 상품 코드 # string
        custom_product_code # 판매자가 부여한 상품 코드 # string;
        custom_product_item_code # 판매자가 부여한 옵션 코드 # string;
      }
      status # 상태 # OrderItemStatus
      active_request {
        # 취소/반품 요청
        date_requested # 요청일자 (unix timestamp) # int
        type # 종류 # OrderItemRequestType
        status # 요청 상태 # OrderItemRequestStatus
        collecting_type # 상품 수거 방법 # CollectingType
        requested_reason_category # 취소/환불 이유  # RequestedReasonCategory
        requested_reason # 취소/환불 상세 이유
      }
      shipping_company # 택배사 # OrderShippingCompany
      shipping_fee_type # 배송비 결제 시점 # ShippingFeeType
      invoice_number # 운송장 번호 # string
      site # 상품이 주문된 # Site
      order_shop {
        # 주문 정보
        total_item_amount # 결제한 상품의 총 금액 # int
        total_shipping_fee # 결제시 지불한 총 배송비 # int
      }
      order {
        # 주문 정보
        order_number # 주문번호 # string
        date_created # 주문생성일자 (unix timestamp) # int
        payment_method # 결제수단 # PaymentMethod
        payment_status # 결제상태 # PaymentStatus
        shipping_memo # 배송메모 # string
        orderer {
          # 주문자 정보
          name # 이름 # string
          email # 이메일 # string
          mobile_tel # 전화번호 # string
        }
        receiver {
          # 수령자 정보
          name # 이름 # string
          mobile_tel # 전화번호 # string
          postcode # 우편번호 # string
          address1 # 주소 # string
          address2 # 상제주소 # string
        }
      }
    }
  }
}`
