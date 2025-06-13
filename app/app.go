package app

import "dps-scanner-gateout/services"

func SetupApp(
// repoRunNumb runningNumberRepository.Repository,
) services.UsecaseService {

	// Repository
	// runNumberRepo := runningNumberRepository.NewRunningNumberRepository(repoRunNumb)

	// Services
	usecaseSvc := services.NewUsecaseService(
	// runNumberRepo,
	)

	return usecaseSvc
}
