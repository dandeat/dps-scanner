package handlers

import (
	"dps-scanner-gateout/config"
	"dps-scanner-gateout/odoo"
	"dps-scanner-gateout/websocket"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all the routes for the application.
func SetupRoutes(r *gin.Engine, cfg *config.Config) {
	// Create a new Odoo client
	odooClient := odoo.NewClient(cfg.OdooURL, cfg.OdooDB, cfg.OdooUsername, cfg.OdooPassword)

	// Create a new WebSocket manager
	wsManager := websocket.NewManager()
	go wsManager.Run()

	// Create the main handler which holds dependencies
	h := &APIHandler{
		OdooClient: odooClient,
		WSManager:  wsManager,
	}

	// WebSocket and Core Scanner API Routes
	r.GET("/ws/:session_id", h.WebSocketHandler)
	r.POST("/scan", h.ScanHandler)
	r.POST("/tambah-muat", h.TambahMuatHandler)
	r.GET("/muat-list", h.MuatListHandler)
	r.GET("/kriteria-list", h.KriteriaListHandler)

	// Generic Odoo API proxy and static file serving
	r.POST("/api/odoo", h.OdooProxyHandler)
	r.GET("/api/load_layout", h.LoadLayoutHandler)
	r.POST("/api/save_layout", h.SaveLayoutHandler)

	// Serve the frontend UI
	r.Static("/ui", "./static")
}
