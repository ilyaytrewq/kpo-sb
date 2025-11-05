package csvimporter

import (
	"context"
	"encoding/csv"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	repository "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository"
	bankaccountrepo "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository/BankAccountRepo"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	bankaccount "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/BankAccount"
)

type CSVBankAccountImporter struct {
	repo     *bankaccountrepo.BankAccountRepo
	filepath string
}

func NewCSVBankAccountImporter(filepath string) *CSVBankAccountImporter {
	return &CSVBankAccountImporter{repo: bankaccountrepo.NewBankAccountRepo(), filepath: filepath}
}

func (b *CSVBankAccountImporter) Read() error {
	data, err := os.ReadFile(b.filepath)
	if err != nil {
		return err
	}
	reader := csv.NewReader(strings.NewReader(string(data)))
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}
	for i, rec := range records {
		if i == 0 || len(rec) < 3 { // пропускаем хедер
			continue
		}
		// rec[0]=id, rec[1]=name, rec[2]=balance
		id, err := uuid.Parse(rec[0])
		if err != nil {
			continue
		}
		balance, err := strconv.ParseFloat(rec[2], 64)
		if err != nil {
			continue
		}
		acc, err := bankaccount.NewCopyBankAccount(service.ObjectID(id), rec[1], balance)
		if err != nil {
			continue
		}
		_ = b.repo.Save(context.Background(), acc)
	}
	return nil
}

func (b *CSVBankAccountImporter) Data() repository.ICommonRepo { return b.repo }
