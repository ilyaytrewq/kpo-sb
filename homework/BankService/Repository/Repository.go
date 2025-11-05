package repository


import (
	"context"
	
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"

)

type ICommonRepo interface {
		ByID(ctx context.Context, id service.ObjectID) (service.ICommonObject, error)
		Save(ctx context.Context, obj service.ICommonObject) error
		All(ctx context.Context) ([]service.ICommonObject, error)
}