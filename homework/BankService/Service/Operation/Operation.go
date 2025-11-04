package operation

import (
	"time"

	"github.com/google/uuid"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"

)

type OperationType int

const (
	Spending OperationType = iota
	Income
)

type IOperation interface {
	service.ICommonObject
	Type() OperationType
	BankAccountID() service.ObjectID
	Amount() float64
	Date() time.Time
	Description() string
	CategoryID() service.ObjectID
}

type Operation struct {
	id            service.ObjectID
	opType        OperationType
	bankAccountID service.ObjectID
	amount        float64
	date          time.Time
	description   string
	categoryID    service.ObjectID
}

func NewOperation(
	opType OperationType,
	bankAccountID service.ObjectID,
	amount float64,
	date time.Time,
	categoryID service.ObjectID,
	description ...string,
) (*Operation, error) {
	desc := ""
	if len(description) > 0 {
		desc = description[0]
	}

	return &Operation{
		id:            service.ObjectID(uuid.New()),
		opType:        opType,
		bankAccountID: bankAccountID,
		amount:        amount,
		date:          date,
		description:   desc,
		categoryID:    categoryID,
	}, nil
}

func NewCopyOperation(
	id service.ObjectID,
	opType OperationType,
	bankAccountID service.ObjectID,
	amount float64,
	date time.Time,
	categoryID service.ObjectID,
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

func (o *Operation) ID() service.ObjectID                          { return o.id }
func (o *Operation) Type() OperationType                      { return o.opType }
func (o *Operation) BankAccountID() service.ObjectID         { return o.bankAccountID }
func (o *Operation) Amount() float64                          { return o.amount }
func (o *Operation) Date() time.Time                          { return o.date }
func (o *Operation) Description() string                      { return o.description }
func (o *Operation) CategoryID() service.ObjectID            { return o.categoryID }
