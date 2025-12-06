package csvexporter

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	exporter "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Exporter"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	bankaccount "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/BankAccount"
)

type csvBankAccountFormatter struct{}

func (f *csvBankAccountFormatter) FormatData(data interface{}) ([]byte, error) {
	objs, ok := data.([]service.ICommonObject)
	if !ok {
		return nil, fmt.Errorf("invalid data type: expected []service.ICommonObject")
	}
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	if err := w.Write([]string{"id", "name", "balance"}); err != nil {
		return nil, err
	}
	for _, o := range objs {
		acc, ok := o.(bankaccount.IBankAccount)
		if !ok {
			continue
		}
		if err := w.Write([]string{
			uuid.UUID(acc.ID()).String(),
			acc.Name(),
			strconv.FormatFloat(acc.Balance(), 'f', -1, 64),
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

func NewCSVBankAccountExporter(filepath string) *exporter.BaseExporter {
	return exporter.NewExporter(filepath, &csvBankAccountFormatter{})
}
