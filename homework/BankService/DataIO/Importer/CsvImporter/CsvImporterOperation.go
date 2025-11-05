package csvimporter

import (
	"context"
	"encoding/csv"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	repository "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository"
	operationrepo "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository/OperationRepo"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	operation "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/Operation"
)

type CSVOperationImporter struct {
	repo     *operationrepo.OperationRepo
	filepath string
}

func NewCSVOperationImporter(filepath string) *CSVOperationImporter {
	return &CSVOperationImporter{repo: operationrepo.NewOperationRepo(), filepath: filepath}
}

func (o *CSVOperationImporter) Read() error {
	data, err := os.ReadFile(o.filepath)
	if err != nil {
		return err
	}
	reader := csv.NewReader(strings.NewReader(string(data)))
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}
	for i, rec := range records {
		if i == 0 || len(rec) < 7 {
			continue
		}
		id, err := uuid.Parse(rec[0])
		if err != nil {
			continue
		}
		opType, err := strconv.Atoi(rec[1])
		if err != nil {
			continue
		}
		bankAccID, err := uuid.Parse(rec[2])
		if err != nil {
			continue
		}
		amount, err := strconv.ParseFloat(rec[3], 64)
		if err != nil {
			continue
		}
		date, err := time.Parse(time.RFC3339, rec[4])
		if err != nil {
			continue
		}
		catID, err := uuid.Parse(rec[6])
		if err != nil {
			continue
		}
		obj, err := operation.NewCopyOperation(
			service.ObjectID(id),
			operation.OperationType(opType),
			service.ObjectID(bankAccID),
			amount,
			date,
			service.ObjectID(catID),
			rec[5],
		)
		if err != nil {
			continue
		}
		_ = o.repo.Save(context.Background(), obj)
	}
	return nil
}

func (o *CSVOperationImporter) Data() repository.ICommonRepo { return o.repo }
