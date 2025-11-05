package timer

import (
	"time"
	"fmt"
	
	command "github.com/ilyaytrewq/kpo-sb/homework/BankService/Command"
)

type TimerDecorator struct {
    command command.Command
}

func (t *TimerDecorator) Execute() error {
    start := time.Now()
    err := t.command.Execute()
    fmt.Printf("Command executed in %v\n", time.Since(start))
    return err
}