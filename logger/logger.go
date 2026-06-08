package logger

import (
	"fmt"
	"log/syslog"
)

var sl *syslog.Writer

func Init() error {
	var err error
	sl, err = syslog.New(syslog.LOG_INFO|syslog.LOG_USER, "grocy-scanner")
	return err
}

func Close() {
	if sl != nil {
		sl.Close()
	}
}

func LogAdd(productName string, quantity float64, location, expiry string) {
	if sl == nil {
		return
	}
	msg := fmt.Sprintf("ADD product=%q qty=%g", productName, quantity)
	if location != "" {
		msg += fmt.Sprintf(" location=%q", location)
	}
	if expiry != "" && expiry != "2999-12-31" {
		msg += fmt.Sprintf(" expiry=%q", expiry)
	}
	sl.Info(msg)
}

func LogConsume(productName string, quantity float64) {
	if sl == nil {
		return
	}
	sl.Info(fmt.Sprintf("CONSUME product=%q qty=%g", productName, quantity))
}

func LogEditName(productName, oldName string) {
	if sl == nil {
		return
	}
	sl.Info(fmt.Sprintf("EDIT_NAME product=%q old_name=%q", productName, oldName))
}

func LogShoppingList(productName string) {
	if sl == nil {
		return
	}
	sl.Info(fmt.Sprintf("SHOPPING_LIST product=%q", productName))
}

func LogError(msg string) {
	if sl == nil {
		return
	}
	sl.Err(msg)
}
