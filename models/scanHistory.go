package models

const table = `
	CREATE TABLE IF NOT EXISTS scan_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT,
		muat_id TEXT,
		barcode TEXT,
		ip_address TEXT,
		location TEXT,
		scanned_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

type ScanHistory struct {
	ID        int    `json:"id"`
	SessionID string `json:"session_id"`
	MuatID    string `json:"muat_id"`
	Barcode   string `json:"barcode"`
	IPAddress string `json:"ip_address"`
	Location  string `json:"location"`
	ScannedAt string `json:"scanned_at"`
}

type ScanHistoryFilter struct {
	MuatID    string `json:"muat_id"`
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"`
}