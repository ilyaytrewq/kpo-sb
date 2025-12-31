package facade

import (
	"context"
	"errors"

	repository "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	category "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/Category"
)

type CategoryFacade struct {
	repo repository.ICommonRepo
}

func NewCategoryFacade(repo repository.ICommonRepo) *CategoryFacade {
	return &CategoryFacade{repo: repo}
}

func (f *CategoryFacade) CreateCategory(name string, ctype category.CategoryType) (service.ObjectID, error) {
	cat, err := category.NewCategory(name, ctype)
	if err != nil {
		return service.ObjectID{}, err
	}
	if err := f.repo.Save(context.Background(), cat); err != nil {
		return service.ObjectID{}, err
	}
	return cat.ID(), nil
}

func (f *CategoryFacade) GetCategory(id service.ObjectID) (category.ICategory, error) {
	obj, err := f.repo.ByID(context.Background(), id)
	if err != nil {
		return nil, err
	}
	cat, ok := obj.(category.ICategory)
	if !ok {
		return nil, errors.New("invalid type")
	}
	return cat, nil
}

func (f *CategoryFacade) UpdateCategoryName(id service.ObjectID, newName string) error {
	obj, err := f.repo.ByID(context.Background(), id)
	if err != nil {
		return err
	}
	cat, ok := obj.(*category.Category)
	if !ok {
		return errors.New("invalid type")
	}
	cat.SetName(newName)
	return nil
}

func (f *CategoryFacade) ListAllCategories() ([]category.ICategory, error) {
	objs, err := f.repo.All(context.Background())
	if err != nil {
		return nil, err
	}
	var categories []category.ICategory
	for _, obj := range objs {
		if cat, ok := obj.(category.ICategory); ok {
			categories = append(categories, cat)
		}
	}
	return categories, nil
}

func (f *CategoryFacade) DeleteCategory(id service.ObjectID) error {
	return f.repo.Delete(context.Background(), id)
}
