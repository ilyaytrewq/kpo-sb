package timer

import (
	"fmt"
	"time"

	command "github.com/ilyaytrewq/kpo-sb/homework/BankService/Command"
)

type TimerDecorator struct {
	command command.Command
}

func NewTimerDecorator(cmd command.Command) *TimerDecorator {
	return &TimerDecorator{command: cmd}
}

func (t *TimerDecorator) Execute() error {
	start := time.Now()
	err := t.command.Execute()
	fmt.Printf("Command executed in %v\n", time.Since(start))
	return err
}
