package operationrepo

import (
	"context"
	"errors"
	"time"

	repository "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	operation "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/Operation"
)

type IOperationRepo interface {
	repository.ICommonRepo
	SliceByAccountAndPeriod(ctx context.Context, id service.ObjectID, from time.Time, to time.Time) ([]*service.ICommonObject, error)
}

type OperationRepo struct {
	repo map[service.ObjectID]*operation.Operation
}

func NewOperationRepo() *OperationRepo {
	return &OperationRepo{make(map[service.ObjectID]*operation.Operation)}
}

func NewCopyOperationRepo(repo map[service.ObjectID]*operation.Operation) *OperationRepo {
	newRepo := make(map[service.ObjectID]*operation.Operation)
	for k, v := range repo {
		newRepo[k] = v
	}
	return &OperationRepo{repo: newRepo}
}

func (r *OperationRepo) ByID(ctx context.Context, id service.ObjectID) (service.ICommonObject, error) {
	op, ok := r.repo[id]
	if !ok {
		return nil, errors.New("operation not found")
	}
	return op, nil
}

func (r *OperationRepo) Save(ctx context.Context, op service.ICommonObject) error {
	if _, ok := r.repo[op.ID()]; ok {
		return errors.New("operation already saved")
	}
	i, ok := op.(*operation.Operation)
	if !ok {
		return errors.New("invalid operation type")
	}
	r.repo[op.ID()] = i
	return nil
}

func (r *OperationRepo) All(ctx context.Context) ([]service.ICommonObject, error) {
	ops := make([]service.ICommonObject, 0, len(r.repo))
	for _, op := range r.repo {
		ops = append(ops, op)
	}
	return ops, nil
}

func (r *OperationRepo) SliceByAccountAndPeriod(ctx context.Context, _ service.ObjectID, from time.Time, to time.Time) ([]service.ICommonObject, error) {
	ops := make([]service.ICommonObject, 0)
	for _, op := range r.repo {
		d := op.Date()
		if (d.Equal(from) || d.After(from)) && (d.Equal(to) || d.Before(to)) {
			ops = append(ops, op)
		}
	}
	return ops, nil
}
