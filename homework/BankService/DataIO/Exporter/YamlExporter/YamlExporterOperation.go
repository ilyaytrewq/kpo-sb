package yamlexporter

import (
	"time"

	yaml "gopkg.in/yaml.v3"

	"github.com/google/uuid"
	exporter "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Exporter"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	operation "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/Operation"
)

type yamlOperationFormatter struct{}

func (f *yamlOperationFormatter) FormatData(data interface{}) ([]byte, error) {
	objs, _ := data.([]service.ICommonObject)
	type out struct {
		ID            string  `yaml:"id"`
		Type          int     `yaml:"type"`
		BankAccountID string  `yaml:"bank_account_id"`
		Amount        float64 `yaml:"amount"`
		Date          string  `yaml:"date"`
		Description   string  `yaml:"description"`
		CategoryID    string  `yaml:"category_id"`
	}
	res := make([]out, 0, len(objs))
	for _, o := range objs {
		op, ok := o.(operation.IOperation)
		if !ok {
			continue
		}
		res = append(res, out{
			ID:            uuid.UUID(op.ID()).String(),
			Type:          int(op.Type()),
			BankAccountID: uuid.UUID(op.BankAccountID()).String(),
			Amount:        op.Amount(),
			Date:          op.Date().Format(time.RFC3339),
			Description:   op.Description(),
			CategoryID:    uuid.UUID(op.CategoryID()).String(),
		})
	}
	return yaml.Marshal(res)
}

func NewYAMLOperationExporter(filepath string) *exporter.BaseExporter {
	return exporter.NewExporter(filepath, &yamlOperationFormatter{})
}
