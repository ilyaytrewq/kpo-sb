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
	SliceByAccountAndPeriod(ctx context.Context, id service.ObjectID, from time.Time, to time.Time) ([]*operation.IOperation, error)
}

type OperationRepo struct {
	repo map[service.ObjectID]*operation.Operation
}

func NewOperation() *OperationRepo {
	return &OperationRepo{make(map[service.ObjectID]*operation.Operation)}
}

func NewCopyOperation(repo map[service.ObjectID]*operation.Operation) *OperationRepo {
	newRepo := make(map[service.ObjectID]*operation.Operation)
	for k, v := range repo {
		newRepo[k] = v
	}
	return &OperationRepo{repo: newRepo}
}

func (r *OperationRepo) ByID(ctx context.Context, id service.ObjectID) (*operation.IOperation, error) {
	op, ok := r.repo[id]
	if !ok {
		return nil, errors.New("operation not found")
	}
	var i operation.IOperation = op
	return &i, nil
}

func (r *OperationRepo) Save(ctx context.Context, op operation.IOperation) error {
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

func (r *OperationRepo) All(ctx context.Context) ([]*operation.IOperation, error) {
	ops := make([]*operation.IOperation, 0, len(r.repo))
	for _, op := range r.repo {
		var i operation.IOperation = op
		ops = append(ops, &i)
	}
	return ops, nil
}

func (r *OperationRepo) SliceByAccountAndPeriod(ctx context.Context, _ service.ObjectID, from time.Time, to time.Time) ([]*operation.IOperation, error) {
	ops := make([]*operation.IOperation, 0)
	for _, op := range r.repo {
		d := op.Date()
		if (d.Equal(from) || d.After(from)) && (d.Equal(to) || d.Before(to)) {
			var i operation.IOperation = op
			ops = append(ops, &i)
		}
	}
	return ops, nil
}
