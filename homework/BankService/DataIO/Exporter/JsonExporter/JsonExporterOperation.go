package jsonexporter

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	exporter "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Exporter"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	operation "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/Operation"
)

type jsonOperationFormatter struct{}

func (f *jsonOperationFormatter) FormatData(data interface{}) ([]byte, error) {
	objs, ok := data.([]service.ICommonObject)
	if !ok {
		return nil, fmt.Errorf("invalid data type: expected []service.ICommonObject")
	}
	type out struct {
		ID            string  `json:"id"`
		Type          int     `json:"type"`
		BankAccountID string  `json:"bank_account_id"`
		Amount        float64 `json:"amount"`
		Date          string  `json:"date"`
		Description   string  `json:"description"`
		CategoryID    string  `json:"category_id"`
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
	return json.Marshal(res)
}

func NewJSONOperationExporter(filepath string) *exporter.BaseExporter {
	return exporter.NewExporter(filepath, &jsonOperationFormatter{})
}
