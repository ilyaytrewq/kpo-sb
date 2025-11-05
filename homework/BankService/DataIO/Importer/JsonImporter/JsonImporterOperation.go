package jsonimporter

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	importer "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Importer"
	operationrepo "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository/OperationRepo"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	operation "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/Operation"
)

type operationJSON struct {
	ID            string  `json:"id"`
	Type          int     `json:"type"` // 0 - Spending, 1 - Income
	BankAccountID string  `json:"bank_account_id"`
	Amount        float64 `json:"amount"`
	Date          string  `json:"date"` // RFC3339 format
	Description   string  `json:"description"`
	CategoryID    string  `json:"category_id"`
}

type jsonOperationParser struct{}

func (p *jsonOperationParser) Parse(data []byte) ([]service.ICommonObject, error) {
	var ops []operationJSON
	if err := json.Unmarshal(data, &ops); err != nil {
		return nil, err
	}
	var (
		result []service.ICommonObject
		errs   []string
	)
	for _, op := range ops {
		id, err := uuid.Parse(op.ID)
		if err != nil {
			errs = append(errs, fmt.Sprintf("invalid id '%s'", op.ID))
			continue
		}
		bankID, err := uuid.Parse(op.BankAccountID)
		if err != nil {
			errs = append(errs, fmt.Sprintf("invalid bank_account_id '%s'", op.BankAccountID))
			continue
		}
		catID, err := uuid.Parse(op.CategoryID)
		if err != nil {
			errs = append(errs, fmt.Sprintf("invalid category_id '%s'", op.CategoryID))
			continue
		}
		dt, err := time.Parse(time.RFC3339, op.Date)
		if err != nil {
			errs = append(errs, fmt.Sprintf("invalid date '%s'", op.Date))
			continue
		}
		el, err := operation.NewCopyOperation(
			service.ObjectID(id),
			operation.OperationType(op.Type),
			service.ObjectID(bankID),
			op.Amount,
			dt,
			service.ObjectID(catID),
			op.Description,
		)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		result = append(result, el)
	}
	if len(errs) > 0 {
		return result, fmt.Errorf("parse finished with %d errors: %s", len(errs), strings.Join(errs, "; "))
	}
	return result, nil
}

func NewJSONOperationImporter(filepath string) *importer.BaseImporter {
	return importer.NewImporter(filepath, operationrepo.NewOperationRepo(), &jsonOperationParser{})
}
