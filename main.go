package main

import (
	"bytes"
	"crypto/tls"
	"dps-scanner-gateout/constants"
	"dps-scanner-gateout/utils"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	mkpmobileutils "github.com/dandeat/mkpmobile-utils/src/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	err error
)

// === Models ===
type ScanPayload struct {
	MuatID    string `json:"muat_id"`
	SessionID string `json:"session_id"`
	Barcode   string `json:"barcode"`
}

type WebhookResponse struct {
	Status bool         `json:"status"`
	Data   DataMuatScan `json:"data"`
}

type DataMuatScan struct {
	StatusScan       string `json:"status_scan"` // "approved", "rejected"
	StatusDesc       string `json:"status_desc"`
	Barcode          string `json:"barcode"`
	NoKemasan        string `json:"no_kemasan"`
	NoSPPB           string `json:"no_sppb"`
	TglSPPB          string `json:"tgl_sppb"`
	HasilPeriksa     string `json:"hasil_periksa"`
	WaktuGateIn      string `json:"waktu_gate_in"`
	WaktuGateOut     string `json:"waktu_gate_out"`
	ProvinsiPenerima string `json:"provinsi_penerima"`
	KodeAgen         string `json:"kode_agen"`
}

// === WebSocket Management ===
type SessionHub struct {
	Clients    map[*websocket.Conn]bool
	Broadcast  chan []byte
	Register   chan *websocket.Conn
	Unregister chan *websocket.Conn
}

var (
	hubs     = make(map[string]*SessionHub)
	hubsMu   sync.Mutex
	upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
)

// Create or get existing session hub
func getOrCreateHub(kodeMuat string) *SessionHub {
	hubsMu.Lock()
	defer hubsMu.Unlock()

	if hub, exists := hubs[kodeMuat]; exists {
		return hub
	}

	hub := &SessionHub{
		Clients:    make(map[*websocket.Conn]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *websocket.Conn),
		Unregister: make(chan *websocket.Conn),
	}
	hubs[kodeMuat] = hub

	go hub.run()
	return hub
}

func (hub *SessionHub) run() {
	for {
		select {
		case conn := <-hub.Register:
			hub.Clients[conn] = true
		case conn := <-hub.Unregister:
			if _, ok := hub.Clients[conn]; ok {
				delete(hub.Clients, conn)
				conn.Close()
			}
		case message := <-hub.Broadcast:
			for conn := range hub.Clients {
				conn.WriteMessage(websocket.TextMessage, message)
			}
		}
	}
}

// === WebSocket Endpoint ===
func wsHandler(c *gin.Context) {
	sessionID := c.Param("session_id")
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket Upgrade error:", err)
		return
	}

	hub := getOrCreateHub(sessionID)
	hub.Register <- conn

	defer func() {
		hub.Unregister <- conn
	}()

	// Keep alive
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

// === REST /scan Endpoint ===
func scanHandler(c *gin.Context) {
	var payload ScanPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// Simulate webhook response (replace with real HTTP call)
	result := WebhookResponse{
		Status: false,
		Data:   DataMuatScan{},
	}

	result.Data, err = CallScanMuat(payload.MuatID, payload.Barcode)
	if err != nil {
		result.Data.StatusScan = "error"
		result.Data.StatusDesc = "Gagal Melakukan Scan, Harap Coba Kembali"
	}

	data, _ := json.Marshal(result)
	hub := getOrCreateHub(payload.SessionID)
	hub.Broadcast <- data

	c.JSON(http.StatusOK, result)
}

// Request KW
type RequestKW struct {
	Params RequestKWParams `json:"params"`
}

type RequestKWParams struct {
	Model  string `json:"model"`
	Method string `json:"method"`
	Args   []any  `json:"args"`
	Kwargs Kwargs `json:"kwargs"`
}

type Kwargs struct {
	Context map[string]bool `json:"context"`
	Domain  [][]any         `json:"domain"`
	Fields  []string        `json:"fields"`
	Limit   int             `json:"limit"`
}

type ResponseScanSuccess struct {
	Result []struct {
		KodeAgen         any    `json:"kode_agen"`
		Name             string `json:"name"`
		HasilPeriksa     any    `json:"hasil_periksa"`
		ProvinsiPenerima any    `json:"provinsi_penerima"`
		ID               int    `json:"id"`
		NoSPPB           any    `json:"no_sppb"`
		CNID             []any  `json:"cn_id"`
		Note             any    `json:"note"`
		WaktuGateIn      string `json:"waktu_gatein"`
		WaktuGateOut     any    `json:"waktu_gateout"`
	} `json:"result"`
	ID      any    `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
}

// Fake webhook logic
func CallScanMuat(muatId, barcode string) (res DataMuatScan, err error) {

	res = DataMuatScan{
		StatusScan: "error",
		StatusDesc: "Terjadi Kesalahan",
	}

	requestScan := RequestKW{
		Params: RequestKWParams{
			Model:  "dps.kemasan",
			Method: "search_read",
			Args:   []any{},
			Kwargs: Kwargs{
				Context: map[string]bool{"bin_size": true},
				Domain:  [][]any{{"name", "=", barcode}},
				Fields:  []string{"id", "cn_id", "name", "no_sppb", "waktu_gatein", "hasil_periksa", "provinsi_penerima", "kode_agen", "note", "waktu_gateout"},
				Limit:   1,
			},
		},
	}

	sessId, err := getTokenCron()
	if err != nil {
		log.Println("Error getting session ID:", err)
		return res, err
	}

	// Call API POST
	result, _, code, err := utils.WorkerRequestPOST(
		constants.REQ_JSON,
		"https://transmarine.oneerp.app/web/dataset/call_kw",
		requestScan,
		mkpmobileutils.ReqHeader{},
		time.Second*60,
		sessId,
	)
	if err != nil {
		log.Println("Error calling API:", err)
		return res, err
	} else if code != http.StatusOK {
		log.Println("Error response code:", code)
		return res, err
	}

	var response ResponseScanSuccess
	if err := json.Unmarshal(result, &response); err != nil {
		log.Println("Error unmarshalling response:", err)
		return res, err
	}
	if len(response.Result) > 0 {
		res = DataMuatScan{
			StatusScan:       "approved",
			StatusDesc:       "Silahkan Lanjutkan ke Gate Out",
			Barcode:          barcode,
			NoKemasan:        response.Result[0].Name,
			WaktuGateIn:      response.Result[0].WaktuGateIn,
			WaktuGateOut:     time.Now().Format("2006-01-02 15:04:05"),
			ProvinsiPenerima: "Jawa Timur",
			KodeAgen:         "Agen1",
		}

		_, ok := response.Result[0].NoSPPB.(string)
		if ok {
			res.StatusScan = "approved"
			res.NoSPPB = response.Result[0].NoSPPB.(string)
			if res.NoSPPB == "" {
				res.StatusScan = "rejected"
				res.StatusDesc = "Kemasan Belum SPPB, Silahkan Kembali ke Gudang"

				return res, nil
			}
		} else {
			res.StatusScan = "rejected"
			res.StatusDesc = "Kemasan Belum SPPB, Silahkan Kembali ke Gudang"
			return res, nil
		}

		isGateout := response.Result[0].WaktuGateOut != nil && response.Result[0].WaktuGateOut != ""
		if isGateout {
			res.StatusScan = "approved"
			res.StatusDesc = "Kemasan Sudah Scan Gate Out, Silahkan Lanjutkan ke Proses Berikutnya"
			res.WaktuGateOut = response.Result[0].WaktuGateOut.(string)
			return res, nil
		}

		_, ok = response.Result[0].HasilPeriksa.(string)
		if ok {
			res.StatusScan = "approved"
			// res.HasilPeriksa =
			if res.HasilPeriksa, ok = response.Result[0].HasilPeriksa.(string); !ok {
				res.StatusScan = "rejected"
				res.StatusDesc = "terjadi kesalahan, hasil periksa tidak valid"
				// return res, nil
			} else if strings.HasPrefix(res.HasilPeriksa, "P2") {
				res.StatusScan = "kuning"
				res.StatusDesc = "Hasil Periksa " + res.HasilPeriksa + ", Harap Konfirmasi ke Petugas"
				// return res, nil
			}

			// {
			// 	model: 'dps.muat.ids',
			// 	method: 'create',
			// 	args: [{ muat_id: id_muat, kemasan_id: vKode }],
			// 	kwargs: {},
			// }

			reqAddMuatIds := map[string]any{
				"params": map[string]any{
					"model":  "dps.muat.ids",
					"method": "create",
					"args": []map[string]any{
						{
							"muat_id":    muatId,
							"kemasan_id": response.Result[0].ID,
						},
					},
					"kwargs": map[string]any{},
				},
			}

			// Call API POST
			resp, _, code, err := utils.WorkerRequestPOST(
				constants.REQ_JSON,
				"https://transmarine.oneerp.app/web/dataset/call_kw",
				reqAddMuatIds,
				mkpmobileutils.ReqHeader{},
				time.Second*60,
				sessId,
			)
			if err != nil {
				log.Println("Error calling API:", err)
				res.StatusScan = "rejected"
				res.StatusDesc = "Terjadi Kesalahan, Silahkan Coba Lagi"
			}
			if code != http.StatusOK {
				log.Println("Error response code:", code)
				res.StatusScan = "rejected"
				res.StatusDesc = "Terjadi Kesalahan, Silahkan Coba Lagi"
			}

			var response map[string]any
			if err := json.Unmarshal(resp, &response); err != nil {
				log.Println("Error unmarshalling response:", err)
				res.StatusScan = "rejected"
				res.StatusDesc = "Terjadi Kesalahan, Silahkan Coba Lagi"
			}
			if response["result"] == nil {
				log.Println("Failed to add muat:", response)
				res.StatusScan = "rejected"
				res.StatusDesc = "Terjadi Kesalahan, Silahkan Coba Lagi"
			}
			log.Println("Muat added successfully:", response)

			return res, nil
		} else {
			res.StatusScan = "rejected"
			res.StatusDesc = "Kemasan Belum di Periksa, Silahkan Kembali ke Gudang"
		}
	} else {
		res = DataMuatScan{
			StatusScan: "rejected",
			StatusDesc: "Kemasan Tidak Ditemukan, Silahkan Coba Lagi",
		}
	}

	return res, nil
}

type TambahMuatPayload struct {
	SessionID string `json:"session_id"`
	NoPol     string `json:"no_pol"`
	Driver    string `json:"driver"`
	Tujuan    string `json:"tujuan"`
}

func tambahMuatHandler(c *gin.Context) {
	var payload TambahMuatPayload
	// Validate payload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	sessId, err := getTokenCron()
	if err != nil {
		log.Println("Error getting session ID:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get session ID"})
		return
	}

	req := map[string]any{
		"params": map[string]any{
			"model":  "dps.muat",
			"method": "create",
			"args": []map[string]string{
				{
					"nopol":  payload.NoPol,
					"driver": payload.Driver,
					"tujuan": payload.Tujuan,
				},
			},
			"kwargs": map[string]any{},
		},
	}

	// Call API POST
	resp, _, code, err := utils.WorkerRequestPOST(
		constants.REQ_JSON,
		"https://transmarine.oneerp.app/web/dataset/call_kw",
		req,
		mkpmobileutils.ReqHeader{},
		time.Second*60,
		sessId,
	)
	if err != nil {
		log.Println("Error calling API:", err)
		return
	} else if code != http.StatusOK {
		log.Println("Error response code:", code)
		return
	}

	var response ResponseLogin
	if err := json.Unmarshal(resp, &response); err != nil {
		log.Println("Error unmarshalling response:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse response"})
		return
	}
	log.Println("Muat added successfully:", response)

	c.JSON(http.StatusOK, true)
}

func muatListHandler(c *gin.Context) {

	sessId, err := getTokenCron()
	if err != nil {
		log.Println("Error getting session ID:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get session ID"})
		return
	}

	req := map[string]any{
		"params": map[string]any{
			"model":  "dps.muat",
			"method": "search_read",
			"args":   []any{},
			"kwargs": map[string]any{
				"context": map[string]bool{"bin_size": true},
				"fields":  []string{"id", "nopol", "driver", "tujuan"},
				"order":   "id desc",
			},
		},
	}

	// Call API POST
	resp, _, code, err := utils.WorkerRequestPOST(
		constants.REQ_JSON,
		"https://transmarine.oneerp.app/web/dataset/call_kw",
		req,
		mkpmobileutils.ReqHeader{},
		time.Second*60,
		sessId,
	)
	if err != nil {
		log.Println("Error calling API:", err)
		return
	}
	if code != http.StatusOK {
		log.Println("Error response code:", code)
		return
	}
	var response map[string]any
	if err := json.Unmarshal(resp, &response); err != nil {
		log.Println("Error unmarshalling response:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse response"})
		return
	}
	if response["result"] == nil {
		log.Println("Failed to get muat list:", response)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get muat list"})
		return
	}
	log.Println("Muat list retrieved successfully:", response)
	c.JSON(http.StatusOK, response["result"])
}

type ResponseLogin struct {
	Result struct {
		SessionID any `json:"session_id"`
	} `json:"result"`
	ID      any    `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
}

func getTokenCron() (sessId *http.Cookie, err error) {
	var (
	// ok       bool
	// response ResponseLogin
	)

	req := map[string]any{
		"jsonrpc": "2.0",
		"method":  "call",
		"params": map[string]string{
			"db":       "transmarine_cn",
			"login":    "transmarine",
			"password": "transmarine",
		},
	}

	_, respRaw, code, err := utils.WorkerRequestPOST(
		constants.REQ_JSON,
		"https://transmarine.oneerp.app/web/session/authenticate",
		req,
		mkpmobileutils.ReqHeader{},
		time.Second*60,
		nil,
	)
	if err != nil {
		log.Println("Error calling API:", err)
		return
	}
	if code != http.StatusOK {
		log.Println("Error response code:", code)
		err = errors.New("failed to authenticate")
		return
	}

	// Get Cookies From Response
	cookies := respRaw.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "session_id" {
			sessId = cookie
			break
		}
	}
	if sessId == nil {
		log.Println("Session ID not found in cookies")
		err = errors.New("session_id not found in cookies")
		return
	}

	log.Println("Session ID:", sessId)
	return sessId, nil
}

// OdooConfig holds your Odoo connection details.
type OdooConfig struct {
	AuthURL  string
	CallURL  string
	DB       string
	Username string
	Password string
}

// DataRequest defines the structure for incoming API requests from the frontend.
type DataRequest struct {
	Model  string          `json:"model"`
	Fields []string        `json:"fields"`
	Domain [][]interface{} `json:"domain"` // NEW: Field for Odoo domain filter
	Limit  int             `json:"limit"`  // NEW: Field for limiting the number of records
}

// DataResponse is the structure sent back to the frontend, including the total count.
type DataResponse struct {
	Records    []interface{} `json:"records"`
	TotalCount int           `json:"total_count"`
}

// AuthResponse defines the structure of a successful authentication response.
type AuthResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  struct {
		SessionID string `json:"session_id"`
		UID       int    `json:"uid"`
	} `json:"result"`
}

// CountResponse defines the structure for an Odoo search_count result.
type CountResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  int    `json:"result"`
}

// ReadResponse defines the structure for an Odoo search_read result.
type ReadResponse struct {
	Jsonrpc string        `json:"jsonrpc"`
	Result  []interface{} `json:"result"`
}

// OdooErrorResponse defines the structure of an Odoo error response.
type OdooErrorResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Error   struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
		Data    struct {
			Debug string `json:"debug"`
			Name  string `json:"name"`
		} `json:"data"`
	} `json:"error"`
}

// workerRequest sends a POST request with a JSON body.
func workerRequest(url string, payload map[string]interface{}, sessionCookie *http.Cookie) ([]byte, []*http.Cookie, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if sessionCookie != nil {
		req.AddCookie(sessionCookie)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr, Timeout: 30 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return responseBody, nil, fmt.Errorf("API returned non-200 status: %d", resp.StatusCode)
	}

	return responseBody, resp.Cookies(), nil
}

// authenticate with Odoo and get the session cookie.
func (c *OdooConfig) authenticate() (*http.Cookie, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0", "method": "call", "params": map[string]interface{}{"db": c.DB, "login": c.Username, "password": c.Password, "context": map[string]interface{}{}},
	}
	body, cookies, err := workerRequest(c.AuthURL, payload, nil)
	if err != nil {
		return nil, fmt.Errorf("authentication worker failed: %w", err)
	}
	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		var errResp OdooErrorResponse
		if json.Unmarshal(body, &errResp) == nil {
			return nil, fmt.Errorf("Odoo auth error: %s", errResp.Error.Data.Debug)
		}
		return nil, fmt.Errorf("failed to unmarshal auth response: %w", err)
	}
	for _, cookie := range cookies {
		if cookie.Name == "session_id" {
			return cookie, nil
		}
	}
	return nil, fmt.Errorf("session_id cookie not found")
}

// fetchDataAndCount gets the total record count and a limited list of records.
func (c *OdooConfig) fetchDataAndCount(sessionCookie *http.Cookie, model string, fields []string, domain [][]interface{}, limit int) (*DataResponse, error) {
	// Use an empty domain if the provided one is nil
	if domain == nil {
		domain = [][]interface{}{}
	}

	// 1. Get total count with the filter
	countPayload := map[string]interface{}{
		"jsonrpc": "2.0", "method": "call", "params": map[string]interface{}{"model": model, "method": "search_count", "args": []interface{}{domain}, "kwargs": map[string]interface{}{}},
	}
	countBody, _, err := workerRequest(c.CallURL, countPayload, sessionCookie)
	if err != nil {
		return nil, fmt.Errorf("count worker failed: %w", err)
	}
	var countResp CountResponse
	if err := json.Unmarshal(countBody, &countResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal count response: %w", err)
	}

	// 2. Get limited data with the filter
	readPayload := map[string]interface{}{
		"jsonrpc": "2.0", "method": "call", "params": map[string]interface{}{"model": model, "method": "search_read", "args": []interface{}{domain}, "kwargs": map[string]interface{}{"fields": fields, "limit": limit}},
	}
	readBody, _, err := workerRequest(c.CallURL, readPayload, sessionCookie)
	if err != nil {
		return nil, fmt.Errorf("read worker failed: %w", err)
	}
	var readResp ReadResponse
	if err := json.Unmarshal(readBody, &readResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal read response: %w", err)
	}

	// 3. Combine into a single response
	finalResponse := &DataResponse{
		Records:    readResp.Result,
		TotalCount: countResp.Result,
	}
	return finalResponse, nil
}

func main() {
	r := gin.Default()

	odooConfig := &OdooConfig{
		AuthURL:  "https://transmarine.oneerp.app/web/session/authenticate",
		CallURL:  "https://transmarine.oneerp.app/web/dataset/call_kw",
		DB:       "transmarine_cn",
		Username: "transmarine",
		Password: "transmarine",
	}

	// r.Use(func(c *gin.Context) {
	// 	if c.Request.URL.Path == "/static/" || c.Request.URL.Path == "/static" {
	// 		fs.ServeHTTP(c.Writer, c.Request)
	// 		c.Abort()
	// 	}
	// })

	// WebSocket + API
	r.GET("/ws/:session_id", wsHandler)
	r.POST("/scan", scanHandler)
	r.POST("/tambah-muat", tambahMuatHandler)
	r.GET("/muat-list", muatListHandler)

	// ✅ Serve frontend at /ui
	r.Static("/ui", "./static")

	// r.Static("/ui/dashboard", "./static/dashboard.html")
	r.POST("/api/odoo", func(c *gin.Context) {
		var req DataRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		// Authenticate and get session cookie

		sessionCookie, err := odooConfig.authenticate()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "authentication failed", "details": err.Error()})
			return
		}

		// Fetch data and count
		response, err := odooConfig.fetchDataAndCount(sessionCookie, req.Model, req.Fields, req.Domain, req.Limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "data fetch failed", "details": err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	})

	r.GET("/api/load_layout", func(c *gin.Context) {
		data, err := os.ReadFile("./layout.json")
		if err != nil {
			if os.IsNotExist(err) {
				c.JSON(http.StatusOK, []map[string]interface{}{}) // Return empty layout if file does not exist
				return
			}
			log.Println("Error reading layout file:", err)
			// Return a 500 Internal Server Error if reading the file fails
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read layout file", "details": err.Error()})
			return
		}
		c.JSON(http.StatusOK, json.RawMessage(data))
	})

	r.POST("/api/save_layout", func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Println("Error reading request body:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read request body", "details": err.Error()})
			return
		}
		if err := os.WriteFile("./layout.json", body, 0644); err != nil {
			log.Println("Error saving layout file:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save layout file", "details": err.Error()})
			return
		}
		c.JSON(http.StatusOK, nil)
	})

	log.Println("Server running at http://localhost:8080")
	r.Run(":8080")
}
