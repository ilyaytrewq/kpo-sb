package yamlexporter

import (
	yaml "gopkg.in/yaml.v3"

	"github.com/google/uuid"
	exporter "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Exporter"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	category "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/Category"
)

type yamlCategoryFormatter struct{}

func (f *yamlCategoryFormatter) FormatData(data interface{}) ([]byte, error) {
	objs, _ := data.([]service.ICommonObject)
	type out struct {
		ID   string `yaml:"id"`
		Name string `yaml:"name"`
		Type int    `yaml:"type"`
	}
	res := make([]out, 0, len(objs))
	for _, o := range objs {
		c, ok := o.(category.ICategory)
		if !ok {
			continue
		}
		res = append(res, out{ID: uuid.UUID(c.ID()).String(), Name: c.Name(), Type: int(c.Type())})
	}
	return yaml.Marshal(res)
}

func NewYAMLCategoryExporter(filepath string) *exporter.BaseExporter {
	return exporter.NewExporter(filepath, &yamlCategoryFormatter{})
}
