package jsonexporter

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	exporter "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Exporter"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	category "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/Category"
)

type jsonCategoryFormatter struct{}

func (f *jsonCategoryFormatter) FormatData(data interface{}) ([]byte, error) {
	objs, ok := data.([]service.ICommonObject)
	if !ok {
		return nil, fmt.Errorf("invalid data type: expected []service.ICommonObject")
	}
	type out struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type int    `json:"type"`
	}
	res := make([]out, 0, len(objs))
	for _, o := range objs {
		c, ok := o.(category.ICategory)
		if !ok {
			continue
		}
		res = append(res, out{
			ID:   uuid.UUID(c.ID()).String(),
			Name: c.Name(),
			Type: int(c.Type()),
		})
	}
	return json.MarshalIndent(res, "", "\t")
}

func NewJSONCategoryExporter(filepath string) *exporter.BaseExporter {
	return exporter.NewExporter(filepath, &jsonCategoryFormatter{})
}
