package muatsessionservice

import (
	"dps-scanner-gateout/services"

	"github.com/gin-gonic/gin"
)

type muatSessionService struct {
	service services.UsecaseService
}

func NewMuatSessionService(service services.UsecaseService) muatSessionService {
	return muatSessionService{
		service: service,
	}
}

func (svc muatSessionService) MuatListService(ctx *gin.Context) (err error) {

	// var (
	// 	ok bool

	// 	sessionHub *models.SessionHub
	// )

	return
}
