package service

import(
	"github.com/go-credit/internal/core/model"
	"github.com/go-credit/internal/adapter/database"
	"github.com/rs/zerolog/log"
)

var childLogger = log.With().Str("component","go-credit").Str("package","internal.core.service").Logger()

type WorkerService struct {
	workerRepository *database.WorkerRepository
	apiService		[]model.ApiService
}

// About create a ner worker service
func NewWorkerService(	workerRepository *database.WorkerRepository,
						apiService		[]model.ApiService) *WorkerService{
	childLogger.Debug().Msg("NewWorkerService")

	return &WorkerService{
		workerRepository: workerRepository,
		apiService: apiService,
	}
}