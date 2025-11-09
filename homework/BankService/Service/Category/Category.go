package category

import (
	"errors"

	"github.com/google/uuid"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
)

type CategoryType int

const (
	Spending CategoryType = iota
	Income
)

func (ct CategoryType) String() string {
	switch ct {
	case Spending:
		return "Spending"
	case Income:
		return "Income"
	default:
		return "Unknown"
	}
}

type ICategory interface {
	service.ICommonObject
	Name() string
	Type() CategoryType

	SetName(newName string)
}

type Category struct {
	id    service.ObjectID
	name  string
	ctype CategoryType
}

func NewCategory(name string, ctype CategoryType) (*Category, error) {
	if name == "" {
		return nil, errors.New("category name cannot be empty")
	}
	if ctype != Spending && ctype != Income {
		return nil, errors.New("invalid category type")
	}
	return &Category{
		id:    service.ObjectID(uuid.New()),
		name:  name,
		ctype: ctype,
	}, nil
}

func NewCopyCategory(id service.ObjectID, name string, ctype CategoryType) (*Category, error) {
	if name == "" {
		return nil, errors.New("category name cannot be empty")
	}
	if ctype != Spending && ctype != Income {
		return nil, errors.New("invalid category type")
	}
	return &Category{
		id:    id,
		name:  name,
		ctype: ctype,
	}, nil
}

func (c *Category) ID() service.ObjectID   { return c.id }
func (c *Category) Name() string           { return c.name }
func (c *Category) Type() CategoryType     { return c.ctype }
func (c *Category) SetName(newName string) { c.name = newName }
