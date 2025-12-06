package yamlexporter

import (
	yaml "gopkg.in/yaml.v3"

	"github.com/google/uuid"
	exporter "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Exporter"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	bankaccount "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/BankAccount"
)

type yamlBankAccountFormatter struct{}

func (f *yamlBankAccountFormatter) FormatData(data interface{}) ([]byte, error) {
	objs, ok := data.([]service.ICommonObject)
	if !ok {
		return yaml.Marshal([]any{})
	}
	type out struct {
		ID      string  `yaml:"id"`
		Name    string  `yaml:"name"`
		Balance float64 `yaml:"balance"`
	}
	res := make([]out, 0, len(objs))
	for _, o := range objs {
		acc, ok := o.(bankaccount.IBankAccount)
		if !ok {
			continue
		}
		res = append(res, out{ID: uuid.UUID(acc.ID()).String(), Name: acc.Name(), Balance: acc.Balance()})
	}
	return yaml.Marshal(res)
}

func NewYAMLBankAccountExporter(filepath string) *exporter.BaseExporter {
	return exporter.NewExporter(filepath, &yamlBankAccountFormatter{})
}
