package importer

import (
	"context"
	"fmt"
	"os"
	"strings"

	repository "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
)

// Importer is the high-level interface used by callers.
type Importer interface {
	Read() error
	Data() repository.ICommonRepo
}

// DataParser is the varying part: parse raw file bytes into domain objects.
type DataParser interface {
	Parse(data []byte) ([]service.ICommonObject, error)
}

// BaseImporter implements the Template Method: read file -> parse -> save -> aggregate errors.
type BaseImporter struct {
	filepath string
	repo     repository.ICommonRepo
	parser   DataParser
}

// NewImporter constructs a template importer with a concrete repo and parser.
func NewImporter(filepath string, repo repository.ICommonRepo, parser DataParser) *BaseImporter {
	return &BaseImporter{filepath: filepath, repo: repo, parser: parser}
}

// Read performs the template workflow.
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

// Data returns the underlying repository with imported data.
func (b *BaseImporter) Data() repository.ICommonRepo { return b.repo }
