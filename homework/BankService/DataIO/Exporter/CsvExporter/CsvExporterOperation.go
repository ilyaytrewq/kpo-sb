package csvexporter

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	exporter "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Exporter"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	operation "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/Operation"
)

type csvOperationFormatter struct{}

func (f *csvOperationFormatter) FormatData(data interface{}) ([]byte, error) {
	objs, ok := data.([]service.ICommonObject)
	if !ok {
		return nil, fmt.Errorf("invalid data type: expected []service.ICommonObject")
	}
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	if err := w.Write([]string{"id", "type", "bank_account_id", "amount", "date", "description", "category_id"}); err != nil {
		return nil, err
	}
	for _, o := range objs {
		op, ok := o.(operation.IOperation)
		if !ok {
			continue
		}
		if err := w.Write([]string{
			uuid.UUID(op.ID()).String(),
			fmt.Sprintf("%d", int(op.Type())),
			uuid.UUID(op.BankAccountID()).String(),
			strconv.FormatFloat(op.Amount(), 'f', -1, 64),
			op.Date().Format(time.RFC3339),
			op.Description(),
			uuid.UUID(op.CategoryID()).String(),
		}); err != nil {
			return nil, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func NewCSVOperationExporter(filepath string) *exporter.BaseExporter {
	return exporter.NewExporter(filepath, &csvOperationFormatter{})
}
