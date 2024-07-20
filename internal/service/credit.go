package service

import (
	"context"
	"errors"
	"github.com/rs/zerolog/log"
	
	"github.com/sony/gobreaker"
	"github.com/mitchellh/mapstructure"
	"github.com/go-credit/internal/core"
	"github.com/go-credit/internal/erro"
	"github.com/go-credit/internal/lib"
	"github.com/go-credit/internal/adapter/restapi"
	"github.com/go-credit/internal/repository/pg"
)

var childLogger = log.With().Str("service", "service").Logger()

type WorkerService struct {
	workerRepo		 		*pg.WorkerRepository
	appServer				*core.AppServer
	restApiService			*restapi.RestApiService
	circuitBreaker			*gobreaker.CircuitBreaker
}

func NewWorkerService(	workerRepo		*pg.WorkerRepository,
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

func (s WorkerService) SetSessionVariable(ctx context.Context, userCredential string) (bool, error){
	childLogger.Debug().Msg("SetSessionVariable")

	res, err := s.workerRepo.SetSessionVariable(ctx, userCredential)
	if err != nil {
		return false, err
	}

	return res, nil
}

func (s WorkerService) Add(ctx context.Context, credit core.AccountStatement) (*core.AccountStatement, error){
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

		path := s.appServer.RestEndpoint.ServiceUrlDomainCB + "/creditFundSchedule"
		_, err := s.restApiService.CallRestApi(ctx,	"POST",	path, &s.appServer.RestEndpoint.XApigwIdCB ,transfer)
	
		if err != nil {
			return nil, err
		}

		credit.Obs =  "Transação enviada via Circuit Breaker !!!!"
		
		spanCB.End()
		return &credit, nil
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

	path := s.appServer.RestEndpoint.ServiceUrlDomain + "/get/" + credit.AccountID
	rest_interface_data, err := s.restApiService.CallRestApi(ctx,"GET", path, &s.appServer.RestEndpoint.XApigwId, nil)

	if err != nil {
		return nil, err
	}
	var account_parsed core.Account
	err = mapstructure.Decode(rest_interface_data, &account_parsed)
    if err != nil {
		childLogger.Error().Err(err).Msg("error parse interface")
		return nil, errors.New(err.Error())
    }

	credit.FkAccountID = account_parsed.ID
	res, err := s.workerRepo.Add(ctx, tx, credit)
	if err != nil {
		return nil, err
	}

	childLogger.Debug().Interface("credit:",credit).Msg("")

	path = s.appServer.RestEndpoint.ServiceUrlDomain + "/add/fund"
	_, err = s.restApiService.CallRestApi(ctx,"POST",path, &s.appServer.RestEndpoint.XApigwId ,credit)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s WorkerService) List(ctx context.Context, credit core.AccountStatement) (*[]core.AccountStatement, error){
	childLogger.Debug().Msg("List")
	childLogger.Debug().Interface("credit:",credit).Msg("")
	
	span := lib.Span(ctx, "service.List")	
    defer span.End()

	path := s.appServer.RestEndpoint.ServiceUrlDomain + "/get/" + credit.AccountID
	rest_interface_data, err := s.restApiService.CallRestApi(ctx,"GET", path, &s.appServer.RestEndpoint.XApigwId, nil)

	if err != nil {
		return nil, err
	}

	var account_parsed core.Account
	err = mapstructure.Decode(rest_interface_data, &account_parsed)
    if err != nil {
		childLogger.Error().Err(err).Msg("error parse interface")
		return nil, errors.New(err.Error())
    }

	credit.FkAccountID = account_parsed.ID
	credit.Type = "CREDIT"

	res, err := s.workerRepo.List(ctx, credit)
	if err != nil {
		return nil, err
	}

	return res, nil
}