package bankaccountrepo

import (
	"context"
	"errors"

	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	bankaccount "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/BankAccount"
)

type BankAccountRepo struct {
	repo map[service.ObjectID]*bankaccount.BankAccount
}

func NewBankAccountRepo() *BankAccountRepo {
	return &BankAccountRepo{make(map[service.ObjectID]*bankaccount.BankAccount)}
}

func NewCopyBankAccountRepo(repo map[service.ObjectID]*bankaccount.BankAccount) *BankAccountRepo {
	newRepo := make(map[service.ObjectID]*bankaccount.BankAccount)
	for k, v := range repo {
		newRepo[k] = v
	}
	return &BankAccountRepo{repo: newRepo}
}

func (r *BankAccountRepo) ByID(ctx context.Context, id service.ObjectID) (service.ICommonObject, error) {
	acc, ok := r.repo[id]
	if !ok {
		return nil, errors.New("account not found")
	}
	return acc, nil
}

func (r *BankAccountRepo) Save(ctx context.Context, acc service.ICommonObject) error {
	if _, ok := r.repo[acc.ID()]; ok {
		return errors.New("account already saved")
	}
	i, ok := acc.(*bankaccount.BankAccount)
	if !ok {
		return errors.New("invalid account type")
	}
	r.repo[acc.ID()] = i
	return nil
}

func (r *BankAccountRepo) All(ctx context.Context) ([]service.ICommonObject, error) {
	accs := make([]service.ICommonObject, 0, len(r.repo))
	for _, acc := range r.repo {
		accs = append(accs, acc)
	}
	return accs, nil
}

func (r *BankAccountRepo) Delete(ctx context.Context, id service.ObjectID) error {
	if _, ok := r.repo[id]; !ok {
		return errors.New("account not found")
	}
	delete(r.repo, id)
	return nil
}
