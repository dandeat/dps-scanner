package mkpmobileutils

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/streadway/amqp"
)

// WORKER REST API
func WorkerRequestPOST(tipeRequest, urlApi string, requestBody interface{}, requestHeader ReqHeader) (result []byte, err error) {

	bodyRequest, _ := json.Marshal(requestBody)

	// CREATING REQUEST HTTP
	reqHTTP, err := http.NewRequest("POST", urlApi, bytes.NewBuffer(bodyRequest))
	if err != nil {
		return result, err
	}
	// END CREATING REQUEST HTTP

	reqHTTP = GenRequestHeader(reqHTTP, requestHeader)

	// Set Content-type header
	if tipeRequest == TIPE_REQUEST_URL_ENCODED {
		reqHTTP.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else if tipeRequest == TIPE_REQUEST_JSON {
		reqHTTP.Header.Add("Content-Type", "application/json")
	}
	reqHTTP.Header.Add("Content-Length", strconv.FormatInt(reqHTTP.ContentLength, 10))
	reqHTTP.Header.Set("Connection", "close")

	if bodyRequest != nil {
		defer reqHTTP.Body.Close()
	}
	reqHTTP.Close = true
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: TRUE_VALUE},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(reqHTTP)
	if err != nil {
		return result, err
	}
	resp.Header.Set("Connection", "close")
	defer resp.Body.Close()
	resp.Close = true

	result, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	return result, nil
}

func WorkerRequestPOSTJWT(tipeRequest, urlApi, accessToken string, requestBody interface{}) (result []byte, err error) {

	bodyRequest, _ := json.Marshal(requestBody)

	// CREATING REQUEST HTTP
	reqHTTP, err := http.NewRequest("POST", urlApi, bytes.NewBuffer(bodyRequest))
	if err != nil {
		return result, err
	}
	// END CREATING REQUEST HTTP

	requestHeader := ReqHeader{
		Header: []Header{
			{
				Key:      "Authorization",
				Val:      "Bearer " + accessToken,
				IsUpCase: false,
			},
		},
	}

	reqHTTP = GenRequestHeader(reqHTTP, requestHeader)

	// Set Content-type header
	if tipeRequest == TIPE_REQUEST_URL_ENCODED {
		reqHTTP.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else if tipeRequest == TIPE_REQUEST_JSON {
		reqHTTP.Header.Add("Content-Type", "application/json")
	}
	reqHTTP.Header.Add("Content-Length", strconv.FormatInt(reqHTTP.ContentLength, 10))
	reqHTTP.Header.Set("Connection", "close")

	if bodyRequest != nil {
		defer reqHTTP.Body.Close()
	}
	reqHTTP.Close = true
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: TRUE_VALUE},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(reqHTTP)
	if err != nil {
		return result, err
	}
	resp.Header.Set("Connection", "close")
	defer resp.Body.Close()
	resp.Close = true

	result, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	return result, nil
}

func WorkerRequestPOSTBasicAuth(tipeRequest, urlApi string, requestBody interface{}, requestHeader ReqHeader, basicAuth BasicAuth) (result []byte, err error) {

	bodyRequest, _ := json.Marshal(requestBody)

	// CREATING REQUEST HTTP
	reqHTTP, err := http.NewRequest("POST", urlApi, bytes.NewBuffer(bodyRequest))
	if err != nil {
		return result, err
	}
	// END CREATING REQUEST HTTP

	reqHTTP = GenRequestHeader(reqHTTP, requestHeader)
	reqHTTP.SetBasicAuth(basicAuth.Username, basicAuth.Password)

	// Set Content-type header
	if tipeRequest == "urlencoded" {
		reqHTTP.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else if tipeRequest == "json" {
		reqHTTP.Header.Add("Content-Type", "application/json")
	}

	if bodyRequest != nil {
		defer reqHTTP.Body.Close()
	}
	reqHTTP.Close = true
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: TRUE_VALUE},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(reqHTTP)
	if err != nil {
		return result, err
	}
	resp.Header.Set("Connection", "close")
	defer resp.Body.Close()
	resp.Close = true

	result, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	return result, nil
}

func GenRequestHeader(req *http.Request, reqHeader ReqHeader) *http.Request {

	for _, v := range reqHeader.Header {
		if v.IsUpCase {
			req.Header.Set(strings.ToUpper(v.Key), v.Val)
			continue
		}
		req.Header.Set(v.Key, v.Val)
	}

	return req
}

func GenerateRequestHeader(request *http.Request, requestHeader RequestHeaderBIFAST) *http.Request {

	if requestHeader.XSignature != EMPTY_VALUE {
		request.Header.Set("X-SIGNATURE", requestHeader.XSignature)
	}

	if requestHeader.XTimestamp != EMPTY_VALUE {
		request.Header.Set("X-TIMESTAMP", requestHeader.XTimestamp)
	}

	if requestHeader.XClientKey != EMPTY_VALUE {
		request.Header.Set("X-CLIENT-KEY", requestHeader.XClientKey)
	}

	if requestHeader.XIpAddress != EMPTY_VALUE {
		request.Header.Set("X-IP-ADDRESS", requestHeader.XIpAddress)
	}

	if requestHeader.XDeviceId != EMPTY_VALUE {
		request.Header.Set("X-DEVICE-ID", requestHeader.XDeviceId)
	}

	if requestHeader.XExternalId != EMPTY_VALUE {
		request.Header.Set("X-EXTERNAL-ID", requestHeader.XExternalId)
	}

	if requestHeader.XPartnerId != EMPTY_VALUE {
		request.Header.Set("X-PARTNER-ID", requestHeader.XPartnerId)
	}

	if requestHeader.ChannelId != EMPTY_VALUE {
		request.Header.Set("CHANNEL-ID", requestHeader.ChannelId)
	}

	if requestHeader.XLatitude != EMPTY_VALUE {
		request.Header.Set("X-LATITUDE", requestHeader.XLatitude)
	}

	if requestHeader.XLongitude != EMPTY_VALUE {
		request.Header.Set("X-LONGITUDE", requestHeader.XLongitude)
	}

	upperCaseHeader := make(http.Header)
	for key, value := range request.Header {
		upperCaseHeader[strings.ToUpper(key)] = value
	}
	request.Header = upperCaseHeader

	if requestHeader.AuthAccessToken != EMPTY_VALUE {
		request.Header.Set("Authorization", requestHeader.AuthAccessToken)
	}

	return request

}

// END WORKER REST API

// WORKER RABBITMQ
func WorkerRbtPublish(chn *amqp.Channel, exchange, queue, contentType string, mandatory, immediate bool, body interface{}) error {
	byteBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	err = chn.Publish(
		exchange,
		queue,
		mandatory,
		immediate,
		amqp.Publishing{
			ContentType: contentType,
			Body:        byteBody,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

// type WorkerRbtSubOpt struct {
// 	queueName string
// 	consumer  string
// 	autoAck   bool
// 	exclusive bool
// 	noLocal   bool
// 	noWait    bool
// 	arg       amqp.Table
// }

// func WorkerRbtSub(chn *amqp.Channel, opt WorkerRbtSubOpt) () {

// 	messages, err := chn.Consume(
// 		opt.queueName, // queue name
// 		opt.consumer,  // consumer
// 		opt.autoAck,   // auto-ack
// 		opt.exclusive, // exclusive
// 		opt.noLocal,   // no local
// 		opt.noWait,    // no wait
// 		opt.arg,       // arguments
// 	)
// 	if err != nil {

// 	}
// }

// END WORKER RABBITMQ
