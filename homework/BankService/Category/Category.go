package category

import (
	"errors"

	"github.com/google/uuid"
)

type CategoryID uuid.UUID
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
	ID() CategoryID
	Name() string
	Type() CategoryType

	SetName(newName string)
}

type Category struct {
	id    CategoryID
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
		id:    CategoryID(uuid.New()),
		name:  name,
		ctype: ctype,
	}, nil
}

func NewCopyCategory(id CategoryID, name string, ctype CategoryType) (*Category, error) {
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

func (c *Category) ID() CategoryID       { return c.id }
func (c *Category) Name() string         { return c.name }
func (c *Category) Type() CategoryType   { return c.ctype }
func (c *Category) SetName(newName string) { c.name = newName }
