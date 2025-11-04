package repository

import (
	"context"
	"errors"
	"time"

	bankaccount "github.com/ilyaytrewq/kpo-sb/homework/BankService/BankAccount"
	operation "github.com/ilyaytrewq/kpo-sb/homework/BankService/Operation"
)

type IOperationRepo interface {
	ByID(ctx context.Context, id operation.OperationID) (*operation.IOperation, error)
	Save(ctx context.Context, op operation.IOperation) error
	All(ctx context.Context) ([]*operation.IOperation, error)
	SliceByAccountAndPeriod(ctx context.Context, id bankaccount.BankAccountID, from time.Time, to time.Time) ([]*operation.IOperation, error)
}

type OperationRepo struct {
	repo map[operation.OperationID]*operation.Operation
}

func (r *OperationRepo) ByID(ctx context.Context, id operation.OperationID) (*operation.IOperation, error) {
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

func (r *OperationRepo) SliceByAccountAndPeriod(ctx context.Context, _ bankaccount.BankAccountID, from time.Time, to time.Time) ([]*operation.IOperation, error) {
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