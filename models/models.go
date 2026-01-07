package models

// ScanPayload is the structure for incoming scan requests.
type ScanPayload struct {
	MuatID    string `json:"muat_id"`
	SessionID string `json:"session_id"`
	Barcode   string `json:"barcode"`
	FilterID  int    `json:"kriteria_muat"`
}

// WebhookResponse is the top-level structure for broadcasting scan results.
type WebhookResponse struct {
	Status bool         `json:"status"`
	Data   DataMuatScan `json:"data"`
}

// DataMuatScan holds the detailed result of a barcode scan.
type DataMuatScan struct {
	StatusScan       string `json:"status_scan"` // "approved", "rejected", "error"
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

// TambahMuatPayload is the structure for creating a new "muat".
type TambahMuatPayload struct {
	NoPol  string `json:"no_pol"`
	Driver string `json:"driver"`
	Tujuan string `json:"tujuan"`
}

// DataRequest defines the structure for generic Odoo API requests from the frontend.
type DataRequest struct {
	Type   string          `json:"type"`
	Model  string          `json:"model"`
	Fields []string        `json:"fields"`
	Domain [][]interface{} `json:"domain"`
	Limit  int             `json:"limit"`
	Sort   string          `json:"sort"`
}

// DataResponse is the structure sent back to the frontend from the generic proxy.
type DataResponse struct {
	Records    []interface{} `json:"records"`
	TotalCount int           `json:"total_count"`
}
