package jsonimporter

import (
	"encoding/json"
	"fmt"
	"strings"

	importer "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Importer"
	bankaccountrepo "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository/BankAccountRepo"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	bankaccount "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/BankAccount"
	"github.com/google/uuid"
)

type bankAccountJSON struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Balance float64 `json:"balance"`
}

// jsonBankParser implements DataParser for bank accounts in JSON
type jsonBankParser struct{}

func (p *jsonBankParser) Parse(data []byte) ([]service.ICommonObject, error) {
	var accounts []bankAccountJSON
	if err := json.Unmarshal(data, &accounts); err != nil {
		return nil, err
	}
	var (
		result []service.ICommonObject
		errs   []string
	)
	for _, acc := range accounts {
		id, err := uuid.Parse(acc.ID)
		if err != nil {
			errs = append(errs, fmt.Sprintf("invalid id '%s'", acc.ID))
			continue
		}
		el, err := bankaccount.NewCopyBankAccount(service.ObjectID(id), acc.Name, acc.Balance)
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

// NewJSONBankAccountImporter returns a template importer configured for JSON bank accounts
func NewJSONBankAccountImporter(filepath string) *importer.BaseImporter {
	return importer.NewImporter(filepath, bankaccountrepo.NewBankAccountRepo(), &jsonBankParser{})
}
