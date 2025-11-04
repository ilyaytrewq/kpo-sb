package main

import (
	"fmt"
	"context"

	bankaccount "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/BankAccount"
	bankaccountrepo "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository/BankAccountRepo"

)

func main() {
	fmt.Println("Starting Bank Service...")
	acc, err := bankaccount.NewBankAccount("first", 0)
	if err != nil {
		fmt.Println("failed to create account")
		return
	}
	bankrepo := bankaccountrepo.NewBankAccount()
	var i bankaccount.IBankAccount = acc
	bankrepo.Save(context.Background(), &i)
	all, err := bankrepo.All(context.Background())
	if err != nil {
		fmt.Println("cant get all")
	}
	for _, el := range all {
		fmt.Print(*el)
	}
}