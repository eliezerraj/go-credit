package service

import (
	"context"
	"errors"
	"github.com/rs/zerolog/log"
	"encoding/json"
	
	"github.com/sony/gobreaker"
	"github.com/go-credit/internal/core"
	"github.com/go-credit/internal/erro"
	"github.com/go-credit/internal/lib"
	"github.com/go-credit/internal/adapter/restapi"
	"github.com/go-credit/internal/repository/storage"
)

var childLogger = log.With().Str("service", "service").Logger()
var restApiCallData core.RestApiCallData

type WorkerService struct {
	workerRepo		 		*storage.WorkerRepository
	appServer				*core.AppServer
	restApiService			*restapi.RestApiService
	circuitBreaker			*gobreaker.CircuitBreaker
}

func NewWorkerService(workerRepo	*storage.WorkerRepository,
						appServer		*core.AppServer,
						restApiService	*restapi.RestApiService,
						circuitBreaker	*gobreaker.CircuitBreaker) *WorkerService{
	childLogger.Debug().Msg("NewWorkerService")

	return &WorkerService{
		workerRepo: 		workerRepo,
		appServer:			appServer,
		restApiService:		restApiService,
		circuitBreaker: 	circuitBreaker,
	}
}

func (s WorkerService) Add(ctx context.Context, credit *core.AccountStatement) (*core.AccountStatement, error){
	childLogger.Debug().Msg("Add")
	childLogger.Debug().Interface("credit:",credit).Msg("")

	span := lib.Span(ctx, "service.Add")	
	tx, conn, err := s.workerRepo.StartTx(ctx)
	if err != nil {
		return nil, err
	}
	
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
		s.workerRepo.ReleaseTx(conn)
		span.End()
	}()

	// BEGIN ------- Mock Circuit Breaker
	_, err = s.circuitBreaker.Execute(func() (interface{}, error) {
		if credit.Type == "CREDITX" {
			return nil, erro.ErrTransInvalid
		}
		return nil , nil
	})

	if (err != nil) {
		spanCB := lib.Span(ctx, "service.CIRCUIT-BREAKER")	

		childLogger.Debug().Msg("--------------------------------------------------")
		childLogger.Error().Err(err).Msg(" ****** Circuit Breaker OPEN !!! ******")
		childLogger.Debug().Msg("--------------------------------------------------")

		transfer := core.Transfer{}
		transfer.Currency = credit.Currency
		transfer.Amount = credit.Amount
		transfer.AccountIDTo = credit.AccountID

		restApiCallData.Method = "POST"
		restApiCallData.Url = s.appServer.RestEndpoint.ServiceUrlDomainCB + "/creditFundSchedule"
		restApiCallData.X_Api_Id = &s.appServer.RestEndpoint.XApigwIdCB

		_, err := s.restApiService.CallApiRest(ctx, restApiCallData, transfer)
		if err != nil {
			return nil, err
		}
	
		credit.Obs =  "transaction send via circuit breaker !!!"
		
		spanCB.End()
		return credit, nil
	}
	// END --------- Mock Circuit Breaker

	if credit.Type != "CREDIT" {
		err = erro.ErrTransInvalid
		return nil, err
	}

	if credit.Amount < 0 {
		err = erro.ErrInvalidAmount
		return nil, err
	}
	
	restApiCallData.Method = "GET"
	restApiCallData.Url = s.appServer.RestEndpoint.ServiceUrlDomain + "/get/" + credit.AccountID
	restApiCallData.X_Api_Id = &s.appServer.RestEndpoint.XApigwId

	rest_interface_data, err := s.restApiService.CallApiRest(ctx, restApiCallData, nil)
	if err != nil {
		return nil, err
	}
	jsonString, err  := json.Marshal(rest_interface_data)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Marshal")
		return nil, errors.New(err.Error())
    }
	var account_parsed core.Account
	json.Unmarshal(jsonString, &account_parsed)

	credit.FkAccountID = account_parsed.ID
	res, err := s.workerRepo.Add(ctx, tx, credit)
	if err != nil {
		return nil, err
	}

	childLogger.Debug().Interface("credit:",credit).Msg("")

	restApiCallData.Method = "POST"
	restApiCallData.Url = s.appServer.RestEndpoint.ServiceUrlDomain + "/add/fund"
	restApiCallData.X_Api_Id = &s.appServer.RestEndpoint.XApigwId

	_, err = s.restApiService.CallApiRest(ctx, restApiCallData, credit)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s WorkerService) List(ctx context.Context, credit *core.AccountStatement) (*[]core.AccountStatement, error){
	childLogger.Debug().Msg("List")
	childLogger.Debug().Interface("credit:",credit).Msg("")
	
	span := lib.Span(ctx, "service.List")	
    defer span.End()

	restApiCallData.Method = "GET"
	restApiCallData.Url = s.appServer.RestEndpoint.ServiceUrlDomain + "/get/" + credit.AccountID
	restApiCallData.X_Api_Id = &s.appServer.RestEndpoint.XApigwId

	rest_interface_data, err := s.restApiService.CallApiRest(ctx, restApiCallData, nil)
	if err != nil {
		return nil, err
	}
	jsonString, err  := json.Marshal(rest_interface_data)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Marshal")
		return nil, errors.New(err.Error())
    }
	var account_parsed core.Account
	json.Unmarshal(jsonString, &account_parsed)

	credit.FkAccountID = account_parsed.ID
	credit.Type = "CREDIT"

	res, err := s.workerRepo.List(ctx, credit)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s WorkerService) ListPerDate(ctx context.Context, credit *core.AccountStatement) (*[]core.AccountStatement, error){
	childLogger.Debug().Msg("ListPerDate")
	childLogger.Debug().Interface("credit:",credit).Msg("")
	
	span := lib.Span(ctx, "service.List")	
    defer span.End()

	restApiCallData.Method = "GET"
	restApiCallData.Url = s.appServer.RestEndpoint.ServiceUrlDomain + "/get/" + credit.AccountID
	restApiCallData.X_Api_Id = &s.appServer.RestEndpoint.XApigwId

	rest_interface_data, err := s.restApiService.CallApiRest(ctx, restApiCallData, nil)
	if err != nil {
		return nil, err
	}
	jsonString, err  := json.Marshal(rest_interface_data)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Marshal")
		return nil, errors.New(err.Error())
    }
	var account_parsed core.Account
	json.Unmarshal(jsonString, &account_parsed)

	credit.FkAccountID = account_parsed.ID
	credit.Type = "CREDIT"

	res, err := s.workerRepo.ListPerDate(ctx, credit)
	if err != nil {
		return nil, err
	}

	return res, nil
}