package scanhistoryrepository

import (
	"dps-scanner-gateout/constants"
	"dps-scanner-gateout/models"
	"dps-scanner-gateout/repositories"
	"dps-scanner-gateout/utils"
)

type ScanHistoryRepository struct {
	RepoDB repositories.Repository
}

func NewScanHistoryRepository(repoDB repositories.Repository) ScanHistoryRepository {
	return ScanHistoryRepository{
		RepoDB: repoDB,
	}
}

const defineCol = `
	session_id, muat_id, barcode,
	ip_address, location, scanned_at
`

func (r *ScanHistoryRepository) GetListByIndex(filter models.ScanHistoryFilter) (result []models.ScanHistory, err error) {
	var (
		args []interface{}
	)

	query := `
		SELECT 
			id, ` + defineCol + `
		FROM scan_history 
		WHERE TRUE
	`

	if filter.MuatID != constants.EMPTY_VALUE {
		query += ` AND muat_id = ?`
		args = append(args, filter.MuatID)
	}

	if filter.Limit > 0 {
		query += ` LIMIT ?`
		args = append(args, filter.Limit)

		if filter.Offset > 0 {
			query += ` OFFSET ?`
			args = append(args, filter.Offset)
		}
	}

	if filter.SortBy != constants.EMPTY_VALUE {
		query += ` ORDER BY ` + filter.SortBy
		if filter.SortOrder != constants.EMPTY_VALUE {
			query += ` ` + filter.SortOrder
		}
	} else {
		query += ` ORDER BY scanned_at DESC`
	}

	query = utils.ReplaceSQL(query, "?")

	rows, err := r.RepoDB.DB.Query(query, args...)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var val models.ScanHistory

		err = rows.Scan(
			&val.ID,
			&val.SessionID, &val.MuatID, &val.Barcode,
			&val.IPAddress, &val.Location, &val.ScannedAt,
		)
		if err != nil {
			return
		}
		result = append(result, val)
	}

	return
}

func (r *ScanHistoryRepository) AddScanHistory(scanHistory models.ScanHistory) (ID int64, err error) {
	var args []interface{}

	query := `
		INSERT INTO scan_history (
			` + defineCol + `
		) VALUES (
			` + utils.QueryFill(defineCol) + `
		)
		RETURNING id`

	args = append(args, scanHistory.SessionID, scanHistory.MuatID, scanHistory.Barcode,
		scanHistory.IPAddress, scanHistory.Location, scanHistory.ScannedAt,
	)
	query = utils.ReplaceSQL(query, "?")
	stmt, err := r.RepoDB.DB.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()

	err = stmt.QueryRow(args...).Scan(&ID)
	return
}