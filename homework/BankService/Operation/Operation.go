package operation

import (
	"time"

	"github.com/google/uuid"
	bankaccount "github.com/ilyaytrewq/kpo-sb/homework/BankService/BankAccount"
	category "github.com/ilyaytrewq/kpo-sb/homework/BankService/Category"
)

type OperationID uuid.UUID
type OperationType int

const (
	Spending OperationType = iota
	Income
)

type IOperation interface {
	ID() OperationID
	Type() OperationType
	BankAccountID() bankaccount.BankAccountID
	Amount() float64
	Date() time.Time
	Description() string
	CategoryID() category.CategoryID
}

type Operation struct {
	id            OperationID
	opType        OperationType
	bankAccountID bankaccount.BankAccountID
	amount        float64
	date          time.Time
	description   string
	categoryID    category.CategoryID
}

func NewOperation(
	opType OperationType,
	bankAccountID bankaccount.BankAccountID,
	amount float64,
	date time.Time,
	categoryID category.CategoryID,
	description ...string,
) (*Operation, error) {
	desc := ""
	if len(description) > 0 {
		desc = description[0]
	}

	return &Operation{
		id:            OperationID(uuid.New()),
		opType:        opType,
		bankAccountID: bankAccountID,
		amount:        amount,
		date:          date,
		description:   desc,
		categoryID:    categoryID,
	}, nil
}

func NewCopyOperation(
	id OperationID,
	opType OperationType,
	bankAccountID bankaccount.BankAccountID,
	amount float64,
	date time.Time,
	categoryID category.CategoryID,
	description ...string,
) (*Operation, error) {
	desc := ""
	if len(description) > 0 {
		desc = description[0]
	}

	return &Operation{
		id:            id,
		opType:        opType,
		bankAccountID: bankAccountID,
		amount:        amount,
		date:          date,
		description:   desc,
		categoryID:    categoryID,
	}, nil
}

func (o *Operation) ID() OperationID                          { return o.id }
func (o *Operation) Type() OperationType                      { return o.opType }
func (o *Operation) BankAccountID() bankaccount.BankAccountID { return o.bankAccountID }
func (o *Operation) Amount() float64                          { return o.amount }
func (o *Operation) Date() time.Time                          { return o.date }
func (o *Operation) Description() string                      { return o.description }
func (o *Operation) CategoryID() category.CategoryID          { return o.categoryID }
