package yamlimporter

import (
	"fmt"
	"strings"

	yaml "gopkg.in/yaml.v3"

	"github.com/google/uuid"
	importer "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Importer"
	categoryrepo "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository/CategoryRepo"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	category "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/Category"
)

type categoryYAML struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
	Type int    `yaml:"type"`
}

type yamlCategoryParser struct{}

func (p *yamlCategoryParser) Parse(data []byte) ([]service.ICommonObject, error) {
	var cats []categoryYAML
	if err := yaml.Unmarshal(data, &cats); err != nil {
		return nil, err
	}
	var (
		result []service.ICommonObject
		errs   []string
	)
	for _, c := range cats {
		id, err := uuid.Parse(c.ID)
		if err != nil {
			errs = append(errs, fmt.Sprintf("invalid id '%s'", c.ID))
			continue
		}
		el, err := category.NewCopyCategory(service.ObjectID(id), c.Name, category.CategoryType(c.Type))
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

func NewYAMLCategoryImporter(filepath string) *importer.BaseImporter {
	return importer.NewImporter(filepath, categoryrepo.NewCategoryRepo(), &yamlCategoryParser{})
}
