package importer

import (
	"context"
	"fmt"
	"os"
	"strings"

	repository "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
)

type Importer interface {
	Read() error
	Data() repository.ICommonRepo
}

type DataParser interface {
	Parse(data []byte) ([]service.ICommonObject, error)
}

type BaseImporter struct {
	filepath string
	repo     repository.ICommonRepo
	parser   DataParser
}

func NewImporter(filepath string, repo repository.ICommonRepo, parser DataParser) *BaseImporter {
	return &BaseImporter{filepath: filepath, repo: repo, parser: parser}
}

func (b *BaseImporter) Read() error {
	data, err := os.ReadFile(b.filepath)
	if err != nil {
		return err
	}
	objs, err := b.parser.Parse(data)
	if err != nil {
		return err
	}
	var errs []string
	for _, obj := range objs {
		if err := b.repo.Save(context.Background(), obj); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("import finished with %d errors: %s", len(errs), strings.Join(errs, "; "))
	}
	return nil
}

func (b *BaseImporter) Data() repository.ICommonRepo { return b.repo }
