package jsonimporter

import (
	"encoding/json"
	"fmt"
	"strings"

	importer "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Importer"
	categoryrepo "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository/CategoryRepo"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	category "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/Category"
	"github.com/google/uuid"
)

type categoryJSON struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type int    `json:"type"`
}

type jsonCategoryParser struct{}

func (p *jsonCategoryParser) Parse(data []byte) ([]service.ICommonObject, error) {
	var cats []categoryJSON
	if err := json.Unmarshal(data, &cats); err != nil {
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

func NewJSONCategoryImporter(filepath string) *importer.BaseImporter {
	return importer.NewImporter(filepath, categoryrepo.NewCategoryRepo(), &jsonCategoryParser{})
}
