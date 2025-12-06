package yamlimporter

import (
	"fmt"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v3"

	"github.com/google/uuid"
	importer "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Importer"
	operationrepo "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository/OperationRepo"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	operation "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/Operation"
)

type operationYAML struct {
	ID            string  `yaml:"id"`
	Type          int     `yaml:"type"`
	BankAccountID string  `yaml:"bank_account_id"`
	Amount        float64 `yaml:"amount"`
	Date          string  `yaml:"date"`
	Description   string  `yaml:"description"`
	CategoryID    string  `yaml:"category_id"`
}

type yamlOperationParser struct{}

func (p *yamlOperationParser) Parse(data []byte) ([]service.ICommonObject, error) {
	var ops []operationYAML
	if err := yaml.Unmarshal(data, &ops); err != nil {
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
		el, err := operation.NewCopyOperation(service.ObjectID(id), operation.OperationType(op.Type), service.ObjectID(bankID), op.Amount, dt, service.ObjectID(catID), op.Description)
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

func NewYAMLOperationImporter(filepath string) *importer.BaseImporter {
	return importer.NewImporter(filepath, operationrepo.NewOperationRepo(), &yamlOperationParser{})
}
