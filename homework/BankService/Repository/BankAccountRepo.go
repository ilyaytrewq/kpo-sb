package repository

import (
	"context"
	"errors"

	bankaccount "github.com/ilyaytrewq/kpo-sb/homework/BankService/BankAccount"
)

type IBankAccountRepo interface {
    ByID(ctx context.Context, id bankaccount.BankAccountID) (*bankaccount.IBankAccount, error)
    Save(ctx context.Context, acc *bankaccount.IBankAccount) error
}

type BankAccountRepo struct {
	repo map[bankaccount.BankAccountID] *bankaccount.BankAccount
}

func (r *BankAccountRepo) ByID(ctx context.Context, id bankaccount.BankAccountID) (*bankaccount.IBankAccount, error) {
	acc, ok := r.repo[id]
	if !ok {
		return nil, errors.New("account not found")
	}
	var i bankaccount.IBankAccount = acc
	return &i, nil
}

func (r *BankAccountRepo) Save(ctx context.Context, acc *bankaccount.IBankAccount) error {
	if _, ok := r.repo[(*acc).ID()]; ok {
		return errors.New("account already saved")
	}
	i, ok := (*acc).(*bankaccount.BankAccount)
	if !ok {
		return errors.New("invalid account type")
	}
	r.repo[(*acc).ID()] = i
	return nil
}