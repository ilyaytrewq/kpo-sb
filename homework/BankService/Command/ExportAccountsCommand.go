package command

import (
	exporterCsv "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Exporter/CsvExporter"
	exporterJson "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Exporter/JsonExporter"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
)

type ExportAccountsCommand struct {
	Data     []service.ICommonObject
	Filepath string
	Format   string // "csv" or "json"
}

func (c *ExportAccountsCommand) Execute() error {
	switch c.Format {
	case "csv":
		return exporterCsv.NewCSVBankAccountExporter(c.Filepath).Export(c.Data)
	case "json":
		return exporterJson.NewJSONBankAccountExporter(c.Filepath).Export(c.Data)
	default:
		return nil
	}
}
