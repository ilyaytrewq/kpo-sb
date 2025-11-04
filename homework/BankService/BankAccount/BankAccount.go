package bankaccount

import (
	"errors"

	"github.com/google/uuid"
)

type BankAccountID uuid.UUID

type IBankAccount interface {
	ID() BankAccountID
	Name() string
	Balance() float64

	SetName(newName string)
	SetBalance(newBalance float64) error
}

type BankAccount struct {
	id      BankAccountID
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
		id:      BankAccountID(uuid.New()),
		name:    name,
		balance: balance,
	}, nil
}

func NewCopyBankAccount(id BankAccountID, name string, balance float64) (*BankAccount, error) {
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

func (acc *BankAccount) ID() BankAccountID     { return acc.id }
func (acc *BankAccount) Name() string          { return acc.name }
func (acc *BankAccount) Balance() float64      { return acc.balance }
func (acc *BankAccount) SetName(newName string) { acc.name = newName }

func (acc *BankAccount) SetBalance(newBalance float64) error {
	if newBalance < 0 {
		return errors.New("balance should be >= 0")
	}
	acc.balance = newBalance
	return nil
}
