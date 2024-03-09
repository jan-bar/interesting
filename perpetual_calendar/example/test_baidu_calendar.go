package main

import (
	calendar "github.com/jan-bar/interesting/perpetual_calendar"
)

func main() {
	// dsn := "mysql:user:pass@tcp(127.0.0.1:3306)/janbar?charset=utf8mb4&parseTime=True&loc=Local"
	dsn := "sqlite:test.db"
	err := calendar.SaveCalendar(dsn)
	if err != nil {
		panic(err)
	}
}
