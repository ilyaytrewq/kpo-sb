package csvimporter

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	importer "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Importer"
	bankaccountrepo "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository/BankAccountRepo"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	bankaccount "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/BankAccount"
)

type csvBankParser struct{}

func (p *csvBankParser) Parse(data []byte) ([]service.ICommonObject, error) {
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
		balance, err := strconv.ParseFloat(rec[2], 64)
		if err != nil {
			errs = append(errs, fmt.Sprintf("row %d: invalid balance '%s'", i, rec[2]))
			continue
		}
		acc, err := bankaccount.NewCopyBankAccount(service.ObjectID(id), rec[1], balance)
		if err != nil {
			errs = append(errs, fmt.Sprintf("row %d: %v", i, err))
			continue
		}
		result = append(result, acc)
	}
	if len(errs) > 0 {
		return result, fmt.Errorf("parse finished with %d errors: %s", len(errs), strings.Join(errs, "; "))
	}
	return result, nil
}

func NewCSVBankAccountImporter(filepath string) *importer.BaseImporter {
	return importer.NewImporter(filepath, bankaccountrepo.NewBankAccountRepo(), &csvBankParser{})
}
