package database

import (
	"context"
	"time"
	"errors"
	
	"github.com/go-credit/internal/core/model"
	go_core_observ "github.com/eliezerraj/go-core/observability"
	go_core_pg "github.com/eliezerraj/go-core/database/pg"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

var childLogger = log.With().Str("component","go-credit").Str("package","internal.adapter.database").Logger()

var tracerProvider go_core_observ.TracerProvider

type WorkerRepository struct {
	DatabasePGServer *go_core_pg.DatabasePGServer
}

func NewWorkerRepository(databasePGServer *go_core_pg.DatabasePGServer) *WorkerRepository{
	childLogger.Info().Str("func","NewWorkerRepository").Send()

	return &WorkerRepository{
		DatabasePGServer: databasePGServer,
	}
}

// About add credit
func (w WorkerRepository) AddCredit(ctx context.Context, tx pgx.Tx, credit *model.AccountStatement) (*model.AccountStatement, error){
	childLogger.Info().Str("func","AddCredit").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	//trace
	span := tracerProvider.Span(ctx, "database.AddCredit")
	defer span.End()

	// Prepare
	credit.ChargeAt = time.Now()

	// Query e Execute
	query := `INSERT INTO account_statement (fk_account_id, 
											type_charge,
											charged_at, 
											currency,
											amount,
											tenant_id,
											transaction_id) 
			 VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	row := tx.QueryRow(ctx, query, credit.FkAccountID, credit.Type, credit.ChargeAt, credit.Currency, credit.Amount, credit.TenantID, credit.TransactionID)								
	var id int
	if err := row.Scan(&id); err != nil {
		return nil, errors.New(err.Error())
	}

	credit.ID = id

	return credit , nil
}

// About list credit
func (w WorkerRepository) ListCredit(ctx context.Context, credit *model.AccountStatement) (*[]model.AccountStatement, error){
	childLogger.Info().Str("func","ListCredit").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()
	
	// Trace
	span := tracerProvider.Span(ctx, "database.ListCredit")
	defer span.End()

	// Prepare
	conn, err := w.DatabasePGServer.Acquire(ctx)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer w.DatabasePGServer.Release(conn)

	res_accountStatement := model.AccountStatement{}
	res_accountStatement_list := []model.AccountStatement{}

	// Query e Execute
	query := `SELECT id, 
					fk_account_id, 
					type_charge,
					charged_at,
					currency, 
					amount,																										
					tenant_id,
					transaction_id	
				FROM account_statement 
					WHERE fk_account_id =$1 and type_charge= $2 order by charged_at desc`

	rows, err := conn.Query(ctx, query, credit.FkAccountID, credit.Type)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan( 	&res_accountStatement.ID, 
							&res_accountStatement.FkAccountID, 
							&res_accountStatement.Type, 
							&res_accountStatement.ChargeAt,
							&res_accountStatement.Currency,
							&res_accountStatement.Amount,
							&res_accountStatement.TenantID,
							&res_accountStatement.TransactionID,
						)
		if err != nil {
			return nil, errors.New(err.Error())
        }
		res_accountStatement_list = append(res_accountStatement_list, res_accountStatement)
	}
	
	return &res_accountStatement_list , nil
}

// About list credit per date
func (w WorkerRepository) ListCreditPerDate(ctx context.Context, credit *model.AccountStatement) (*[]model.AccountStatement, error){
	childLogger.Info().Str("func","ListCreditPerDate").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	// Trace
	span := tracerProvider.Span(ctx, "database.ListCreditPerDate")
	defer span.End()

	// Prepare 
	conn, err := w.DatabasePGServer.Acquire(ctx)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer w.DatabasePGServer.Release(conn)

	res_accountStatement := model.AccountStatement{}
	res_accountStatement_list := []model.AccountStatement{}

	// Query e Execute
	query := `SELECT id, 
					fk_account_id, 
					type_charge,
					charged_at,
					currency, 
					amount,																										
					tenant_id,
					transaction_id	
			FROM account_statement 
			WHERE fk_account_id =$1 
			and type_charge= $2
			and charged_at >= $3
			order by charged_at desc`

	rows, err := conn.Query(ctx, query, credit.FkAccountID, credit.Type, credit.ChargeAt)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan( 	&res_accountStatement.ID, 
							&res_accountStatement.FkAccountID, 
							&res_accountStatement.Type, 
							&res_accountStatement.ChargeAt,
							&res_accountStatement.Currency,
							&res_accountStatement.Amount,
							&res_accountStatement.TenantID,
							&res_accountStatement.TransactionID,
						)
		if err != nil {
			return nil, errors.New(err.Error())
        }
		res_accountStatement_list = append(res_accountStatement_list, res_accountStatement)
	}
	
	return &res_accountStatement_list , nil
}

// About create a uuid transaction
func (w WorkerRepository) GetTransactionUUID(ctx context.Context) (*string, error){
	childLogger.Info().Str("func","GetTransactionUUID").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()
		
	// Trace
	span := tracerProvider.Span(ctx, "database.GetTransactionUUID")
	defer span.End()

	conn, err := w.DatabasePGServer.Acquire(ctx)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer w.DatabasePGServer.Release(conn)

	// Prepare
	var uuid string

	// Query and Execute
	query := `SELECT uuid_generate_v4()`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&uuid) 
		if err != nil {
			return nil, errors.New(err.Error())
        }
		return &uuid, nil
	}
	
	return &uuid, nil
}