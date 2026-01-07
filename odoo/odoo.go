package odoo

import (
	"bytes"
	"crypto/tls"
	"dps-scanner-gateout/models"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect" // ✅ added
	"time"
	"math"
	"strconv"
)

// Client is a client for interacting with the Odoo API.
type Client struct {
	baseURL  string
	db       string
	username string
	password string
	client   *http.Client
}

// NewClient creates a new Odoo API client.
func NewClient(baseURL, db, username, password string) *Client {
	return &Client{
		baseURL:  baseURL,
		db:       db,
		username: username,
		password: password,
		client: &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Use only for development
			},
		},
	}
}

// --- Internal Helper for making requests ---

// safeInt tries to coerce v into an int if it represents an integral value.
func safeInt(v any) (int, bool) {
	switch t := v.(type) {
	case int:
		return t, true
	case int8, int16, int32, int64:
		return int(reflect.ValueOf(t).Int()), true
	case uint, uint8, uint16, uint32, uint64:
		u := reflect.ValueOf(t).Uint()
		if u > math.MaxInt {
			return 0, false
		}
		return int(u), true
	case float64:
		if math.Trunc(t) != t { // reject 12.34
			return 0, false
		}
		if t < math.MinInt || t > math.MaxInt {
			return 0, false
		}
		return int(t), true
	case json.Number:
		// First try integer parse
		if i64, err := strconv.ParseInt(string(t), 10, 64); err == nil {
			return int(i64), true
		}
		// Then try float but require integral
		if f64, err := strconv.ParseFloat(string(t), 64); err == nil && math.Trunc(f64) == f64 {
			return int(f64), true
		}
		return 0, false
	case string:
		// allow "123"
		if i64, err := strconv.ParseInt(t, 10, 64); err == nil {
			return int(i64), true
		}
		return 0, false
	default:
		return 0, false
	}
}

// m2oID extracts an Odoo Many2one id from v which can be:
// - []any{id, display_name}
// - scalar id (float64/string/json.Number/int…)
// - false / nil  -> no value
func m2oID(v any) (int, bool) {
	if v == nil {
		return 0, false
	}
	// Odoo false for missing m2o
	if b, ok := v.(bool); ok && !b {
		return 0, false
	}
	if arr, ok := v.([]any); ok {
		if len(arr) == 0 {
			return 0, false
		}
		return safeInt(arr[0])
	}
	return safeInt(v)
}

// workerRequest sends a POST request and handles the response.
func (c *Client) workerRequest(url string, payload interface{}, sessionCookie *http.Cookie) ([]byte, *http.Response, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if sessionCookie != nil {
		req.AddCookie(sessionCookie)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return responseBody, resp, fmt.Errorf("API returned non-200 status: %d", resp.StatusCode)
	}

	return responseBody, resp, nil
}

// authenticate gets a session cookie from Odoo.
func (c *Client) authenticate() (*http.Cookie, error) {
	authURL := fmt.Sprintf("%s/web/session/authenticate", c.baseURL)
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "call",
		"params": map[string]string{
			"db":       c.db,
			"login":    c.username,
			"password": c.password,
		},
	}

	_, resp, err := c.workerRequest(authURL, payload, nil)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "session_id" {
			return cookie, nil
		}
	}
	return nil, errors.New("session_id cookie not found in auth response")
}

// --- Public Methods for Business Logic ---

// CallScanMuat performs the main scan logic by calling the Odoo API.
func (c *Client) CallScanMuat(muatID, barcode string, filterID int) (models.DataMuatScan, error) {
	var atensi = false

	res := models.DataMuatScan{
		Barcode:    barcode,
		StatusScan: "error",
		StatusDesc: "Terjadi Kesalahan, Silahkan Coba Lagi.",
	}

	sessionCookie, err := c.authenticate()
	if err != nil {
		return res, fmt.Errorf("authentication failed: %w", err)
	}

	// 1. Search for the package
	searchPayload := map[string]interface{}{
		"params": map[string]interface{}{
			"model":  "dps.kemasan",
			"method": "search_read",
			"args":   []interface{}{},
			"kwargs": map[string]interface{}{
				"domain": [][]interface{}{{"name", "=", barcode}},
				"fields": []string{"id", "cn_id", "name", "no_sppb", "waktu_gatein", "hasil_periksa", "provinsi_penerima", "kode_agen", "note", "waktu_gateout", "kriteria_muat_id"},
				"limit":  1,
			},
		},
	}

	callURL := fmt.Sprintf("%s/web/dataset/call_kw", c.baseURL)
	body, _, err := c.workerRequest(callURL, searchPayload, sessionCookie)
	if err != nil {
		return res, fmt.Errorf("package search request failed: %w", err)
	}

	var searchResp struct {
		Result []struct {
			KodeAgen         any    `json:"kode_agen"`
			Name             string `json:"name"`
			CnID			 []any  `json:"cn_id"`
			HasilPeriksa     any    `json:"hasil_periksa"`
			ProvinsiPenerima any    `json:"provinsi_penerima"`
			ID               int    `json:"id"`
			NoSPPB           any    `json:"no_sppb"`
			TglSPPB		     any    `json:"tgl_sppb"`
			Note             any    `json:"note"`
			WaktuGateIn      string `json:"waktu_gatein"`
			WaktuGateOut     any    `json:"waktu_gateout"`
			KriteriaMuatID   any    `json:"kriteria_muat_id"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &searchResp); err != nil {
		return res, fmt.Errorf("failed to unmarshal search response: %w", err)
	}

	log.Println("Get Kemasan Response Data : ", string(body))
	
	if len(searchResp.Result) == 0 {
		res.StatusScan = "rejected"
		res.StatusDesc = "Kemasan tidak ditemukan, Silahkan Coba Lagi."
		return res, nil
	}
	// Assign All Data to Response
	res.ProvinsiPenerima,_ = searchResp.Result[0].ProvinsiPenerima.( string )
	res.KodeAgen,_ = searchResp.Result[0].KodeAgen.( string )
	res.TglSPPB, _ = searchResp.Result[0].TglSPPB.(string)
	res.NoSPPB, _ = searchResp.Result[0].NoSPPB.(string)
	res.NoKemasan = searchResp.Result[0].Name
	res.WaktuGateIn = searchResp.Result[0].WaktuGateIn
	res.HasilPeriksa, _ = searchResp.Result[0].HasilPeriksa.(string)

	kemasan := searchResp.Result[0]

	// 2. Validate package status
	noSPPB, ok := kemasan.NoSPPB.(string)
	if !ok || noSPPB == "" {
		res.StatusScan = "rejected"
		res.StatusDesc = "Kemasan Belum SPPB, Silahkan Kembali ke Gudang."
		return res, nil
	}

	// ✅ Check Kriteria Muat using raw value, compare as ints
	kBlocked, kReason := checkKriteria(kemasan.KriteriaMuatID, filterID)
	if kBlocked {
		res.StatusScan = "abu"
		res.StatusDesc = kReason
		return res, nil
	}

	if kemasan.WaktuGateOut != nil && kemasan.WaktuGateOut != "" && kemasan.WaktuGateOut != false {
		res.StatusScan = "approved"
		res.StatusDesc = "Kemasan sudah di Scan, Silahkan Lanjutkan Gate out."
		res.WaktuGateOut, _ = kemasan.WaktuGateOut.(string)
		return res, nil
	}

	hasilPer, ok := kemasan.HasilPeriksa.(string)
	if ok && hasilPer != "" {
		// safer check to avoid slicing panic
		if hasilPer == "p2p" || hasilPer == "p2w" || hasilPer == "hold" {
			res.StatusScan = "rejected"
			res.StatusDesc = "Kemasan " + hasilPer + ", Harap Kembali ke Gudang."
			return res, nil
		} else if (len(hasilPer) >= 2 && hasilPer[:2] == "p2" && (hasilPer == "p2ph" || hasilPer != "p2wh")) || hasilPer == "clear" {
			res.StatusScan = "kuning"
			res.StatusDesc = "Kemasan Atensi " + hasilPer + ", Harap Cek Ulang Kemasan Sebelum Keluar"
			atensi = true
			// return res, nil
		}
	} 

	// Check SPBL
	check306Payload := map[string]interface{}{
		"params": map[string]interface{}{
			"model":  "dps.cn.pibk",
			"method": "check_306_api",
			"args":   []interface{}{kemasan.CnID[1]},
			"kwargs": map[string]interface{}{},
		},
	}

	body, _, err = c.workerRequest(callURL, check306Payload, sessionCookie)
	if err != nil {
		res.StatusScan = "rejected"
		res.StatusDesc = "Terjadi Kesalahan saat cek 306, Silahkan Coba Lagi."
		return res, fmt.Errorf("failed to call check_306_api: %w", err)
	}

	var check306Resp map[string]interface{}
	if err := json.Unmarshal(body, &check306Resp); err != nil {
		res.StatusScan = "rejected"
		res.StatusDesc = "Terjadi Kesalahan saat parsing hasil cek 306."
		return res, fmt.Errorf("failed to unmarshal check_306_api response: %w", err)
	}
	log.Printf("check_306_api result: %v", check306Resp["result"])

	res306, _ := check306Resp["result"].(string)
	if res306 == "1" {
		res.StatusScan = "kuning"
		res.StatusDesc = "Kemasan Atensi SPBL 306, Harap Cek Ulang Kemasan Sebelum Keluar"
		atensi = true
	}

	// 3. If valid, create a muat.ids record
	createMuatIdsPayload := map[string]interface{}{
		"params": map[string]interface{}{
			"model":  "dps.muat.ids",
			"method": "create",
			"args": []map[string]interface{}{
				{"muat_id": muatID, "kemasan_id": kemasan.ID},
			},
			"kwargs": map[string]interface{}{},
		},
	}

	_, _, err = c.workerRequest(callURL, createMuatIdsPayload, sessionCookie)
	if err != nil {
		res.StatusScan = "rejected"
		res.StatusDesc = "Terjadi Kesalahan, Silahkan Coba Lagi."
		return res, fmt.Errorf("failed to create muat.ids: %w", err)
	}

	// 4. Success
	if !atensi {
		res.StatusScan = "approved"
		res.StatusDesc = "Disetujui."
	}
	res.NoKemasan = kemasan.Name
	res.NoSPPB = noSPPB
	res.WaktuGateIn = kemasan.WaktuGateIn
	res.WaktuGateOut = time.Now().Format("2006-01-02 15:04:05")

	return res, nil
}

// CreateMuat creates a new dps.muat record.
func (c *Client) CreateMuat(nopol, driver, tujuan string) error {
	sessionCookie, err := c.authenticate()
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"params": map[string]interface{}{
			"model":  "dps.muat",
			"method": "create",
			"args": []map[string]string{
				{"nopol": nopol, "driver": driver, "tujuan": tujuan},
			},
			"kwargs": map[string]interface{}{},
		},
	}

	callURL := fmt.Sprintf("%s/web/dataset/call_kw", c.baseURL)
	_, _, err = c.workerRequest(callURL, payload, sessionCookie)
	return err
}

// GetMuatList retrieves a list of dps.muat records.
func (c *Client) GetMuatList() (interface{}, error) {
	sessionCookie, err := c.authenticate()
	if err != nil {
		return nil, err
	}

	payload := map[string]interface{}{
		"params": map[string]interface{}{
			"model":  "dps.muat",
			"method": "search_read",
			"args":   []interface{}{},
			"kwargs": map[string]interface{}{
				"fields": []string{"id", "nopol", "driver", "tujuan"},
				"order":  "id desc",
			},
		},
	}

	callURL := fmt.Sprintf("%s/web/dataset/call_kw", c.baseURL)
	body, _, err := c.workerRequest(callURL, payload, sessionCookie)
	if err != nil {
		return nil, err
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	return response["result"], nil
}

// GetKriteriaList retrieves a list of dps.reference records.
func (c *Client) GetKriteriaList() (interface{}, error) {
	sessionCookie, err := c.authenticate()
	if err != nil {
		return nil, err
	}

	payload := map[string]interface{}{
		"params": map[string]interface{}{
			"model":  "dps.reference",
			"method": "search_read",
			"args":   []interface{}{},
			"kwargs": map[string]interface{}{
				"domain": [][]interface{}{{"kode_master", "=", 21}},
				"fields": []string{"id", "name", "uraian"},
				"order":  "create_date desc",
			},
		},
	}

	callURL := fmt.Sprintf("%s/web/dataset/call_kw", c.baseURL)
	body, _, err := c.workerRequest(callURL, payload, sessionCookie)
	if err != nil {
		return nil, err
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	return response["result"], nil
}

// FetchDataAndCount is a generic method to fetch data from Odoo.
func (c *Client) FetchDataAndCount(req models.DataRequest) (*models.DataResponse, error) {
	sessionCookie, err := c.authenticate()
	if err != nil {
		return nil, err
	}
	callURL := fmt.Sprintf("%s/web/dataset/call_kw", c.baseURL)

	// 1. Get total count
	countPayload := map[string]interface{}{
		"params": map[string]interface{}{
			"model":  req.Model,
			"method": "search_count",
			"args":   []interface{}{req.Domain},
			"kwargs": map[string]interface{}{},
		},
	}
	countBody, _, err := c.workerRequest(callURL, countPayload, sessionCookie)
	if err != nil {
		return nil, err
	}
	var countResp struct{ Result int }
	if err := json.Unmarshal(countBody, &countResp); err != nil {
		return nil, err
	}

	// 2. Get records
	readPayload := map[string]interface{}{
		"params": map[string]interface{}{
			"model":  req.Model,
			"method": "search_read",
			"args":   []interface{}{req.Domain},
			"kwargs": map[string]interface{}{
				"fields": req.Fields,
				"limit":  req.Limit,
				"order":  req.Sort,
			},
		},
	}
	readBody, _, err := c.workerRequest(callURL, readPayload, sessionCookie)
	if err != nil {
		return nil, err
	}
	var readResp struct{ Result []interface{} }
	if err := json.Unmarshal(readBody, &readResp); err != nil {
		return nil, err
	}

	return &models.DataResponse{
		Records:    readResp.Result,
		TotalCount: countResp.Result,
	}, nil
}

// ✅ updated: accept raw kriteria value instead of a typed struct
func checkKriteria(kriteriaMuat any, filterID int) (blocked bool, reason string) {
	// Extract Many2one/scalar to int
	idKriteria, hasKriteria := m2oID(kriteriaMuat)

	log.Printf("Kriteria Muat raw: %#v, parsed ID: %d (has=%v), Filter ID: %d",
		kriteriaMuat, idKriteria, hasKriteria, filterID)

	if filterID == 0 {
		// no filter selected; nothing to enforce
		return false, ""
	}

	// If a filter is selected, kemasan must have a valid kriteria and match it
	if !hasKriteria {
		return true, "Kemasan tidak memiliki kriteria muat yang valid, Silahkan Coba Lagi."
	}
	if idKriteria != filterID {
		return true, "Kemasan tidak sesuai dengan kriteria muat yang dipilih, Silahkan Coba Lagi."
	}
	return false, ""
}
