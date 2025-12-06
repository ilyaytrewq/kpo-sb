package csvexporter

import (
	"bytes"
	"encoding/csv"
	"fmt"

	"github.com/google/uuid"
	exporter "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Exporter"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	category "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/Category"
)

type csvCategoryFormatter struct{}

func (f *csvCategoryFormatter) FormatData(data interface{}) ([]byte, error) {
	objs, ok := data.([]service.ICommonObject)
	if !ok {
		return nil, fmt.Errorf("invalid data type: expected []service.ICommonObject")
	}
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	if err := w.Write([]string{"id", "name", "type"}); err != nil {
		return nil, err
	}
	for _, o := range objs {
		c, ok := o.(category.ICategory)
		if !ok {
			continue
		}
		if err := w.Write([]string{
			uuid.UUID(c.ID()).String(),
			c.Name(),
			fmt.Sprintf("%d", int(c.Type())),
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

func NewCSVCategoryExporter(filepath string) *exporter.BaseExporter {
	return exporter.NewExporter(filepath, &csvCategoryFormatter{})
}
