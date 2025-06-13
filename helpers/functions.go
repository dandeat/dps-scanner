package helpers

import (
	"bytes"
	"crypto/tls"
	"dps-scanner-gateout/constants"
	"dps-scanner-gateout/models"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	httpstat "github.com/tcnksm/go-httpstat"
)

// WORKER GET REQUEST WITHOUT SIGNATURE
func WorkerRequestGET(urlApi string) (result []byte, err error) {
	req, err := http.NewRequest("GET", urlApi, nil)
	if err != nil {
		return result, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("status is not ok: %v", resp.StatusCode)
	}

	respByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	return respByte, nil
}

func WorkerRequestGETWithSignature(urlApi, sign string) (result []byte, err error) {
	req, err := http.NewRequest("GET", urlApi, nil)
	if err != nil {
		return result, err
	}

	req.Header.Set("Connection", "close")

	if sign != constants.EMPTY_VALUE {
		req.Header.Set("Signature", sign)
	}

	req.Close = true
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("status is not ok :> %v", resp.StatusCode)
	}

	resp.Header.Set("Connection", "close")
	respByte, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	resp.Close = true
	// resp.Close = true

	return respByte, nil
}

// WORKER REST API
func WorkerRequestPOST(tipeRequest, urlApi string, requestBody interface{}, requestHeader models.ReqHeader, timeout time.Duration) (result []byte, code string, err error) {

	bodyRequest, _ := json.Marshal(requestBody)

	// CREATING REQUEST HTTP
	reqHTTP, err := http.NewRequest("POST", urlApi, bytes.NewBuffer(bodyRequest))
	if err != nil {
		return result, "", err
	}
	// END CREATING REQUEST HTTP

	var resultStat httpstat.Result
	ctx := httpstat.WithHTTPStat(reqHTTP.Context(), &resultStat)
	reqHTTP = reqHTTP.WithContext(ctx)

	reqHTTP = GenRequestHeader(reqHTTP, requestHeader)

	// Set Content-type header
	if tipeRequest == constants.REQ_URL_ENCODED {
		reqHTTP.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else if tipeRequest == constants.REQ_JSON {
		reqHTTP.Header.Add("Content-Type", "application/json")
	}
	reqHTTP.Header.Add("Content-Length", strconv.FormatInt(reqHTTP.ContentLength, 10))
	reqHTTP.Header.Set("Connection", "close")
	// ts := http.Header{
	// 	"HEAFES": []string{"assa"},
	// }
	// reqHTTP.Header.
	// val, _ := json.Marshal(reqHTTP.Header)
	log.Println("Req Header :>", reqHTTP.Header)

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
	resp, err := client.Do(reqHTTP)
	if err != nil {
		return result, "", err
	}
	resp.Header.Set("Connection", "close")
	defer resp.Body.Close()
	resp.Close = true

	result, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, strconv.Itoa(resp.StatusCode), err
	}

	// if resp.StatusCode != http.StatusOK {
	// 	return result, strconv.Itoa(resp.StatusCode), fmt.Errorf("status is not ok :> %v", resp.StatusCode)
	// }
	log.Printf("DNS lookup: %d ms", int(resultStat.DNSLookup/time.Millisecond))
	log.Printf("TCP connection: %d ms", int(resultStat.TCPConnection/time.Millisecond))
	log.Printf("TLS handshake: %d ms", int(resultStat.TLSHandshake/time.Millisecond))
	log.Printf("Server processing: %d ms", int(resultStat.ServerProcessing/time.Millisecond))

	return result, strconv.Itoa(resp.StatusCode), nil
}

// WORKER REST API Without header
func WorkerRequestPOST2(tipeRequest, urlApi string, requestBody interface{}, timeout time.Duration) (result []byte, code string, err error) {

	bodyRequest, _ := json.Marshal(requestBody)

	// CREATING REQUEST HTTP
	reqHTTP, err := http.NewRequest("POST", urlApi, bytes.NewBuffer(bodyRequest))
	if err != nil {
		return result, "", err
	}
	// END CREATING REQUEST HTTP

	var resultStat httpstat.Result
	ctx := httpstat.WithHTTPStat(reqHTTP.Context(), &resultStat)
	reqHTTP = reqHTTP.WithContext(ctx)

	// Set Content-type header
	if tipeRequest == constants.REQ_URL_ENCODED {
		reqHTTP.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else if tipeRequest == constants.REQ_JSON {
		reqHTTP.Header.Add("Content-Type", "application/json")
	}
	reqHTTP.Header.Add("Content-Length", strconv.FormatInt(reqHTTP.ContentLength, 10))
	reqHTTP.Header.Set("Connection", "close")
	// ts := http.Header{
	// 	"HEAFES": []string{"assa"},
	// }
	// reqHTTP.Header.
	// val, _ := json.Marshal(reqHTTP.Header)
	log.Println("Req Header :>", reqHTTP.Header)

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
	resp, err := client.Do(reqHTTP)
	if err != nil {
		return result, "", err
	}
	resp.Header.Set("Connection", "close")
	defer resp.Body.Close()
	resp.Close = true

	result, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, strconv.Itoa(resp.StatusCode), err
	}

	// if resp.StatusCode != http.StatusOK {
	// 	return result, strconv.Itoa(resp.StatusCode), fmt.Errorf("status is not ok :> %v", resp.StatusCode)
	// }
	log.Printf("DNS lookup: %d ms", int(resultStat.DNSLookup/time.Millisecond))
	log.Printf("TCP connection: %d ms", int(resultStat.TCPConnection/time.Millisecond))
	log.Printf("TLS handshake: %d ms", int(resultStat.TLSHandshake/time.Millisecond))
	log.Printf("Server processing: %d ms", int(resultStat.ServerProcessing/time.Millisecond))

	return result, strconv.Itoa(resp.StatusCode), nil
}

// Worker Rest Api body url
func WorkerRequestPOST3(tipeRequest, urlApi string, body url.Values, requestHeader models.ReqHeader, timeout time.Duration) ([]byte, int, error) {

	var resultStat httpstat.Result

	httpReq, err := http.NewRequest("POST", urlApi, bytes.NewBufferString(body.Encode()))
	if err != nil {
		log.Println("Err http.NewRequest ", urlApi, " :", err.Error())
		return nil, http.StatusBadRequest, err
	}

	ctx := httpstat.WithHTTPStat(httpReq.Context(), &resultStat)
	httpReq = httpReq.WithContext(ctx)

	httpReq = GenRequestHeader(httpReq, requestHeader)

	defer httpReq.Body.Close()

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Add("Content-Length", strconv.Itoa(len(body.Encode())))
	httpReq.Header.Set("Connection", "close")

	// Put the body back for FormatRequest to read it
	for name, values := range httpReq.Header {
		// Loop over all values for the name.
		for _, value := range values {
			fmt.Println(name+" :", value)
		}
	}

	log.Println("Req Header :>", httpReq.Header)

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

	log.Printf("DNS lookup: %d ms", int(resultStat.DNSLookup/time.Millisecond))
	log.Printf("TCP connection: %d ms", int(resultStat.TCPConnection/time.Millisecond))
	log.Printf("TLS handshake: %d ms", int(resultStat.TLSHandshake/time.Millisecond))
	log.Printf("Server processing: %d ms", int(resultStat.ServerProcessing/time.Millisecond))

	return bodyBytes, resp.StatusCode, nil
}

// Worker Rest Api body url without header
func WorkerRequestPOST4(tipeRequest, urlApi string, body url.Values, timeout time.Duration) ([]byte, int, error) {

	var resultStat httpstat.Result

	httpReq, err := http.NewRequest("POST", urlApi, bytes.NewBufferString(body.Encode()))
	if err != nil {
		log.Println("Err http.NewRequest ", urlApi, " :", err.Error())
		return nil, http.StatusBadRequest, err
	}

	ctx := httpstat.WithHTTPStat(httpReq.Context(), &resultStat)
	httpReq = httpReq.WithContext(ctx)

	defer httpReq.Body.Close()

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Add("Content-Length", strconv.Itoa(len(body.Encode())))
	httpReq.Header.Set("Connection", "close")

	// Put the body back for FormatRequest to read it
	for name, values := range httpReq.Header {
		// Loop over all values for the name.
		for _, value := range values {
			fmt.Println(name+" :", value)
		}
	}

	log.Println("Req Header :>", httpReq.Header)

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

	log.Printf("DNS lookup: %d ms", int(resultStat.DNSLookup/time.Millisecond))
	log.Printf("TCP connection: %d ms", int(resultStat.TCPConnection/time.Millisecond))
	log.Printf("TLS handshake: %d ms", int(resultStat.TLSHandshake/time.Millisecond))
	log.Printf("Server processing: %d ms", int(resultStat.ServerProcessing/time.Millisecond))

	return bodyBytes, resp.StatusCode, nil
}

func GenRequestHeader(req *http.Request, reqHeader models.ReqHeader) *http.Request {

	for _, v := range reqHeader.Header {
		if v.IsUpCase {
			req.Header[strings.ToUpper(v.Key)] = []string{v.Val}
			continue
		}
		req.Header.Set(v.Key, v.Val)
	}

	return req
}
