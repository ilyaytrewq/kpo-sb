package service

import (
	"fmt"

	"database/sql/driver"
	"github.com/google/uuid"
)

type ObjectID uuid.UUID

type ICommonObject interface {
	ID() ObjectID
}

func (id ObjectID) String() string { return uuid.UUID(id).String() }

func (id ObjectID) Value() (driver.Value, error) {
    return id.String(), nil
}

func (id *ObjectID) Scan(src any) error {
    switch v := src.(type) {
    case []byte:
        u, err := uuid.ParseBytes(v)
        if err != nil {
            return err
        }
        *id = ObjectID(u)
        return nil
    case string:
        u, err := uuid.Parse(v)
        if err != nil {
            return err
        }
        *id = ObjectID(u)
        return nil
    default:
        return fmt.Errorf("unsupported type for ObjectID: %T", src)
    }
}