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

func NewCategoryRepo() *CategoryRepo {
	return &CategoryRepo{make(map[service.ObjectID]*category.Category)}
}

func NewCopyCategoryRepo(repo map[service.ObjectID]*category.Category) *CategoryRepo {
	newRepo := make(map[service.ObjectID]*category.Category)
	for k, v := range repo {
		newRepo[k] = v
	}
	return &CategoryRepo{repo: newRepo}
}

func (r *CategoryRepo) ByID(ctx context.Context, id service.ObjectID) (service.ICommonObject, error) {
	cat, ok := r.repo[id]
	if !ok {
		return nil, errors.New("category not found")
	}
	return cat, nil
}

func (r *CategoryRepo) Save(ctx context.Context, cat service.ICommonObject) error {
	if _, ok := r.repo[cat.ID()]; ok {
		return errors.New("category already saved")
	}
	i, ok := cat.(*category.Category)
	if !ok {
		return errors.New("invalid category type")
	}
	r.repo[cat.ID()] = i
	return nil
}

func (r *CategoryRepo) All(ctx context.Context) ([]service.ICommonObject, error) {
	cats := make([]service.ICommonObject, 0, len(r.repo))
	for _, cat := range r.repo {
		cats = append(cats, cat)
	}
	return cats, nil
}

func (r *CategoryRepo) Delete(ctx context.Context, id service.ObjectID) error {
	if _, ok := r.repo[id]; !ok {
		return errors.New("category not found")
	}
	delete(r.repo, id)
	return nil
}
