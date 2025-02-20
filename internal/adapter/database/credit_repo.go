package database

import (
	"context"
	"time"
	"errors"
	
	"github.com/go-credit/internal/core/model"
	//"github.com/go-credit/internal/core/erro"

	go_core_observ "github.com/eliezerraj/go-core/observability"
	go_core_pg "github.com/eliezerraj/go-core/database/pg"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

var tracerProvider go_core_observ.TracerProvider
var childLogger = log.With().Str("adapter", "database").Logger()

type WorkerRepository struct {
	DatabasePGServer *go_core_pg.DatabasePGServer
}

func NewWorkerRepository(databasePGServer *go_core_pg.DatabasePGServer) *WorkerRepository{
	childLogger.Debug().Msg("NewWorkerRepository")

	return &WorkerRepository{
		DatabasePGServer: databasePGServer,
	}
}

func (w WorkerRepository) AddCredit(ctx context.Context, tx pgx.Tx, credit *model.AccountStatement) (*model.AccountStatement, error){
	childLogger.Debug().Msg("AddCredit")

	span := tracerProvider.Span(ctx, "database.AddCredit")
	defer span.End()

	query := `INSERT INTO account_statement (fk_account_id, 
											type_charge,
											charged_at, 
											currency,
											amount,
											tenant_id) 
			 VALUES($1, $2, $3, $4, $5, $6) RETURNING id`

	credit.ChargeAt = time.Now()

	row := tx.QueryRow(ctx, query, credit.FkAccountID, credit.Type, credit.ChargeAt, credit.Currency, credit.Amount, credit.TenantID)								
	var id int
	if err := row.Scan(&id); err != nil {
		return nil, errors.New(err.Error())
	}

	credit.ID = id

	return credit , nil
}

func (w WorkerRepository) ListCredit(ctx context.Context, credit *model.AccountStatement) (*[]model.AccountStatement, error){
	childLogger.Debug().Msg("ListCredit")
	
	span := tracerProvider.Span(ctx, "database.ListCredit")
	defer span.End()

	conn, err := w.DatabasePGServer.Acquire(ctx)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer w.DatabasePGServer.Release(conn)

	res_accountStatement := model.AccountStatement{}
	res_accountStatement_list := []model.AccountStatement{}

	query := `SELECT id, 
					fk_account_id, 
					type_charge,
					charged_at,
					currency, 
					amount,																										
					tenant_id	
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
						)
		if err != nil {
			return nil, errors.New(err.Error())
        }
		res_accountStatement_list = append(res_accountStatement_list, res_accountStatement)
	}
	
	return &res_accountStatement_list , nil
}

func (w WorkerRepository) ListCreditPerDate(ctx context.Context, credit *model.AccountStatement) (*[]model.AccountStatement, error){
	childLogger.Debug().Msg("ListCreditPerDate")
	
	span := tracerProvider.Span(ctx, "database.ListCreditPerDate")
	defer span.End()

	conn, err := w.DatabasePGServer.Acquire(ctx)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer w.DatabasePGServer.Release(conn)

	res_accountStatement := model.AccountStatement{}
	res_accountStatement_list := []model.AccountStatement{}

	query := `SELECT id, 
					fk_account_id, 
					type_charge,
					charged_at,
					currency, 
					amount,																										
					tenant_id	
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
						)
		if err != nil {
			return nil, errors.New(err.Error())
        }
		res_accountStatement_list = append(res_accountStatement_list, res_accountStatement)
	}
	
	return &res_accountStatement_list , nil
}