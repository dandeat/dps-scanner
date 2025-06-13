package models

import (
	"encoding/json"
	"fmt"
	"time"
)

type ResponseJSON struct {
	StatusCode       string      `json:"statusCode"`
	Success          bool        `json:"success"`
	ResponseDatetime time.Time   `json:"responseDatetime"`
	Result           interface{} `json:"result"`
	Message          string      `json:"message"`
}

type ResponseV1 struct {
	ResponseCode     string      `json:"responseCode"`
	ResponseMessage  string      `json:"responseMessage"`
	ResponseDatetime time.Time   `json:"responseDateTime"`
	Result           interface{} `json:"result"`
}

type ResponseModuleBINA struct {
	ResponseCode          string          `json:"responseCode"`
	ResponseMessage       string          `json:"responseMessage"`
	ResponseCodeDetail    string          `json:"responseCodeDetail"`
	ResponseMessageDetail string          `json:"responseMessageDetail"`
	ResponseDatetime      time.Time       `json:"responseDateTime"`
	Result                json.RawMessage `json:"result"`
}

// JSONTime ..
type JSONTime time.Time

// MarshalJSON ..
func (t JSONTime) MarshalJSON() ([]byte, error) {
	//do your serializing here
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format("2006-01-02 15:04:05"))
	return []byte(stamp), nil
}

type TokenDetail struct {
	PersonID   string `json:"personID" bson:"personID"`
	PersonName string `json:"personName" bson:"personName"`
	CID        string `json:"cid" bson:"cid"`
	Hirarki    string `json:"hirarki" bson:"hirarki"`
	PersonType string `json:"personType" bson:"personType"`
}
