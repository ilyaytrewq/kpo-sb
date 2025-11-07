package dbrepo

import (
	"errors"
	"database/sql"
	
	operation "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/Operation"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
)

func NewOperationDBRepo(db *sql.DB) *CommonDBRepo {
    m := entityMapper{
        table:     "operations",
        byIDQuery: `SELECT id, type, bank_account_id, amount, date, description, category_id FROM operations WHERE id = $1`,
        allQuery:  `SELECT id, type, bank_account_id, amount, date, description, category_id FROM operations`,
        insertSQL: `INSERT INTO operations(id,type,bank_account_id,amount,date,description,category_id) 
                    VALUES($1,$2,$3,$4,$5,$6,$7)
                    ON CONFLICT (id) DO UPDATE 
                    SET type=EXCLUDED.type, bank_account_id=EXCLUDED.bank_account_id, amount=EXCLUDED.amount, date=EXCLUDED.date, description=EXCLUDED.description, category_id=EXCLUDED.category_id`,
        deleteSQL: `DELETE FROM operations WHERE id = $1`,
        scanOne: func(s scanner) (service.ICommonObject, error) {
            var (
                id            service.ObjectID
                t             int
                accID         service.ObjectID
                amount        float64
                date          sql.NullTime
                desc          sql.NullString
                categoryID    service.ObjectID
            )
            if err := s.Scan(&id, &t, &accID, &amount, &date, &desc, &categoryID); err != nil {
                return nil, err
            }
            op, err := operation.NewCopyOperation(
                id,
                operation.OperationType(t),
                accID,
                amount,
                date.Time,
                categoryID,
                desc.String,
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
                int(op.Type()),
                op.BankAccountID(),
                op.Amount(),
                op.Date(),
                op.Description(),
                op.CategoryID(),
            }, nil
        },
    }
    return NewCommonDBRepo(db, m)
}