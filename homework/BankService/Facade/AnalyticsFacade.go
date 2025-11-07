package facade

import (
	"context"
	"time"

	operationrepo "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository/OperationRepo"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	category "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/Category"
	operation "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/Operation"
)

type AnalyticsFacade struct {
	ops *operationrepo.OperationRepo
}

func NewAnalyticsFacade(ops *operationrepo.OperationRepo) *AnalyticsFacade {
	return &AnalyticsFacade{ops: ops}
}

func (a *AnalyticsFacade) IncomeExpenseDelta(accountID service.ObjectID, from, to time.Time) (float64, float64, float64, error) {
	objs, err := a.ops.SliceByAccountAndPeriod(context.Background(), accountID, from, to)
	if err != nil {
		return 0, 0, 0, err
	}
	var income, expense float64
	for _, obj := range objs {
		op := obj.(operation.IOperation)
		if op.Type() == operation.Income {
			income += op.Amount()
		} else {
			expense += op.Amount()
		}
	}
	return income, expense, income - expense, nil
}

func (a *AnalyticsFacade) GroupByCategory(accountID service.ObjectID, from, to time.Time) (map[service.ObjectID]float64, error) {
	objs, err := a.ops.SliceByAccountAndPeriod(context.Background(), accountID, from, to)
	if err != nil {
		return nil, err
	}
	res := make(map[service.ObjectID]float64)
	for _, obj := range objs {
		op := obj.(operation.IOperation)
		res[op.CategoryID()] += op.Amount()
	}
	return res, nil
}

func (a *AnalyticsFacade) SplitByCategoryType(accountID service.ObjectID, from, to time.Time, categories map[service.ObjectID]category.CategoryType) (map[category.CategoryType]float64, error) {
	objs, err := a.ops.SliceByAccountAndPeriod(context.Background(), accountID, from, to)
	if err != nil {
		return nil, err
	}
	res := map[category.CategoryType]float64{
		category.Spending: 0,
		category.Income:   0,
	}
	for _, obj := range objs {
		op := obj.(operation.IOperation)
		ctype, ok := categories[op.CategoryID()]
		if !ok {
			continue
		}
		res[ctype] += op.Amount()
	}
	return res, nil
}
