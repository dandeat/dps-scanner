package main

import (
	"dps-scanner-gateout/constants"
	"dps-scanner-gateout/utils"
	"encoding/json"
	"errors"
	"log"
	"net/http"
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
	// {
	//     "params": {
	//       "model": "dps.kemasan",
	//       "method": "search_read",
	//       "args": [],
	//       "kwargs": {
	//         "context": { "bin_size": true },
	//         "domain": [
	//           ["name", "=", "AL44021"]
	//         ],
	//         // "fields": ["id", "cn_id", "name", "no_sppb", "waktu_gatein", "hasil_periksa", "provinsi_penerima","kriteria_muat_id","kode_agen","note"],
	//         "fields": ["id", "cn_id", "name", "no_sppb", "waktu_gatein", "hasil_periksa", "provinsi_penerima", "kode_agen", "note"],
	//         "limit": 1
	//       }
	//     }
	// }
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
	// Result []struct {
	// 	// "id", "pjt", "name", "no_master", "kode_shipment", "no_sppb", "tgl_sppb", "waktu_gatein", "waktu_gateout", "status_akhir"
	// 	Pjt              string `json:"pjt"`
	// 	Name             string `json:"name"`
	// 	NomorMaster      string `json:"no_master"`
	// 	KodeShipment     any    `json:"kode_shipment"`
	// 	ProvinsiPenerima any    `json:"provinsi_penerima"`
	// 	ID               int    `json:"id"`
	// 	NoSPPB           any    `json:"no_sppb"`
	// 	CNID             []any  `json:"cn_id"`
	// 	Note             any    `json:"note"`
	// 	WaktuGateIn      string `json:"waktu_gatein"`
	// 	WaktuGateOut     any    `json:"waktu_gateout"`
	// } `json:"result"`
	ID      any    `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
}

// Fake webhook logic
func CallScanMuat(muatId, barcode string) (res DataMuatScan, err error) {

	// Simulate call API
	// response := DataMuatScan{
	// 	StatusScan:       "approved",
	// 	StatusDesc:       "Kemasan Approved, Silahkan Lanjutkan ke Gate Out",
	// 	Barcode:          barcode,
	// 	NoKemasan:        "AL44021",
	// 	NoSPPB:           "001234",
	// 	TglSPPB:          "2023-10-01",
	// 	HasilPeriksa:     "Hijau",
	// 	WaktuGateIn:      "2023-10-01T10:00:00Z",
	// 	WaktuGateOut:     "2023-10-01T12:00:00Z",
	// 	ProvinsiPenerima: "Jawa Barat",
	// 	KodeAgen:         "AG",
	// }

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

	// requestScan := RequestKW{
	// 	Params: RequestKWParams{
	// 		Model:  "dps.kemasan.tps",
	// 		Method: "search_read",
	// 		Args:   []any{},
	// 		Kwargs: Kwargs{
	// 			Context: map[string]bool{"bin_size": true},
	// 			Domain:  [][]any{{"name", "=", barcode}},
	// 			Fields:  []string{"id", "pjt", "name", "no_master", "kode_shipment", "no_sppb", "tgl_sppb", "waktu_gatein", "waktu_gateout", "status_akhir"},
	// 			Limit:   1,
	// 		},
	// 	},
	// }

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

	// Parse response
	/*
	   {
	       "result": [
	           {
	               "kode_agen": false,
	               "name": "AL44021",
	               "hasil_periksa": "hijau",
	               "provinsi_penerima": false,
	               "id": 5,
	               "no_sppb": "060820",
	               "cn_id": [
	                   79885,
	                   "AL44021"
	               ],
	               "note": false,
	               "waktu_gatein": "2025-03-08 01:55:28"
	           }
	       ],
	       "id": null,
	       "jsonrpc": "2.0"
	   }
	*/
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

		_, ok = response.Result[0].HasilPeriksa.(string)
		if ok {
			res.StatusScan = "approved"
			res.HasilPeriksa = response.Result[0].HasilPeriksa.(string)

			// if first 2 character is P2
			if strings.HasPrefix(res.HasilPeriksa, "P2") {
				// res.StatusScan = "rejected"
				res.StatusDesc = "Hasil Periksa Belum Hijau, Silahkan Kembali ke Gudang"

				// 	return res, nil
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
				res.StatusDesc = "Kemasan Sudah Termuat"
			}
			if response["result"] == nil {
				log.Println("Failed to add muat:", response)
				res.StatusScan = "rejected"
				res.StatusDesc = "Kemasan Sudah Termuat"
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

func main() {
	r := gin.Default()

	// WebSocket + API
	r.GET("/ws/:session_id", wsHandler)
	r.POST("/scan", scanHandler)
	r.POST("/tambah-muat", tambahMuatHandler)
	r.GET("/muat-list", muatListHandler)

	// ✅ Serve frontend at /ui
	r.Static("/ui", "./static")

	log.Println("Server running at http://localhost:8080")

	// add Logger with request body
	// r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
	// 	// param.Request.Body is a io.ReadCloser, you can read it here
	// 	// but you need to copy it to a buffer if you want to log it
	// 	// param.Request.Body.Close() is not needed here, gin will do it for you
	// 	// param.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	// 	// return the log string
	// 	return fmt.Sprintf("%s %s %s %s %d %s %s\n%s",
	// 		param.TimeStamp.Format(time.RFC3339),
	// 		param.ClientIP,
	// 		param.Method,
	// 		param.Path,
	// 		param.StatusCode,
	// 		param.Latency,
	// 		param.ErrorMessage,
	// 		param.Request.UserAgent(),
	// 	)
	// }))
	r.Run(":8080")
}
