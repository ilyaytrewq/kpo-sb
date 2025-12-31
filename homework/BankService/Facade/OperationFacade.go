package facade

import (
	"context"
	"errors"
	"time"

	repository "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository"
	operationrepo "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository/OperationRepo"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	operation "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/Operation"
)

type OperationFacade struct {
	repo   repository.ICommonRepo
	opRepo operationrepo.IOperationRepo
}

func NewOperationFacade(repo repository.ICommonRepo) *OperationFacade {
	var op operationrepo.IOperationRepo
	if r, ok := repo.(operationrepo.IOperationRepo); ok {
		op = r
	}
	return &OperationFacade{repo: repo, opRepo: op}
}

func (f *OperationFacade) CreateOperation(
	opType operation.OperationType,
	accountID service.ObjectID,
	amount float64,
	date time.Time,
	categoryID service.ObjectID,
	description ...string,
) (service.ObjectID, error) {
	op, err := operation.NewOperation(opType, accountID, amount, date, categoryID, description...)
	if err != nil {
		return service.ObjectID{}, err
	}
	if err := f.repo.Save(context.Background(), op); err != nil {
		return service.ObjectID{}, err
	}
	return op.ID(), nil
}

func (f *OperationFacade) GetOperation(id service.ObjectID) (operation.IOperation, error) {
	obj, err := f.repo.ByID(context.Background(), id)
	if err != nil {
		return nil, err
	}
	op, ok := obj.(operation.IOperation)
	if !ok {
		return nil, errors.New("invalid type")
	}
	return op, nil
}

func (f *OperationFacade) ListAllOperations() ([]operation.IOperation, error) {
	objs, err := f.repo.All(context.Background())
	if err != nil {
		return nil, err
	}
	var operations []operation.IOperation
	for _, obj := range objs {
		if op, ok := obj.(operation.IOperation); ok {
			operations = append(operations, op)
		}
	}
	return operations, nil
}

func (f *OperationFacade) GetOperationsByPeriod(accountID service.ObjectID, from, to time.Time) ([]operation.IOperation, error) {
	if f.opRepo == nil {
		return nil, errors.New("operation repo does not support period slicing")
	}
	objs, err := f.opRepo.SliceByAccountAndPeriod(context.Background(), accountID, from, to)
	if err != nil {
		return nil, err
	}
	var operations []operation.IOperation
	for _, obj := range objs {
		if op, ok := obj.(operation.IOperation); ok {
			operations = append(operations, op)
		}
	}
	return operations, nil
}

func (f *OperationFacade) DeleteOperation(id service.ObjectID) error {
	return f.repo.Delete(context.Background(), id)
}
