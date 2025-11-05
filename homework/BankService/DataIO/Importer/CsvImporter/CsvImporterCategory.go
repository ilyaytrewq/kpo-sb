package csvimporter

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	importer "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Importer"
	categoryrepo "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository/CategoryRepo"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	category "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/Category"
)

type csvCategoryParser struct{}

func (p *csvCategoryParser) Parse(data []byte) ([]service.ICommonObject, error) {
	reader := csv.NewReader(strings.NewReader(string(data)))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	var (
		result []service.ICommonObject
		errs   []string
	)
	for i, rec := range records {
		if i == 0 { // header
			continue
		}
		if len(rec) < 3 {
			errs = append(errs, fmt.Sprintf("row %d: expected 3 columns, got %d", i, len(rec)))
			continue
		}
		id, err := uuid.Parse(rec[0])
		if err != nil {
			errs = append(errs, fmt.Sprintf("row %d: invalid id '%s'", i, rec[0]))
			continue
		}
		t, err := strconv.Atoi(rec[2])
		if err != nil {
			errs = append(errs, fmt.Sprintf("row %d: invalid type '%s'", i, rec[2]))
			continue
		}
		obj, err := category.NewCopyCategory(service.ObjectID(id), rec[1], category.CategoryType(t))
		if err != nil {
			errs = append(errs, fmt.Sprintf("row %d: %v", i, err))
			continue
		}
		result = append(result, obj)
	}
	if len(errs) > 0 {
		return result, fmt.Errorf("parse finished with %d errors: %s", len(errs), strings.Join(errs, "; "))
	}
	return result, nil
}

func NewCSVCategoryImporter(filepath string) *importer.BaseImporter {
	return importer.NewImporter(filepath, categoryrepo.NewCategoryRepo(), &csvCategoryParser{})
}
