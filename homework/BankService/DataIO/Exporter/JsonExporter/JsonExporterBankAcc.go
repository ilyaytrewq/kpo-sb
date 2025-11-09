package jsonexporter

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	exporter "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Exporter"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	bankaccount "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/BankAccount"
)

type jsonBankAccountFormatter struct{}

func (f *jsonBankAccountFormatter) FormatData(data interface{}) ([]byte, error) {
	objs, ok := data.([]service.ICommonObject)
	if !ok {
		return nil, fmt.Errorf("invalid data type: expected []service.ICommonObject")
	}
	type out struct {
		ID      string  `json:"id"`
		Name    string  `json:"name"`
		Balance float64 `json:"balance"`
	}
	res := make([]out, 0, len(objs))
	for _, o := range objs {
		acc, ok := o.(bankaccount.IBankAccount)
		if !ok {
			// skip incompatible object
			continue
		}
		res = append(res, out{
			ID:      uuid.UUID(acc.ID()).String(),
			Name:    acc.Name(),
			Balance: acc.Balance(),
		})
	}
	return json.MarshalIndent(res, "", "\t")
}

func NewJSONBankAccountExporter(filepath string) *exporter.BaseExporter {
	return exporter.NewExporter(filepath, &jsonBankAccountFormatter{})
}
