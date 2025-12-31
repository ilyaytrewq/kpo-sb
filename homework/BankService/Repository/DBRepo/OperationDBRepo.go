package dbrepo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	operation "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/Operation"
)

type OperationDBRepo struct{ *CommonDBRepo }

func NewOperationDBRepo(db *sql.DB) *OperationDBRepo {
	m := entityMapper{
		table: "operations",

		byIDQuery: `SELECT id, op_type, account_id, amount, "timestamp", description, category_id
                      FROM operations
                     WHERE id = $1`,

		allQuery: `SELECT id, op_type, account_id, amount, "timestamp", description, category_id
                      FROM operations`,

		insertSQL: `INSERT INTO operations
                        (id, op_type, account_id, amount, "timestamp", description, category_id)
                    VALUES ($1, $2, $3, $4, $5, $6, $7)
                    ON CONFLICT (id) DO UPDATE SET
                        op_type     = EXCLUDED.op_type,
                        account_id  = EXCLUDED.account_id,
                        amount      = EXCLUDED.amount,
                        "timestamp" = EXCLUDED."timestamp",
                        description = EXCLUDED.description,
                        category_id = EXCLUDED.category_id`,

		deleteSQL: `DELETE FROM operations WHERE id = $1`,

		scanOne: func(s scanner) (service.ICommonObject, error) {
			var (
				idStr, accIDStr, catIDStr string
				t                         int
				amount                    float64
				ts                        time.Time
				desc                      string
			)

			if err := s.Scan(&idStr, &t, &accIDStr, &amount, &ts, &desc, &catIDStr); err != nil {
				return nil, err
			}

			idUUID, err := uuid.Parse(idStr)
			if err != nil {
				return nil, err
			}
			accUUID, err := uuid.Parse(accIDStr)
			if err != nil {
				return nil, err
			}
			catUUID, err := uuid.Parse(catIDStr)
			if err != nil {
				return nil, err
			}

			op, err := operation.NewCopyOperation(
				service.ObjectID(idUUID),
				operation.OperationType(t),
				service.ObjectID(accUUID),
				amount,
				ts,
				service.ObjectID(catUUID),
				desc,
			)
			if err != nil {
				return nil, err
			}
			return op, nil
		},

		argsForInsert: func(obj service.ICommonObject) ([]any, error) {
			op, ok := obj.(operation.IOperation)
			if !ok {
				return nil, errors.New("expected IOperation")
			}
			return []any{
				op.ID(),
				int(op.Type()), // SMALLINT в БД — норм принять как int
				op.BankAccountID(),
				op.Amount(),
				op.Date(), // time.Time → timestamptz
				op.Description(),
				op.CategoryID(),
			}, nil
		},
	}

	return &OperationDBRepo{NewCommonDBRepo(db, m)}
}

func (r *OperationDBRepo) SliceByAccountAndPeriod(ctx context.Context, id service.ObjectID, from time.Time, to time.Time) ([]service.ICommonObject, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, op_type, account_id, amount, "timestamp", description, category_id
       FROM operations
      WHERE account_id = $1
        AND "timestamp" >= $2
        AND "timestamp" <= $3`,
		id, from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []service.ICommonObject
	for rows.Next() {
		obj, err := r.mapper.scanOne(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, obj)
	}
	return out, rows.Err()
}
