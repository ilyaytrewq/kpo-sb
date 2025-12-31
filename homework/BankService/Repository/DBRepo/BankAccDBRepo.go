package dbrepo

import (
	"database/sql"
	"errors"

	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	bankaccount "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/BankAccount"
)

func NewBankAccountDBRepo(db *sql.DB) *CommonDBRepo {
	m := entityMapper{
		table:     "bank_accounts",
		byIDQuery: `SELECT id, name, balance FROM bank_accounts WHERE id = $1`,
		allQuery:  `SELECT id, name, balance FROM bank_accounts`,
		insertSQL: `INSERT INTO bank_accounts(id,name,balance) VALUES($1,$2,$3) ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name, balance=EXCLUDED.balance`,
		deleteSQL: `DELETE FROM bank_accounts WHERE id = $1`,
		scanOne: func(s scanner) (service.ICommonObject, error) {
			var id service.ObjectID
			var name string
			var balance float64
			if err := s.Scan(&id, &name, &balance); err != nil {
				return nil, err
			}
			acc, err := bankaccount.NewCopyBankAccount(id, name, balance)
			if err != nil {
				return nil, err
			}
			return acc, nil
		},
		argsForInsert: func(obj service.ICommonObject) ([]any, error) {
			acc, ok := obj.(bankaccount.IBankAccount)
			if !ok {
				return nil, errors.New("expected IBankAccount")
			}
			return []any{acc.ID(), acc.Name(), acc.Balance()}, nil
		},
	}
	return NewCommonDBRepo(db, m)
}
