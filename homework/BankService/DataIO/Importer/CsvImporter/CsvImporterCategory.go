package csvimporter

import (
	"context"
	"encoding/csv"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	repository "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository"
	categoryrepo "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository/CategoryRepo"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	category "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/Category"
)

type CSVCategoryImporter struct {
	repo     *categoryrepo.CategoryRepo
	filepath string
}

func NewCSVCategoryImporter(filepath string) *CSVCategoryImporter {
	return &CSVCategoryImporter{repo: categoryrepo.NewCategoryRepo(), filepath: filepath}
}

func (c *CSVCategoryImporter) Read() error {
	data, err := os.ReadFile(c.filepath)
	if err != nil {
		return err
	}
	reader := csv.NewReader(strings.NewReader(string(data)))
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}
	for i, rec := range records {
		if i == 0 || len(rec) < 3 {
			continue
		}
		id, err := uuid.Parse(rec[0])
		if err != nil {
			continue
		}
		catType, err := strconv.Atoi(rec[2])
		if err != nil {
			continue
		}
		obj, err := category.NewCopyCategory(service.ObjectID(id), rec[1], category.CategoryType(catType))
		if err != nil {
			continue
		}
		_ = c.repo.Save(context.Background(), obj)
	}
	return nil
}

func (c *CSVCategoryImporter) Data() repository.ICommonRepo { return c.repo }
