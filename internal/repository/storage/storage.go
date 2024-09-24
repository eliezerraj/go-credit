package storage

import (
	"context"
	"time"
	"errors"

	"github.com/go-credit/internal/repository/pg"
	"github.com/go-credit/internal/core"
	"github.com/go-credit/internal/lib"

	"github.com/rs/zerolog/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var childLogger = log.With().Str("repository.pg", "storage").Logger()

type WorkerRepository struct {
	databasePG pg.DatabasePG
}

func NewWorkerRepository(databasePG pg.DatabasePG) WorkerRepository {
	childLogger.Debug().Msg("NewWorkerRepository")
	return WorkerRepository{
		databasePG: databasePG,
	}
}

func (w WorkerRepository) StartTx(ctx context.Context) (pgx.Tx, *pgxpool.Conn,error) {
	childLogger.Debug().Msg("StartTx")

	span := lib.Span(ctx, "storage.StartTx")
	defer span.End()

	conn, err := w.databasePG.Acquire(ctx)
	if err != nil {
		childLogger.Error().Err(err).Msg("erro acquire")
		return nil, nil, errors.New(err.Error())
	}

	tx, err := conn.Begin(ctx)
    if err != nil {
        return nil, nil ,errors.New(err.Error())
    }

	return tx, conn, nil
}

func (w WorkerRepository) ReleaseTx(connection *pgxpool.Conn) {
	childLogger.Debug().Msg("ReleaseTx")

	defer connection.Release()
}
//-----------------------------------------------

func (w WorkerRepository) Add(ctx context.Context, tx pgx.Tx, credit *core.AccountStatement) (*core.AccountStatement, error){
	childLogger.Debug().Msg("Add")

	span := lib.Span(ctx, "storage.Add")	
    defer span.End()
	
	credit.ChargeAt = time.Now()

	query := `INSERT INTO account_statement (fk_account_id, 
											type_charge,
											charged_at, 
											currency,
											amount,
											tenant_id) 
									VALUES($1, $2, $3, $4, $5, $6) RETURNING id`

	row := tx.QueryRow(ctx, query, credit.FkAccountID, credit.Type, credit.ChargeAt, credit.Currency, credit.Amount, credit.TenantID)								
	
	var id int
	if err := row.Scan(&id); err != nil {
		childLogger.Error().Err(err).Msg("error insert statement")
		return nil, errors.New(err.Error())
	}

	credit.ID = id

	return credit , nil
}

func (w WorkerRepository) List(ctx context.Context, credit *core.AccountStatement) (*[]core.AccountStatement, error){
	childLogger.Debug().Msg("List")

	span := lib.Span(ctx, "storage.List")	
    defer span.End()

	conn, err := w.databasePG.Acquire(ctx)
	if err != nil {
		childLogger.Error().Err(err).Msg("erro acquire")
		return nil, errors.New(err.Error())
	}
	defer w.databasePG.Release(conn)
	
	result_query := core.AccountStatement{}
	balance_list := []core.AccountStatement{}

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
		childLogger.Error().Err(err).Msg("errt select statement")
		return nil, errors.New(err.Error())
	}

	for rows.Next() {
		err := rows.Scan( 	&result_query.ID, 
							&result_query.FkAccountID, 
							&result_query.Type, 
							&result_query.ChargeAt,
							&result_query.Currency,
							&result_query.Amount,
							&result_query.TenantID,
						)
		if err != nil {
			childLogger.Error().Err(err).Msg("erro scan statement")
			return nil, errors.New(err.Error())
        }
		balance_list = append(balance_list, result_query)
	}

	defer rows.Close()
	return &balance_list , nil
}

func (w WorkerRepository) ListPerDate(ctx context.Context, credit *core.AccountStatement) (*[]core.AccountStatement, error){
	childLogger.Debug().Msg("ListPerDate")

	span := lib.Span(ctx, "storage.ListPerDate")	
    defer span.End()

	conn, err := w.databasePG.Acquire(ctx)
	if err != nil {
		childLogger.Error().Err(err).Msg("erro acquire")
		return nil, errors.New(err.Error())
	}
	defer w.databasePG.Release(conn)
	
	result_query := core.AccountStatement{}
	balance_list := []core.AccountStatement{}

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
		childLogger.Error().Err(err).Msg("erro select statement")
		return nil, errors.New(err.Error())
	}

	for rows.Next() {
		err := rows.Scan( 	&result_query.ID, 
							&result_query.FkAccountID, 
							&result_query.Type, 
							&result_query.ChargeAt,
							&result_query.Currency,
							&result_query.Amount,
							&result_query.TenantID,
						)
		if err != nil {
			childLogger.Error().Err(err).Msg("erro scan statement")
			return nil, errors.New(err.Error())
        }
		balance_list = append(balance_list, result_query)
	}

	defer rows.Close()
	return &balance_list , nil
}