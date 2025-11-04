package categoryrepo

import (
	"context"
	"errors"

	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	category "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/Category"
)

type CategoryRepo struct {
	repo map[service.ObjectID]*category.Category
}

func NewCategory() *CategoryRepo {
	return &CategoryRepo{make(map[service.ObjectID]*category.Category)}
}

func NewCopyCategory(repo map[service.ObjectID]*category.Category) *CategoryRepo {
	newRepo := make(map[service.ObjectID]*category.Category)
	for k, v := range repo {
		newRepo[k] = v
	}
	return &CategoryRepo{repo: newRepo}
}

func (r *CategoryRepo) ByID(ctx context.Context, id service.ObjectID) (*category.ICategory, error) {
	cat, ok := r.repo[id]
	if !ok {
		return nil, errors.New("category not found")
	}
	var i category.ICategory = cat
	return &i, nil
}

func (r *CategoryRepo) Save(ctx context.Context, cat *category.ICategory) error {
	if _, ok := r.repo[(*cat).ID()]; ok {
		return errors.New("category already saved")
	}
	i, ok := (*cat).(*category.Category)
	if !ok {
		return errors.New("invalid category type")
	}
	r.repo[(*cat).ID()] = i
	return nil
}

func (r *CategoryRepo) All(ctx context.Context) ([]*category.ICategory, error) {
	cats := make([]*category.ICategory, 0, len(r.repo))
	for _, cat := range r.repo {
		var i category.ICategory = cat
		cats = append(cats, &i)
	}
	return cats, nil
}
