package service

import(
	"context"
	"net/http"
	"encoding/json"
	"errors"

	"github.com/go-credit/internal/core/model"
	"github.com/go-credit/internal/core/erro"
	go_core_observ "github.com/eliezerraj/go-core/observability"
	go_core_api "github.com/eliezerraj/go-core/api"
)

var tracerProvider go_core_observ.TracerProvider
var apiService go_core_api.ApiService

func errorStatusCode(statusCode int) error{
	var err error
	switch statusCode {
	case http.StatusUnauthorized:
		err = erro.ErrUnauthorized
	case http.StatusForbidden:
		err = erro.ErrHTTPForbiden
	case http.StatusNotFound:
		err = erro.ErrNotFound
	default:
		err = erro.ErrServer
	}
	return err
}

func (s *WorkerService) AddCredit(ctx context.Context, credit *model.AccountStatement) (*model.AccountStatement, error){
	childLogger.Debug().Msg("AddCredit")
	childLogger.Debug().Interface("credit: ",credit).Msg("")

	// Trace
	span := tracerProvider.Span(ctx, "service.AddCredit")
	defer span.End()
	
	// Get the database connection
	tx, conn, err := s.workerRepository.DatabasePGServer.StartTx(ctx)
	if err != nil {
		return nil, err
	}
	
	// Handle the transaction
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
		s.workerRepository.DatabasePGServer.ReleaseTx(conn)
		span.End()
	}()

	// Business rules
	if credit.Type != "CREDIT" {
		return nil, erro.ErrTransInvalid
	}
	if credit.Amount < 0 {
		return nil, erro.ErrInvalidAmount
	}

	// Get the Account ID from Account-service
	res_payload, statusCode, err := apiService.CallApi(ctx,
														s.apiService[0].Url + "/" + credit.AccountID,
														s.apiService[0].Method,
														&s.apiService[0].Header_x_apigw_api_id,
														nil, 
														nil)
	if err != nil {
		return nil, errorStatusCode(statusCode)
	}

	jsonString, err  := json.Marshal(res_payload)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Marshal")
		return nil, errors.New(err.Error())
    }
	var account_parsed model.Account
	json.Unmarshal(jsonString, &account_parsed)

	// Business rule
	credit.FkAccountID = account_parsed.ID

	// Add the credit
	res, err := s.workerRepository.AddCredit(ctx, tx, credit)
	if err != nil {
		return nil, err
	}

	// Add (POST) the account statement Get the Account ID from Account-service
	_, statusCode, err = apiService.CallApi(ctx,
											s.apiService[1].Url,
											s.apiService[1].Method,
											&s.apiService[1].Header_x_apigw_api_id,
											nil, 
											credit)
	if err != nil {
		return nil, errorStatusCode(statusCode)
	}

	return res, nil
}

func (s *WorkerService) ListCredit(ctx context.Context, credit *model.AccountStatement) (*[]model.AccountStatement, error){
	childLogger.Debug().Msg("ListCredit")
	childLogger.Debug().Interface("credit: ",credit).Msg("")

	// Trace
	span := tracerProvider.Span(ctx, "service.ListCredit")
	defer span.End()
	
	// Get the Account ID from Account-service
	res_payload, statusCode, err := apiService.CallApi(ctx,
														s.apiService[0].Url + "/" + credit.AccountID,
														s.apiService[0].Method,
														&s.apiService[0].Header_x_apigw_api_id,
														nil, 
														nil)
	if err != nil {
		return nil, errorStatusCode(statusCode)
	}

	jsonString, err  := json.Marshal(res_payload)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Marshal")
		return nil, errors.New(err.Error())
    }
	var account_parsed model.Account
	json.Unmarshal(jsonString, &account_parsed)

	// Business rule
	credit.FkAccountID = account_parsed.ID
	credit.Type = "CREDIT"

	res, err := s.workerRepository.ListCredit(ctx, credit)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *WorkerService) ListCreditPerDate(ctx context.Context, credit *model.AccountStatement) (*[]model.AccountStatement, error){
	childLogger.Debug().Msg("ListCreditPerDate")
	childLogger.Debug().Interface("credit: ",credit).Msg("")

	// Trace
	span := tracerProvider.Span(ctx, "service.ListCreditPerDate")
	defer span.End()
	
	// Get the Account ID from Account-service
	res_payload, statusCode, err := apiService.CallApi(ctx,
														s.apiService[0].Url + "/" + credit.AccountID,
														s.apiService[0].Method,
														&s.apiService[0].Header_x_apigw_api_id,
														nil, 
														nil)
	if err != nil {
		return nil, errorStatusCode(statusCode)
	}

	jsonString, err  := json.Marshal(res_payload)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Marshal")
		return nil, errors.New(err.Error())
    }
	var account_parsed model.Account
	json.Unmarshal(jsonString, &account_parsed)

	// Business rule
	credit.FkAccountID = account_parsed.ID
	credit.Type = "CREDIT"

	res, err := s.workerRepository.ListCreditPerDate(ctx, credit)
	if err != nil {
		return nil, err
	}
	return res, nil
}