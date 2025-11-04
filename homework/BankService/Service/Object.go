package service

import "github.com/google/uuid"


type ObjectID uuid.UUID

type ICommonObject interface {
	ID() ObjectID
}