package yamlimporter

import (
	"fmt"
	"strings"

	yaml "gopkg.in/yaml.v3"

	"github.com/google/uuid"
	importer "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Importer"
	bankaccountrepo "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository/BankAccountRepo"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	bankaccount "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/BankAccount"
)

type bankAccountYAML struct {
	ID      string  `yaml:"id"`
	Name    string  `yaml:"name"`
	Balance float64 `yaml:"balance"`
}

type yamlBankParser struct{}

func (p *yamlBankParser) Parse(data []byte) ([]service.ICommonObject, error) {
	var accounts []bankAccountYAML
	if err := yaml.Unmarshal(data, &accounts); err != nil {
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

func NewYAMLBankAccountImporter(filepath string) *importer.BaseImporter {
	return importer.NewImporter(filepath, bankaccountrepo.NewBankAccountRepo(), &yamlBankParser{})
}
