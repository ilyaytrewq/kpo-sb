package bankaccountrepo

import (
	"context"
	"errors"

	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	bankaccount "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/BankAccount"
)

type BankAccountRepo struct {
	repo map[service.ObjectID] *bankaccount.BankAccount
}

func NewBankAccount() *BankAccountRepo {
	return &BankAccountRepo{make(map[service.ObjectID]*bankaccount.BankAccount)}
}

func NewCopyBankAccount(repo map[service.ObjectID] *bankaccount.BankAccount) *BankAccountRepo {
	newRepo := make(map[service.ObjectID]*bankaccount.BankAccount)
	for k, v := range repo {
		newRepo[k] = v
	}
	return &BankAccountRepo{repo: newRepo}
}

func (r *BankAccountRepo) ByID(ctx context.Context, id service.ObjectID) (*bankaccount.IBankAccount, error) {
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

func (r *BankAccountRepo) All(ctx context.Context) ([]*bankaccount.IBankAccount, error) {
	accs := make([]*bankaccount.IBankAccount, 0, len(r.repo))
	for _, acc := range r.repo {
		var i bankaccount.IBankAccount = acc
		accs = append(accs, &i)
	}
	return accs, nil
}