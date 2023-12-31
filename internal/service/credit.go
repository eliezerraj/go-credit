package service

import (
	"context"
	"errors"
	"github.com/rs/zerolog/log"

	"github.com/sony/gobreaker"
	"github.com/mitchellh/mapstructure"
	"github.com/go-credit/internal/core"
	"github.com/go-credit/internal/erro"
	"github.com/go-credit/internal/adapter/restapi"
	"github.com/go-credit/internal/repository/postgre"
	"github.com/aws/aws-xray-sdk-go/xray"

)

var childLogger = log.With().Str("service", "service").Logger()

type WorkerService struct {
	workerRepository 		*postgre.WorkerRepository
	restapi					*restapi.RestApiSConfig
	circuitBreaker			*gobreaker.CircuitBreaker
}

func NewWorkerService(	workerRepository 	*postgre.WorkerRepository,
						restapi				*restapi.RestApiSConfig,
						circuitBreaker	*gobreaker.CircuitBreaker) *WorkerService{
	childLogger.Debug().Msg("NewWorkerService")

	return &WorkerService{
		workerRepository:	workerRepository,
		restapi:			restapi,
		circuitBreaker: 	circuitBreaker,
	}
}

func (s WorkerService) SetSessionVariable(ctx context.Context, userCredential string) (bool, error){
	childLogger.Debug().Msg("SetSessionVariable")

	res, err := s.workerRepository.SetSessionVariable(ctx, userCredential)
	if err != nil {
		return false, err
	}

	return res, nil
}

func (s WorkerService) Add(ctx context.Context, credit core.AccountStatement) (*core.AccountStatement, error){
	childLogger.Debug().Msg("Add")

	_, root := xray.BeginSubsegment(ctx, "Service.Add")

	tx, err := s.workerRepository.StartTx(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
		root.Close(nil)
	}()

	childLogger.Debug().Interface("credit:",credit).Msg("")

	// Mock Circuit Breaker
	_, err = s.circuitBreaker.Execute(func() (interface{}, error) {
		if credit.Type == "CREDITX" {
			return nil, erro.ErrInvalid
		}
		return nil , nil
	})

	if (err != nil) {
		_, segCB := xray.BeginSubsegment(ctx, "Service.CIRCUIT-BREAKER")

		childLogger.Debug().Msg("--------------------------------------------------")
		childLogger.Error().Err(err).Msg(" ****** Circuit Breaker OPEN !!! ******")
		childLogger.Debug().Msg("--------------------------------------------------")

		serverUrlDomainCB := "https://go-fund-transfer.architecture.caradhras.io"
		xApigwIdCP := ""

		transfer := core.Transfer{}
		transfer.Currency = credit.Currency
		transfer.Amount = credit.Amount
		transfer.AccountIDTo = credit.AccountID

		_, err = s.restapi.PostData(ctx, serverUrlDomainCB, xApigwIdCP, "/creditFundSchedule", transfer)
		if err != nil {
			return nil, err
		}

		credit.Obs =  "Transação enviada via Circuit Breaker !!!!"
		
		segCB.Close(nil)
		return &credit, nil
	}
	// Mock Circuit Breaker

	credit.Type = "CREDIT"
	if credit.Amount < 0 {
		return nil, erro.ErrInvalidAmount
	}

	rest_interface_data, err := s.restapi.GetData(ctx, s.restapi.ServerUrlDomain, s.restapi.XApigwId, "/get", credit.AccountID)
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
	res, err := s.workerRepository.Add(ctx, tx, credit)
	if err != nil {
		return nil, err
	}

	childLogger.Debug().Interface("credit:",credit).Msg("")

	_, err = s.restapi.PostData(ctx, s.restapi.ServerUrlDomain, s.restapi.XApigwId, "/add/fund", credit)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s WorkerService) List(ctx context.Context, credit core.AccountStatement) (*[]core.AccountStatement, error){
	childLogger.Debug().Msg("List")
	childLogger.Debug().Interface("credit:",credit).Msg("")
	
	_, root := xray.BeginSubsegment(ctx, "Service.List")
	defer root.Close(nil)

	rest_interface_data, err := s.restapi.GetData(ctx, s.restapi.ServerUrlDomain, s.restapi.XApigwId, "/get", credit.AccountID)
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

	res, err := s.workerRepository.List(ctx, credit)
	if err != nil {
		return nil, err
	}

	return res, nil
}