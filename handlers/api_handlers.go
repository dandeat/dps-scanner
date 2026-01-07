package handlers

import (
	"dps-scanner-gateout/models"
	"dps-scanner-gateout/odoo"
	"dps-scanner-gateout/websocket"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// APIHandler holds dependencies for the API handlers.
type APIHandler struct {
	OdooClient *odoo.Client
	WSManager  *websocket.Manager
}

// WebSocketHandler handles WebSocket upgrade requests.
func (h *APIHandler) WebSocketHandler(c *gin.Context) {
	sessionID := c.Param("session_id")
	h.WSManager.HandleConnection(c.Writer, c.Request, sessionID)
}

// ScanHandler processes barcode scans and broadcasts results.
func (h *APIHandler) ScanHandler(c *gin.Context) {
	var payload models.ScanPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	scanResult, err := h.OdooClient.CallScanMuat(payload.MuatID, payload.Barcode, payload.FilterID)
	if err != nil {
		log.Printf("Error from CallScanMuat: %v", err)
		scanResult.StatusScan = "error"
		scanResult.StatusDesc = "Failed to process scan, please try again."
	}

	response := models.WebhookResponse{
		Status: err == nil,
		Data:   scanResult,
	}

	data, _ := json.Marshal(response)
	h.WSManager.Broadcast(payload.SessionID, data)

	c.JSON(http.StatusOK, response)
}

// TambahMuatHandler handles adding a new "muat".
func (h *APIHandler) TambahMuatHandler(c *gin.Context) {
	var payload models.TambahMuatPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	err := h.OdooClient.CreateMuat(payload.NoPol, payload.Driver, payload.Tujuan)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create muat", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// MuatListHandler retrieves the list of "muat".
func (h *APIHandler) MuatListHandler(c *gin.Context) {
	muatList, err := h.OdooClient.GetMuatList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get muat list", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, muatList)
}

// KriteriaListHandler retrieves the list of criteria.
func (h *APIHandler) KriteriaListHandler(c *gin.Context) {
	kriteriaList, err := h.OdooClient.GetKriteriaList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get kriteria list", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, kriteriaList)
}

// OdooProxyHandler acts as a generic proxy for Odoo API calls.
func (h *APIHandler) OdooProxyHandler(c *gin.Context) {
	var req models.DataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	response, err := h.OdooClient.FetchDataAndCount(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "data fetch failed", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// LoadLayoutHandler loads the UI layout from a JSON file.
func (h *APIHandler) LoadLayoutHandler(c *gin.Context) {
	data, err := os.ReadFile("layout.json")
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusOK, []map[string]interface{}{}) // Return empty if not found
			return
		}
		log.Println("Error reading layout file:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read layout file"})
		return
	}
	c.JSON(http.StatusOK, json.RawMessage(data))
}

// SaveLayoutHandler saves the UI layout to a JSON file.
func (h *APIHandler) SaveLayoutHandler(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Println("Error reading request body:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read request body"})
		return
	}
	if err := os.WriteFile("layout.json", body, 0644); err != nil {
		log.Println("Error saving layout file:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save layout file"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
