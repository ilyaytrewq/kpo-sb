package facade

import (
	"context"
	"errors"

	repository "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	bankaccount "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/BankAccount"
)

type BankAccountFacade struct {
	repo repository.ICommonRepo
}

func NewBankAccountFacade(repo repository.ICommonRepo) *BankAccountFacade {
	return &BankAccountFacade{repo: repo}
}

func (f *BankAccountFacade) CreateAccount(name string, balance float64) (service.ObjectID, error) {
	acc, err := bankaccount.NewBankAccount(name, balance)
	if err != nil {
		return service.ObjectID{}, err
	}
	if err := f.repo.Save(context.Background(), acc); err != nil {
		return service.ObjectID{}, err
	}
	return acc.ID(), nil
}

func (f *BankAccountFacade) GetAccount(id service.ObjectID) (bankaccount.IBankAccount, error) {
	obj, err := f.repo.ByID(context.Background(), id)
	if err != nil {
		return nil, err
	}
	acc, ok := obj.(bankaccount.IBankAccount)
	if !ok {
		return nil, errors.New("invalid type")
	}
	return acc, nil
}

func (f *BankAccountFacade) UpdateAccountName(id service.ObjectID, newName string) error {
	obj, err := f.repo.ByID(context.Background(), id)
	if err != nil {
		return err
	}
	acc, ok := obj.(*bankaccount.BankAccount)
	if !ok {
		return errors.New("invalid type")
	}
	acc.SetName(newName)
	return nil
}

func (f *BankAccountFacade) UpdateAccountBalance(id service.ObjectID, newBalance float64) error {
	obj, err := f.repo.ByID(context.Background(), id)
	if err != nil {
		return err
	}
	acc, ok := obj.(*bankaccount.BankAccount)
	if !ok {
		return errors.New("invalid type")
	}
	return acc.SetBalance(newBalance)
}

func (f *BankAccountFacade) ListAllAccounts() ([]bankaccount.IBankAccount, error) {
	objs, err := f.repo.All(context.Background())
	if err != nil {
		return nil, err
	}
	var accounts []bankaccount.IBankAccount
	for _, obj := range objs {
		if acc, ok := obj.(bankaccount.IBankAccount); ok {
			accounts = append(accounts, acc)
		}
	}
	return accounts, nil
}

func (f *BankAccountFacade) DeleteAccount(id service.ObjectID) error {
	return f.repo.Delete(context.Background(), id)
}
