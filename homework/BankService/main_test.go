package main

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	commandpkg "github.com/ilyaytrewq/kpo-sb/homework/BankService/Command"
	csvexporter "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Exporter/CsvExporter"
	jsonexporter "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Exporter/JsonExporter"
	csvimporter "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Importer/CsvImporter"
	jsonimporter "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Importer/JsonImporter"
	facade "github.com/ilyaytrewq/kpo-sb/homework/BankService/Facade"
	bankaccountrepo "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository/BankAccountRepo"
	categoryrepo "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository/CategoryRepo"
	operationrepo "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository/OperationRepo"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	bankaccount "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/BankAccount"
	category "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/Category"
	operation "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/Operation"
	timer "github.com/ilyaytrewq/kpo-sb/homework/BankService/Timer"
)

// ---------- Domain factories validation ----------
func TestFactoriesValidation(t *testing.T) {
	if _, err := bankaccount.NewBankAccount("", 10); err == nil {
		t.Errorf("expected error for empty account name")
	}
	if _, err := bankaccount.NewBankAccount("A", -1); err == nil {
		t.Errorf("expected error for negative balance")
	}
	if _, err := category.NewCategory("", category.Spending); err == nil {
		t.Errorf("expected error for empty category name")
	}
	if _, err := category.NewCategory("Food", category.CategoryType(99)); err == nil {
		t.Errorf("expected error for invalid category type")
	}
	if _, err := operation.NewOperation(operation.Spending, service.ObjectID(uuid.New()), -1, time.Now(), service.ObjectID(uuid.New())); err == nil {
		t.Errorf("expected error for negative amount")
	}
}

// ---------- OperationRepo filtering + Analytics ----------
func TestOperationRepoSliceAndAnalytics(t *testing.T) {
	opRepo := operationrepo.NewOperationRepo()
	accID := service.ObjectID(uuid.New())
	catID := service.ObjectID(uuid.New())
	now := time.Now()
	// operations inside period
	op1, _ := operation.NewOperation(operation.Income, accID, 100, now.Add(-1*time.Hour), catID)
	op2, _ := operation.NewOperation(operation.Spending, accID, 40, now.Add(-30*time.Minute), catID)
	// outside period (before)
	op3, _ := operation.NewOperation(operation.Income, accID, 55, now.Add(-10*time.Hour), catID)
	_ = opRepo.Save(context.Background(), op1)
	_ = opRepo.Save(context.Background(), op2)
	_ = opRepo.Save(context.Background(), op3)

	from := now.Add(-2 * time.Hour)
	to := now
	slice, err := opRepo.SliceByAccountAndPeriod(context.Background(), accID, from, to)
	if err != nil {
		t.Fatalf("slice error: %v", err)
	}
	if len(slice) != 2 {
		t.Fatalf("expected 2 ops in period, got %d", len(slice))
	}

	analytics := facade.NewAnalyticsFacade(opRepo)
	inc, exp, delta, err := analytics.IncomeExpenseDelta(accID, from, to)
	if err != nil {
		t.Fatalf("analytics error: %v", err)
	}
	if inc != 100 || exp != 40 || delta != 60 {
		t.Fatalf("unexpected analytics values inc=%v exp=%v delta=%v", inc, exp, delta)
	}
}

// ---------- Exporter & Importer roundtrip (JSON) ----------
func TestJSONExportImportAccounts(t *testing.T) {
	repo := bankaccountrepo.NewBankAccountRepo()
	acc, _ := bankaccount.NewBankAccount("Main", 123.45)
	if err := repo.Save(context.Background(), acc); err != nil {
		t.Fatalf("save err: %v", err)
	}
	data, _ := repo.All(context.Background())

	tmpFile, err := os.CreateTemp(t.TempDir(), "acc*.json")
	if err != nil {
		t.Fatalf("temp file err: %v", err)
	}
	path := tmpFile.Name()
	tmpFile.Close()

	if err := jsonexporter.NewJSONBankAccountExporter(path).Export(data); err != nil {
		t.Fatalf("export err: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read exported err: %v", err)
	}
	if !strings.Contains(string(raw), "\"balance\":123.45") {
		t.Fatalf("export content mismatch: %s", string(raw))
	}

	// import back
	imp := jsonimporter.NewJSONBankAccountImporter(path)
	if err := imp.Read(); err != nil {
		t.Fatalf("import read err: %v", err)
	}
	imported, _ := imp.Data().All(context.Background())
	if len(imported) != 1 {
		t.Fatalf("expected 1 imported account, got %d", len(imported))
	}
}

// ---------- CSV exporter/importer basic test ----------
func TestCSVExportImportCategories(t *testing.T) {
	repo := categoryrepo.NewCategoryRepo()
	cat, _ := category.NewCategory("Food", category.Spending)
	_ = repo.Save(context.Background(), cat)
	data, _ := repo.All(context.Background())
	tmpFile, _ := os.CreateTemp(t.TempDir(), "cat*.csv")
	path := tmpFile.Name()
	tmpFile.Close()
	if err := csvexporter.NewCSVCategoryExporter(path).Export(data); err != nil {
		t.Fatalf("csv export err: %v", err)
	}
	imp := csvimporter.NewCSVCategoryImporter(path)
	if err := imp.Read(); err != nil {
		t.Fatalf("csv import err: %v", err)
	}
	imported, _ := imp.Data().All(context.Background())
	if len(imported) != 1 {
		t.Fatalf("expected 1 imported category, got %d", len(imported))
	}
}

// ---------- Timer decorator test ----------
func TestTimerDecorator(t *testing.T) {
	bankRepo := bankaccountrepo.NewBankAccountRepo()
	fac := facade.NewBankAccountFacade(bankRepo)
	cmd := &commandpkg.CreateAccountCommand{Facade: fac, Name: "X", Balance: 10}
	timed := timer.NewTimerDecorator(cmd)
	start := time.Now()
	if err := timed.Execute(); err != nil {
		t.Fatalf("timed execute error: %v", err)
	}
	if time.Since(start) <= 0 {
		t.Fatalf("expected positive duration")
	}
	all, _ := bankRepo.All(context.Background())
	if len(all) != 1 {
		t.Fatalf("expected created account in repo")
	}
}
