package exporter

import (
	"fmt"
	"os"
)

type DataFormatter interface {
	FormatData(data interface{}) ([]byte, error)
}

type BaseExporter struct {
	filepath  string
	formatter DataFormatter
}

func NewExporter(filepath string, formatter DataFormatter) *BaseExporter {
	return &BaseExporter{filepath: filepath, formatter: formatter}
}

func (e *BaseExporter) Export(data interface{}) error {
	bytes, err := e.formatter.FormatData(data)
	if err != nil {
		return fmt.Errorf("format data: %w", err)
	}
	if err := os.WriteFile(e.filepath, bytes, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}
