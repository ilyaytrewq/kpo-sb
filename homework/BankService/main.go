package main

import (
	"context"
	"fmt"

	bankaccountrepo "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository/BankAccountRepo"
	bankaccount "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/BankAccount"

	csvimporter "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Importer/CsvImporter"
)

func main() {
	fmt.Println("Starting Bank Service...")
	acc, err := bankaccount.NewBankAccount("first", 0)
	if err != nil {
		fmt.Println("failed to create account")
		return
	}
	bankrepo := bankaccountrepo.NewBankAccountRepo()
	var i bankaccount.IBankAccount = acc
	bankrepo.Save(context.Background(), i)
	all, err := bankrepo.All(context.Background())
	if err != nil {
		fmt.Println("cant get all")
	}
	csvimporter.NewBankAccountImporter()
	for _, el := range all {
		fmt.Print(el)
	}
}