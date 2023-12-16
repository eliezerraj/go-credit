package postgre

import (
	"context"
	"time"
	"errors"

	_ "github.com/lib/pq"
	"database/sql"

	"github.com/go-credit/internal/core"
	"github.com/aws/aws-xray-sdk-go/xray"

)

func (w WorkerRepository) Add(ctx context.Context, tx *sql.Tx ,credit core.AccountStatement) (*core.AccountStatement, error){
	childLogger.Debug().Msg("Add")

	_, root := xray.BeginSubsegment(ctx, "Repository.Add.Credit")
	defer func() {
		root.Close(nil)
	}()

	stmt, err := tx.Prepare(`INSERT INTO account_statement ( 	fk_account_id, 
																type_charge,
																charged_at, 
																currency,
																amount,
																tenant_id) 
									VALUES($1, $2, $3, $4, $5, $6) `)
	if err != nil {
		childLogger.Error().Err(err).Msg("INSERT statement")
		return nil, errors.New(err.Error())
	}

	_, err = stmt.ExecContext(	ctx,
								credit.FkAccountID, 
								credit.Type,
								time.Now(),
								credit.Currency,
								credit.Amount,
								credit.TenantID)
	if err != nil {
		childLogger.Error().Err(err).Msg("Exec statement")
		return nil, errors.New(err.Error())
	}

	defer stmt.Close()
	return &credit , nil
}

func (w WorkerRepository) List(ctx context.Context, credit core.AccountStatement) (*[]core.AccountStatement, error){
	childLogger.Debug().Msg("List")

	_, root := xray.BeginSubsegment(ctx, "Repository.List.Credit")
	defer func() {
		root.Close(nil)
	}()

	client:= w.databaseHelper.GetConnection()
	
	result_query := core.AccountStatement{}
	balance_list := []core.AccountStatement{}

	rows, err := client.QueryContext(ctx, `SELECT 	id, 
													fk_account_id, 
													type_charge,
													charged_at,
													currency, 
													amount,																										
													tenant_id	
											FROM account_statement 
											WHERE fk_account_id =$1 order by charged_at desc`, credit.FkAccountID)
	if err != nil {
		childLogger.Error().Err(err).Msg("SELECT statement")
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
			childLogger.Error().Err(err).Msg("Scan statement")
			return nil, errors.New(err.Error())
        }
		balance_list = append(balance_list, result_query)
	}

	defer rows.Close()
	return &balance_list , nil
}