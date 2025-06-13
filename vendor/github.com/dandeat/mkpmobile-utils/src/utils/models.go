package mkpmobileutils

import (
	"crypto/rsa"
	"time"
)

type TokenSignature struct {
	Username  string `json:"username"`
	Timestamp string `json:"timestamp"`
}

type DigitalSignature struct {
	HttpMethod  string      `json:"HTTPmethod"`
	EndpointUrl string      `json:"EndpointUrl"`
	AccessToken string      `json:"AccessToken"`
	RequestBody interface{} `json:"RequestBody"`
	Timestamp   string      `json:"Timestamp"`
}

type RequestHeaderBIFAST struct {
	ContentType     string `json:"Content-Type"`
	AuthAccessToken string `json:"Authorization"`
	XTimestamp      string `json:"X-TIMESTAMP"`
	XClientKey      string `json:"X-CLIENT-KEY"`
	XSignature      string `json:"X-SIGNATURE"`
	XIpAddress      string `json:"X-IP-ADDRESS"`
	XDeviceId       string `json:"X-DEVICE-ID"`
	XExternalId     string `json:"X-EXTERNAL-ID"`
	XPartnerId      string `json:"X-PARTNER-ID"`
	ChannelId       string `json:"CHANNEL-ID"`
	XLatitude       string `json:"X-LATITUDE"`
	XLongitude      string `json:"X-LONGITUDE"`
}

type ReqHeader struct {
	Header []Header
}

type Header struct {
	Key      string
	Val      string
	IsUpCase bool
}

type BasicAuth struct {
	Username string
	Password string
}

//Response Model
type Response struct {
	Meta             Meta        `json:"meta"`
	Result           interface{} `json:"result"`
	ResponseDatetime time.Time   `json:"responseDatetime"`
}
type Meta struct {
	Code          string   `json:"code"`
	Success       bool     `json:"success"`
	Message       string   `json:"message"`
	AdditionalMsg []string `json:"additionalMsg"`
}

// Token Signature
type Signer interface {
	SHA256withRSA(stringToSign string) string
}
type rsaPrivateKey struct {
	*rsa.PrivateKey
}

type ResponseV1 struct {
	ResponseCode     string      `json:"responseCode"`
	ResponseMessage  string      `json:"responseMessage"`
	ResponseDatetime time.Time   `json:"responseDateTime"`
	Result           interface{} `json:"result"`
}

type ResponseV1Enc struct {
	ResponseCode     string      `json:"responseCode"`
	ResponseMessage  string      `json:"responseMessage"`
	ResponseDatetime time.Time   `json:"responseDateTime"`
	Result           interface{} `json:"result"`
	MetaResult       string      `json:"metaResult"`
}
