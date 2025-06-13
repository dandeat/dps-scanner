package utils

import (
	"bytes"
	"crypto/tls"
	"dps-scanner-gateout/constants"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"time"

	mkpmobileutils "github.com/dandeat/mkpmobile-utils/src/utils"
)

func WorkerGetToken(body url.Values, basicAuthUsername string, basicAuthPassword string, suffixUrl string) ([]byte, int, error) {

	httpReq, err := http.NewRequest("POST", suffixUrl, bytes.NewBufferString(body.Encode()))
	if err != nil {
		log.Println("Err http.NewRequest ", suffixUrl, " :", err.Error())
		return nil, http.StatusBadRequest, err
	}

	defer httpReq.Body.Close()

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.SetBasicAuth(basicAuthUsername, basicAuthPassword)
	httpReq.Header.Add("Content-Length", strconv.Itoa(len(body.Encode())))
	httpReq.Header.Set("Connection", "close")

	// Put the body back for FormatRequest to read it
	for name, values := range httpReq.Header {
		// Loop over all values for the name.
		for _, value := range values {
			fmt.Println(name+" :", value)
		}
	}

	httpReq.Close = true

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, http.StatusBadGateway, err
	}

	resp.Header.Set("Connection", "close")
	defer resp.Body.Close()
	resp.Close = true

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	log.Println("Worker Request Url : ", suffixUrl)
	log.Println("Worker Request Data : ", body)
	log.Println("Worker Response Data : ", string(bodyBytes))

	return bodyBytes, resp.StatusCode, nil
}

func WorkerPostWithBearer(suffixUrl string, accessToken string, dataRequest interface{}) ([]byte, int, error) {

	bodyRequest, err := json.Marshal(dataRequest)
	if err != nil {
		fmt.Println("Err Worker Post - json.Marshal : ", err.Error())
		return nil, http.StatusBadRequest, err
	}

	httpReq, err := http.NewRequest("POST", suffixUrl, bytes.NewBuffer(bodyRequest))
	if err != nil {
		return nil, 0, err
	}
	defer httpReq.Body.Close()

	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Set("Connection", "close")

	if accessToken != constants.EMPTY_VALUE {
		httpReq.Header.Add("Authorization", "Bearer "+accessToken)
	}

	httpReq.Close = true

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, http.StatusBadGateway, err
	}

	resp.Header.Set("Connection", "close")
	defer resp.Body.Close()
	resp.Close = true

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	log.Println("Worker Request Url : ", suffixUrl)
	log.Println("Worker Request Data : ", string(bodyRequest))
	log.Println("Worker Response Data : ", string(bodyBytes))

	return bodyBytes, resp.StatusCode, nil
}

func WorkerPostWithBearerBINA(suffixUrl string, accessToken string, dataRequest interface{}) ([]byte, int, error) {

	bodyRequest, err := json.Marshal(dataRequest)
	if err != nil {
		fmt.Println("Err Worker Post - json.Marshal : ", err.Error())
		return nil, http.StatusBadRequest, err
	}

	httpReq, err := http.NewRequest("POST", suffixUrl, bytes.NewBuffer(bodyRequest))
	if err != nil {
		return nil, 0, err
	}
	defer httpReq.Body.Close()

	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Set("Connection", "close")

	if accessToken != constants.EMPTY_VALUE {
		httpReq.Header.Add("Authorization", "Bearer "+accessToken)
	}

	httpReq.Close = true

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, http.StatusBadGateway, err
	}

	resp.Header.Set("Connection", "close")
	defer resp.Body.Close()
	resp.Close = true

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	log.Println("Worker Request Url : ", suffixUrl)
	log.Println("Worker Request Data : ", string(bodyRequest))
	log.Println("Worker Response Data : ", string(bodyBytes))

	return bodyBytes, resp.StatusCode, nil
}

func WorkerRequestPOST(tipeRequest, urlApi string, requestBody interface{}, requestHeader mkpmobileutils.ReqHeader, timeout time.Duration, sessionId *http.Cookie) (result []byte, resp *http.Response, statusCode int, err error) {

	bodyRequest, _ := json.Marshal(requestBody)

	log.Printf("Request Body: %v\n", string(bodyRequest))

	// CREATING REQUEST HTTP
	reqHTTP, err := http.NewRequest("POST", urlApi, bytes.NewBuffer(bodyRequest))
	if err != nil {
		return result, resp, statusCode, err
	}
	// END CREATING REQUEST HTTP

	reqHTTP = mkpmobileutils.GenRequestHeader(reqHTTP, requestHeader)

	// Set Content-type header
	if tipeRequest == mkpmobileutils.TIPE_REQUEST_URL_ENCODED {
		reqHTTP.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else if tipeRequest == mkpmobileutils.TIPE_REQUEST_JSON {
		reqHTTP.Header.Add("Content-Type", "application/json")
	}
	reqHTTP.Header.Add("Content-Length", strconv.FormatInt(reqHTTP.ContentLength, 10))
	reqHTTP.Header.Set("Connection", "close")

	log.Printf("Request Header: %v\n", reqHTTP.Header)

	if bodyRequest != nil {
		defer reqHTTP.Body.Close()
	}
	reqHTTP.Close = true
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: constants.TRUE_VALUE},
	}


	client := &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}

	if sessionId != nil {

	cookie, _ := cookiejar.New(nil)
	u, _ := url.Parse("https://transmarine.oneerp.app")
	cookie.SetCookies(u, []*http.Cookie{
		sessionId,
	})

	client = &http.Client{
		Transport: tr,
		Timeout:   timeout,
		Jar:       cookie,
	}
}

	resp, err = client.Do(reqHTTP)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return result, resp, statusCode, err
	}
	resp.Header.Set("Connection", "close")
	defer resp.Body.Close()
	resp.Close = true

	statusCode = resp.StatusCode

	result, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return result, resp, statusCode, err
	}

	return result, resp, statusCode, nil
}

func WorkerRequestPOSTStatusCode(tipeRequest, urlApi string, requestBody interface{}, requestHeader mkpmobileutils.ReqHeader) (result []byte, statusCode int, err error) {

	bodyRequest, _ := json.Marshal(requestBody)

	// CREATING REQUEST HTTP
	reqHTTP, err := http.NewRequest("POST", urlApi, bytes.NewBuffer(bodyRequest))
	if err != nil {
		return result, statusCode, err
	}
	// END CREATING REQUEST HTTP

	reqHTTP = mkpmobileutils.GenRequestHeader(reqHTTP, requestHeader)

	// Set Content-type header
	if tipeRequest == mkpmobileutils.TIPE_REQUEST_URL_ENCODED {
		reqHTTP.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else if tipeRequest == mkpmobileutils.TIPE_REQUEST_JSON {
		reqHTTP.Header.Add("Content-Type", "application/json")
	}
	reqHTTP.Header.Add("Content-Length", strconv.FormatInt(reqHTTP.ContentLength, 10))
	reqHTTP.Header.Set("Connection", "close")

	log.Printf("Request Header: %v\n", reqHTTP.Header)

	if bodyRequest != nil {
		defer reqHTTP.Body.Close()
	}
	reqHTTP.Close = true
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: constants.TRUE_VALUE},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(reqHTTP)
	if err != nil {
		return result, statusCode, err
	}
	resp.Header.Set("Connection", "close")
	defer resp.Body.Close()
	resp.Close = true

	statusCode = resp.StatusCode

	result, err = io.ReadAll(resp.Body)
	if err != nil {
		return result, statusCode, err
	}

	return result, statusCode, nil
}

func WorkerRequestPOSTGateaway(tipeRequest, urlApi string, requestBody []byte, requestHeader mkpmobileutils.ReqHeader, additionalData ...interface{}) (result []byte, statusCode int, err error) {
	/* additionalData
	0 : Content-Type | string
	*/
	var reqHTTP *http.Request
	if tipeRequest == constants.REQ_FORM_DATA {

		// Create the new request with the multipart body
		reqHTTP, err = http.NewRequest("POST", urlApi, bytes.NewBuffer(requestBody))
		if err != nil {
			return nil, 0, fmt.Errorf("error creating new request: %v", err)
		}

		reqHTTP = mkpmobileutils.GenRequestHeader(reqHTTP, requestHeader)
		if len(additionalData) < 1 {
			return nil, 0, fmt.Errorf("error getting additional data")
		} else {
			reqHTTP.Header.Set("Content-Type", additionalData[0].(string))
		}

	} else {
		// CREATING REQUEST HTTP
		reqHTTP, err = http.NewRequest("POST", urlApi, bytes.NewBuffer(requestBody))
		if err != nil {
			return result, statusCode, err
		}
		// END CREATING REQUEST HTTP

		reqHTTP = mkpmobileutils.GenRequestHeader(reqHTTP, requestHeader)

		// Set Content-type header
		if tipeRequest == mkpmobileutils.TIPE_REQUEST_URL_ENCODED {
			reqHTTP.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		} else if tipeRequest == mkpmobileutils.TIPE_REQUEST_JSON {
			reqHTTP.Header.Add("Content-Type", "application/json")
		}

	}

	reqHTTP.Header.Add("Content-Length", strconv.FormatInt(reqHTTP.ContentLength, 10))
	reqHTTP.Header.Set("Connection", "close")

	log.Printf("Request Header: %v\n", reqHTTP.Header)

	if requestBody != nil {
		defer reqHTTP.Body.Close()
	}
	reqHTTP.Close = true
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: constants.TRUE_VALUE},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(reqHTTP)
	if err != nil {
		return result, statusCode, err
	}
	resp.Header.Set("Connection", "close")
	defer resp.Body.Close()
	resp.Close = true

	statusCode = resp.StatusCode

	result, err = io.ReadAll(resp.Body)
	if err != nil {
		return result, statusCode, err
	}

	return result, statusCode, nil
}
