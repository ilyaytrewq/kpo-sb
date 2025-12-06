package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	commandpkg "github.com/ilyaytrewq/kpo-sb/homework/BankService/Command"
	csvexporter "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Exporter/CsvExporter"
	jsonexporter "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Exporter/JsonExporter"
	yamlexporter "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Exporter/YamlExporter"
	csvimporter "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Importer/CsvImporter"
	jsonimporter "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Importer/JsonImporter"
	yamlimporter "github.com/ilyaytrewq/kpo-sb/homework/BankService/DataIO/Importer/YamlImporter"
	facade "github.com/ilyaytrewq/kpo-sb/homework/BankService/Facade"
	repository "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository"
	// bankaccountrepo "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository/BankAccountRepo"
	// categoryrepo "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository/CategoryRepo"
	// operationrepo "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository/OperationRepo"
	dbrepo "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository/DBRepo"
	postgresrepo "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository/DBRepo/PostgresRepo"
	proxyrepo "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository/ProxyRepo"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	category "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/Category"
	operation "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/Operation"
	timer "github.com/ilyaytrewq/kpo-sb/homework/BankService/Timer"
)

func main() {
	in := bufio.NewReader(os.Stdin)

	// bankRepo := bankaccountrepo.NewBankAccountRepo()
	// catRepo := categoryrepo.NewCategoryRepo()
	// opRepo := operationrepo.NewOperationRepo()

	user := getEnv("DB_USER", "bankservice")
	pass := getEnv("DB_PASSWORD", "password")
	name := getEnv("DB_NAME", "bankservice")
	host := getEnv("DB_HOST", "db")
	port := getEnv("DB_PORT", "5432")

	postgreRepo := postgresrepo.NewPostgresRepo()
	if err := postgreRepo.Init(user, pass, name, host, port); err != nil {
		fmt.Println("error initializing postgres repo:", err)
		return
	}
	defer postgreRepo.Close()

	bankRepo := dbrepo.NewBankAccountDBRepo(postgreRepo.DB())
	catRepo := dbrepo.NewCategoryDBRepo(postgreRepo.DB())
	opRepo := dbrepo.NewOperationDBRepo(postgreRepo.DB())

	ctx := context.Background()
	bankCached, err := proxyrepo.NewCachedRepo(ctx, bankRepo)
	if err != nil {
		fmt.Println("warn: bank proxy init failed:", err)
	}
	catCached, err := proxyrepo.NewCachedRepo(ctx, catRepo)
	if err != nil {
		fmt.Println("warn: category proxy init failed:", err)
	}

	var bankFacadeRepo repository.ICommonRepo = bankRepo
	if bankCached != nil {
		bankFacadeRepo = bankCached
	}
	var catFacadeRepo repository.ICommonRepo = catRepo
	if catCached != nil {
		catFacadeRepo = catCached
	}

	bankF := facade.NewBankAccountFacade(bankFacadeRepo)
	catF := facade.NewCategoryFacade(catFacadeRepo)
	opF := facade.NewOperationFacade(opRepo)
	analyticsF := facade.NewAnalyticsFacade(opRepo)

	fmt.Println("Bank Service CLI. Type a number and press Enter.")
	for {
		fmt.Println("\nMenu:")
		fmt.Println(" 1) Create account (timed)")
		fmt.Println(" 2) List accounts")
		fmt.Println(" 3) Delete account")
		fmt.Println(" 4) Create category")
		fmt.Println(" 5) List categories")
		fmt.Println(" 6) Create operation")
		fmt.Println(" 7) List operations")
		fmt.Println(" 8) Export accounts (csv/json/yaml)")
		fmt.Println(" 9) Export categories (csv/json/yaml)")
		fmt.Println("10) Export operations (csv/json/yaml)")
		fmt.Println("11) Analytics: income/expense delta")
		fmt.Println("12) Analytics: group by category")
		fmt.Println("13) Import accounts (csv/json/yaml)")
		fmt.Println("14) Import categories (csv/json/yaml)")
		fmt.Println("15) Import operations (csv/json/yaml)")
		fmt.Println("16) Get account by ID")
		fmt.Println("17) Update account name")
		fmt.Println("18) Update account balance")
		fmt.Println("19) Get category by ID")
		fmt.Println("20) Update category name")
		fmt.Println("21) Delete category")
		fmt.Println("22) Get operation by ID")
		fmt.Println("23) Delete operation")
		fmt.Println(" 0) Exit")
		fmt.Print("> ")
		choice, _ := in.ReadString('\n')
		choice = strings.TrimSpace(choice)
		switch choice {
		case "1":
			name := readString(in, "Account name: ")
			bal := readFloat(in, "Initial balance: ")
			cmd := &commandpkg.CreateAccountCommand{Facade: bankF, Name: name, Balance: bal}
			timed := timer.NewTimerDecorator(cmd)
			if err := timed.Execute(); err != nil {
				fmt.Println("error:", err)
			} else {
				fmt.Println("created account:", uuid.UUID(cmd.CreatedID).String())
			}
		case "2":
			accs, err := bankF.ListAllAccounts()
			if err != nil {
				fmt.Println("error:", err)
				break
			}
			for _, a := range accs {
				fmt.Printf("%s | %s | %.2f\n", uuid.UUID(a.ID()).String(), a.Name(), a.Balance())
			}
		case "3":
			id := readUUID(in, "Account ID (uuid): ")
			if err := bankF.DeleteAccount(service.ObjectID(id)); err != nil {
				fmt.Println("error:", err)
			} else {
				fmt.Println("deleted")
			}
		case "4":
			name := readString(in, "Category name: ")
			t := readInt(in, "Type (0=Spending,1=Income): ")
			ccmd := &commandpkg.CreateCategoryCommand{Facade: catF, Name: name, Type: category.CategoryType(t)}
			if err := ccmd.Execute(); err != nil {
				fmt.Println("error:", err)
			} else {
				fmt.Println("created category:", uuid.UUID(ccmd.CreatedID).String())
			}
		case "5":
			cats, err := catF.ListAllCategories()
			if err != nil {
				fmt.Println("error:", err)
				break
			}
			for _, c := range cats {
				fmt.Printf("%s | %s | %d\n", uuid.UUID(c.ID()).String(), c.Name(), int(c.Type()))
			}
		case "6":
			t := readInt(in, "Type (0=Spending,1=Income): ")
			accID := readUUID(in, "Account ID: ")
			amount := readFloat(in, "Amount: ")
			date := readTime(in, "Date (RFC3339): ")
			catID := readUUID(in, "Category ID: ")
			descr := readString(in, "Description (optional): ")
			ocmd := &commandpkg.AddOperationCommand{
				Facade:      opF,
				Type:        operation.OperationType(t),
				AccountID:   service.ObjectID(accID),
				Amount:      amount,
				Date:        date,
				CategoryID:  service.ObjectID(catID),
				Description: descr,
			}
			if err := ocmd.Execute(); err != nil {
				fmt.Println("error:", err)
			} else {
				fmt.Println("created operation:", uuid.UUID(ocmd.CreatedID).String())
			}
		case "7":
			ops, err := opF.ListAllOperations()
			if err != nil {
				fmt.Println("error:", err)
				break
			}
			for _, o := range ops {
				fmt.Printf("%s | %d | %.2f | %s | %s\n", uuid.UUID(o.ID()).String(), int(o.Type()), o.Amount(), o.Date().Format(time.RFC3339), o.Description())
			}
		case "8":
			format := strings.ToLower(readString(in, "Format (csv/json/yaml): "))
			path := readString(in, "File path: ")
			data, _ := bankFacadeRepo.All(context.Background())
			cmd := commandpkg.CommandFunc(func() error {
				var err error
				switch format {
				case "csv":
					err = csvexporter.NewCSVBankAccountExporter(path).Export(data)
				case "json":
					err = jsonexporter.NewJSONBankAccountExporter(path).Export(data)
				case "yaml":
					err = yamlexporter.NewYAMLBankAccountExporter(path).Export(data)
				default:
					fmt.Println("unknown format")
				}
				if err == nil {
					fmt.Println("exported")
				}
				return err
			})
			if err := timer.NewTimerDecorator(cmd).Execute(); err != nil {
				fmt.Println("error:", err)
			}
		case "9":
			format := strings.ToLower(readString(in, "Format (csv/json/yaml): "))
			path := readString(in, "File path: ")
			data, _ := catFacadeRepo.All(context.Background())
			cmd := commandpkg.CommandFunc(func() error {
				var err error
				switch format {
				case "csv":
					err = csvexporter.NewCSVCategoryExporter(path).Export(data)
				case "json":
					err = jsonexporter.NewJSONCategoryExporter(path).Export(data)
				case "yaml":
					err = yamlexporter.NewYAMLCategoryExporter(path).Export(data)
				default:
					fmt.Println("unknown format")
				}
				if err == nil {
					fmt.Println("exported")
				}
				return err
			})
			if err := timer.NewTimerDecorator(cmd).Execute(); err != nil {
				fmt.Println("error:", err)
			}
		case "10":
			format := strings.ToLower(readString(in, "Format (csv/json/yaml): "))
			path := readString(in, "File path: ")
			data, _ := opRepo.All(context.Background())
			cmd := commandpkg.CommandFunc(func() error {
				var err error
				switch format {
				case "csv":
					err = csvexporter.NewCSVOperationExporter(path).Export(data)
				case "json":
					err = jsonexporter.NewJSONOperationExporter(path).Export(data)
				case "yaml":
					err = yamlexporter.NewYAMLOperationExporter(path).Export(data)
				default:
					fmt.Println("unknown format")
				}
				if err == nil {
					fmt.Println("exported")
				}
				return err
			})
			if err := timer.NewTimerDecorator(cmd).Execute(); err != nil {
				fmt.Println("error:", err)
			}
		case "11":
			accID := readUUID(in, "Account ID: ")
			from := readTime(in, "From (RFC3339): ")
			to := readTime(in, "To (RFC3339): ")
			inc, exp, delta, err := analyticsF.IncomeExpenseDelta(service.ObjectID(accID), from, to)
			if err != nil {
				fmt.Println("error:", err)
			} else {
				fmt.Printf("income=%.2f expense=%.2f delta=%.2f\n", inc, exp, delta)
			}
		case "12":
			accID := readUUID(in, "Account ID: ")
			from := readTime(in, "From (RFC3339): ")
			to := readTime(in, "To (RFC3339): ")
			m, err := analyticsF.GroupByCategory(service.ObjectID(accID), from, to)
			if err != nil {
				fmt.Println("error:", err)
				break
			}
			for k, v := range m {
				fmt.Printf("%s -> %.2f\n", uuid.UUID(k).String(), v)
			}
		case "13":
			format := strings.ToLower(readString(in, "Format (csv/json/yaml): "))
			path := readString(in, "File path: ")
			cmd := commandpkg.CommandFunc(func() error {
				var imp interface {
					Read() error
					Data() repository.ICommonRepo
				}
				switch format {
				case "csv":
					imp = csvimporter.NewCSVBankAccountImporter(path)
				case "json":
					imp = jsonimporter.NewJSONBankAccountImporter(path)
				case "yaml":
					imp = yamlimporter.NewYAMLBankAccountImporter(path)
				}
				if imp == nil {
					fmt.Println("unknown format")
					return nil
				}
				if err := imp.Read(); err != nil {
					return err
				}
				objs, _ := imp.Data().All(context.Background())
				added, failed := 0, 0
				for _, obj := range objs {
					if e := bankFacadeRepo.Save(context.Background(), obj); e != nil {
						failed++
					} else {
						added++
					}
				}
				fmt.Printf("imported: %d, skipped: %d\n", added, failed)
				return nil
			})
			if err := timer.NewTimerDecorator(cmd).Execute(); err != nil {
				fmt.Println("error:", err)
			}
		case "14":
			format := strings.ToLower(readString(in, "Format (csv/json/yaml): "))
			path := readString(in, "File path: ")
			cmd := commandpkg.CommandFunc(func() error {
				var imp interface {
					Read() error
					Data() repository.ICommonRepo
				}
				switch format {
				case "csv":
					imp = csvimporter.NewCSVCategoryImporter(path)
				case "json":
					imp = jsonimporter.NewJSONCategoryImporter(path)
				case "yaml":
					imp = yamlimporter.NewYAMLCategoryImporter(path)
				}
				if imp == nil {
					fmt.Println("unknown format")
					return nil
				}
				if err := imp.Read(); err != nil {
					return err
				}
				objs, _ := imp.Data().All(context.Background())
				added, failed := 0, 0
				for _, obj := range objs {
					if e := catFacadeRepo.Save(context.Background(), obj); e != nil {
						failed++
					} else {
						added++
					}
				}
				fmt.Printf("imported: %d, skipped: %d\n", added, failed)
				return nil
			})
			if err := timer.NewTimerDecorator(cmd).Execute(); err != nil {
				fmt.Println("error:", err)
			}
		case "15":
			format := strings.ToLower(readString(in, "Format (csv/json/yaml): "))
			path := readString(in, "File path: ")
			cmd := commandpkg.CommandFunc(func() error {
				var imp interface {
					Read() error
					Data() repository.ICommonRepo
				}
				switch format {
				case "csv":
					imp = csvimporter.NewCSVOperationImporter(path)
				case "json":
					imp = jsonimporter.NewJSONOperationImporter(path)
				case "yaml":
					imp = yamlimporter.NewYAMLOperationImporter(path)
				}
				if imp == nil {
					fmt.Println("unknown format")
					return nil
				}
				if err := imp.Read(); err != nil {
					return err
				}
				objs, _ := imp.Data().All(context.Background())
				added, failed := 0, 0
				for _, obj := range objs {
					if e := opRepo.Save(context.Background(), obj); e != nil {
						failed++
					} else {
						added++
					}
				}
				fmt.Printf("imported: %d, skipped: %d\n", added, failed)
				return nil
			})
			if err := timer.NewTimerDecorator(cmd).Execute(); err != nil {
				fmt.Println("error:", err)
			}
		case "16":
			id := readUUID(in, "Account ID (uuid): ")
			a, err := bankF.GetAccount(service.ObjectID(id))
			if err != nil {
				fmt.Println("error:", err)
				break
			}
			fmt.Printf("ID=%s | name=%s | balance=%.2f\n",
				uuid.UUID(a.ID()).String(), a.Name(), a.Balance())

		case "17":
			id := readUUID(in, "Account ID (uuid): ")
			newName := readString(in, "New name: ")
			if err := bankF.UpdateAccountName(service.ObjectID(id), newName); err != nil {
				fmt.Println("error:", err)
			} else {
				fmt.Println("ok")
			}

		case "18":
			id := readUUID(in, "Account ID (uuid): ")
			newBal := readFloat(in, "New balance: ")
			if err := bankF.UpdateAccountBalance(service.ObjectID(id), newBal); err != nil {
				fmt.Println("error:", err)
			} else {
				fmt.Println("ok")
			}

		case "19":
			id := readUUID(in, "Category ID (uuid): ")
			c, err := catF.GetCategory(service.ObjectID(id))
			if err != nil {
				fmt.Println("error:", err)
				break
			}
			fmt.Printf("ID=%s | name=%s | type=%d\n",
				uuid.UUID(c.ID()).String(), c.Name(), int(c.Type()))

		case "20":
			id := readUUID(in, "Category ID (uuid): ")
			newName := readString(in, "New name: ")
			if err := catF.UpdateCategoryName(service.ObjectID(id), newName); err != nil {
				fmt.Println("error:", err)
			} else {
				fmt.Println("ok")
			}

		case "21":
			id := readUUID(in, "Category ID (uuid): ")
			if err := catF.DeleteCategory(service.ObjectID(id)); err != nil {
				fmt.Println("error:", err)
			} else {
				fmt.Println("deleted")
			}

		// 22) Get operation by ID
		case "22":
			id := readUUID(in, "Operation ID (uuid): ")
			o, err := opF.GetOperation(service.ObjectID(id))
			if err != nil {
				fmt.Println("error:", err)
				break
			}
			fmt.Printf("ID=%s | type=%d | account=%s | amount=%.2f | ts=%s | cat=%s | descr=%s\n",
				uuid.UUID(o.ID()).String(),
				int(o.Type()),
				uuid.UUID(o.BankAccountID()).String(),
				o.Amount(),
				o.Date().Format(time.RFC3339),
				uuid.UUID(o.CategoryID()).String(),
				o.Description(),
			)

		case "23":
			id := readUUID(in, "Operation ID (uuid): ")
			if err := opF.DeleteOperation(service.ObjectID(id)); err != nil {
				fmt.Println("error:", err)
			} else {
				fmt.Println("deleted")
			}

		case "0":
			fmt.Println("Bye!")
			return
		default:
			fmt.Println("Unknown choice")
		}
	}
}

func readString(in *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	s, _ := in.ReadString('\n')
	return strings.TrimSpace(s)
}

func readFloat(in *bufio.Reader, prompt string) float64 {
	for {
		s := readString(in, prompt)
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			return v
		}
		fmt.Println("Invalid float, try again")
	}
}

func readInt(in *bufio.Reader, prompt string) int {
	for {
		s := readString(in, prompt)
		if v, err := strconv.Atoi(s); err == nil {
			return v
		}
		fmt.Println("Invalid int, try again")
	}
}

func readUUID(in *bufio.Reader, prompt string) uuid.UUID {
	for {
		s := readString(in, prompt)
		if id, err := uuid.Parse(strings.TrimSpace(s)); err == nil {
			return id
		}
		fmt.Println("Invalid uuid, try again")
	}
}

func readTime(in *bufio.Reader, prompt string) time.Time {
	for {
		s := readString(in, prompt)
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			return t
		}
		fmt.Println("Invalid time, expected RFC3339, try again")
	}
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}
