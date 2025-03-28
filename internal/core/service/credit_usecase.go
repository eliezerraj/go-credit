package service

import(
	"fmt"
	"context"
	"net/http"
	"encoding/json"
	"errors"

	"github.com/go-credit/internal/infra/circuitbreaker"
	"github.com/go-credit/internal/core/model"
	"github.com/go-credit/internal/core/erro"
	go_core_observ "github.com/eliezerraj/go-core/observability"
	go_core_api "github.com/eliezerraj/go-core/api"
)

var tracerProvider go_core_observ.TracerProvider
var apiService go_core_api.ApiService

// About handle/convert http status code
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

// About add credit
func (s *WorkerService) AddCredit(ctx context.Context, credit *model.AccountStatement) (*model.AccountStatement, error){
	childLogger.Info().Str("func","AddCredit").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("credit", credit).Send()

	// Trace
	span := tracerProvider.Span(ctx, "service.AddCredit")
	trace_id := fmt.Sprintf("%v",ctx.Value("trace-request-id"))

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

	//Open CB - MOCK
	circuitBreaker := circuitbreaker.CircuitBreakerConfig()
	_, err = circuitBreaker.Execute(func() (interface{}, error) {
		if credit.Type == "CREDIT-CB" {
			return nil, erro.ErrTransInvalid
		}
		return nil , nil
	})
	if (err != nil) {
		spanCB := tracerProvider.Span(ctx, "service.AddCredit-CIRCUIT-BREAKER")

		childLogger.Error().Err(err).Interface("trace-resquest-id", ctx.Value("trace-request-id")).Msg(" ****** Circuit Breaker OPEN !!! ******")

		transfer := model.Transfer{}
		transfer.Currency = credit.Currency
		transfer.Amount = credit.Amount
		transfer.Type = "CREDIT"
		transfer.AccountFrom = credit

		childLogger.Info().Interface("trace_id", trace_id).Interface("=========>>>>> transfer: ",transfer).Msg("<==========")

		_, _, err := apiService.CallApi(ctx,
												s.apiService[2].Url,
												s.apiService[2].Method,
												&s.apiService[2].Header_x_apigw_api_id,
												nil,
												&trace_id, 
												transfer)
		if err != nil {
			return nil, err
		}
		credit.Obs =  "transaction send via circuit breaker !!!"
		
		spanCB.End()
		return credit, nil
	}

	// Business rules
	if credit.Type != "CREDIT" {
		return nil, erro.ErrTransInvalid
	}
	if credit.Amount < 0 {
		return nil, erro.ErrInvalidAmount
	}

	// Get the Account ID (PK) from Account-service
	res_payload, statusCode, err := apiService.CallApi(ctx,
														s.apiService[0].Url + "/" + credit.AccountID,
														s.apiService[0].Method,
														&s.apiService[0].Header_x_apigw_api_id,
														nil,
														&trace_id, 
														nil)
	if err != nil {
		return nil, errorStatusCode(statusCode)
	}

	jsonString, err  := json.Marshal(res_payload)
	if err != nil {
		return nil, errors.New(err.Error())
    }
	var account_parsed model.Account
	json.Unmarshal(jsonString, &account_parsed)

	// Business rule
	credit.FkAccountID = account_parsed.ID

	// Get transaction UUID 
	res_uuid, err := s.workerRepository.GetTransactionUUID(ctx)
	if err != nil {
		return nil, err
	}
	credit.TransactionID = res_uuid

	// Add the credit (create account_statement)
	res, err := s.workerRepository.AddCredit(ctx, tx, credit)
	if err != nil {
		return nil, err
	}

	// Add (POST/AddFundBalanceAccount) the updat account statement 
	_, statusCode, err = apiService.CallApi(ctx,
											s.apiService[1].Url,
											s.apiService[1].Method,
											&s.apiService[1].Header_x_apigw_api_id,
											nil,
											&trace_id,
											credit)
	if err != nil {
		return nil, errorStatusCode(statusCode)
	}

	return res, nil
}

// About list credit
func (s *WorkerService) ListCredit(ctx context.Context, credit *model.AccountStatement) (*[]model.AccountStatement, error){
	childLogger.Info().Str("func","ListCredit").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("credit", credit).Send()

	// Trace
	span := tracerProvider.Span(ctx, "service.ListCredit")
	trace_id := fmt.Sprintf("%v",ctx.Value("trace-request-id"))
	defer span.End()
	
	// Get the Account ID from Account-service
	res_payload, statusCode, err := apiService.CallApi(ctx,
														s.apiService[0].Url + "/" + credit.AccountID,
														s.apiService[0].Method,
														&s.apiService[0].Header_x_apigw_api_id,
														nil,
														&trace_id,
														nil)
	if err != nil {
		return nil, errorStatusCode(statusCode)
	}

	jsonString, err  := json.Marshal(res_payload)
	if err != nil {
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

// About list credit per date
func (s *WorkerService) ListCreditPerDate(ctx context.Context, credit *model.AccountStatement) (*[]model.AccountStatement, error){
	childLogger.Info().Str("func","ListCreditPerDate").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("credit", credit).Send()

	// Trace
	span := tracerProvider.Span(ctx, "service.ListCreditPerDate")
	trace_id := fmt.Sprintf("%v",ctx.Value("trace-request-id"))
	defer span.End()
	
	// Get the Account ID from Account-service
	res_payload, statusCode, err := apiService.CallApi(ctx,
														s.apiService[0].Url + "/" + credit.AccountID,
														s.apiService[0].Method,
														&s.apiService[0].Header_x_apigw_api_id,
														nil,
														&trace_id,
														nil)
	if err != nil {
		return nil, errorStatusCode(statusCode)
	}

	jsonString, err  := json.Marshal(res_payload)
	if err != nil {
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