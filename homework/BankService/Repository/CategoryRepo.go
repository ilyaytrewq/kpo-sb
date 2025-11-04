package repository

import (
	"context"
	"errors"

	category "github.com/ilyaytrewq/kpo-sb/homework/BankService/Category"
)

type ICategoryRepo interface {
	ByID(ctx context.Context, id category.CategoryID) (*category.ICategory, error)
	Save(ctx context.Context, cat *category.ICategory) error
}

type CategoryRepo struct {
	repo map[category.CategoryID]*category.Category
}

func (r *CategoryRepo) ByID(ctx context.Context, id category.CategoryID) (*category.ICategory, error) {
	cat, ok := r.repo[id]
	if !ok {
		return nil, errors.New("account not found")
	}
	var i category.ICategory = cat
	return &i, nil
}

func (r *CategoryRepo) Save(ctx context.Context, cat *category.ICategory) error {
	if _, ok := r.repo[(*cat).ID()]; ok {
		return errors.New("account already saved")
	}
	i, ok := (*cat).(*category.Category)
	if !ok {
		return errors.New("invalid account type")
	}
	r.repo[(*cat).ID()] = i
	return nil
}
