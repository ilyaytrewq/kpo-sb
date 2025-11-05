package csvimporter

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	importer "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Importer"
	operationrepo "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository/OperationRepo"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	operation "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/Operation"
)

type csvOperationParser struct{}

func (p *csvOperationParser) Parse(data []byte) ([]service.ICommonObject, error) {
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
		if len(rec) < 7 {
			errs = append(errs, fmt.Sprintf("row %d: expected 7 columns, got %d", i, len(rec)))
			continue
		}
		id, err := uuid.Parse(rec[0])
		if err != nil {
			errs = append(errs, fmt.Sprintf("row %d: invalid id '%s'", i, rec[0]))
			continue
		}
		t, err := strconv.Atoi(rec[1])
		if err != nil {
			errs = append(errs, fmt.Sprintf("row %d: invalid type '%s'", i, rec[1]))
			continue
		}
		bankAccID, err := uuid.Parse(rec[2])
		if err != nil {
			errs = append(errs, fmt.Sprintf("row %d: invalid bank_account_id '%s'", i, rec[2]))
			continue
		}
		amount, err := strconv.ParseFloat(rec[3], 64)
		if err != nil {
			errs = append(errs, fmt.Sprintf("row %d: invalid amount '%s'", i, rec[3]))
			continue
		}
		date, err := time.Parse(time.RFC3339, rec[4])
		if err != nil {
			errs = append(errs, fmt.Sprintf("row %d: invalid date '%s' (expected RFC3339)", i, rec[4]))
			continue
		}
		descr := rec[5]
		catID, err := uuid.Parse(rec[6])
		if err != nil {
			errs = append(errs, fmt.Sprintf("row %d: invalid category_id '%s'", i, rec[6]))
			continue
		}
		obj, err := operation.NewCopyOperation(
			service.ObjectID(id),
			operation.OperationType(t),
			service.ObjectID(bankAccID),
			amount,
			date,
			service.ObjectID(catID),
			descr,
		)
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

func NewCSVOperationImporter(filepath string) *importer.BaseImporter {
	return importer.NewImporter(filepath, operationrepo.NewOperationRepo(), &csvOperationParser{})
}
