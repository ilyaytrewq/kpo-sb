package bankaccount

import (
	"errors"

	"github.com/google/uuid"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
)

type IBankAccount interface {
	service.ICommonObject
	Name() string
	Balance() float64

	SetName(newName string)
	SetBalance(newBalance float64) error
}

type BankAccount struct {
	id      service.ObjectID
	name    string
	balance float64
}

func NewBankAccount(name string, balance float64) (*BankAccount, error) {
	if balance < 0 {
		return nil, errors.New("balance should be >= 0")
	}
	if name == "" {
		return nil, errors.New("account name cannot be empty")
	}
	return &BankAccount{
		id:      service.ObjectID(uuid.New()),
		name:    name,
		balance: balance,
	}, nil
}

func NewCopyBankAccount(id service.ObjectID, name string, balance float64) (*BankAccount, error) {
	if balance < 0 {
		return nil, errors.New("balance should be >= 0")
	}
	if name == "" {
		return nil, errors.New("account name cannot be empty")
	}
	return &BankAccount{
		id:      id,
		name:    name,
		balance: balance,
	}, nil
}

func (acc *BankAccount) ID() service.ObjectID   { return acc.id }
func (acc *BankAccount) Name() string           { return acc.name }
func (acc *BankAccount) Balance() float64       { return acc.balance }
func (acc *BankAccount) SetName(newName string) { acc.name = newName }

func (acc *BankAccount) SetBalance(newBalance float64) error {
	if newBalance < 0 {
		return errors.New("balance should be >= 0")
	}
	acc.balance = newBalance
	return nil
}
