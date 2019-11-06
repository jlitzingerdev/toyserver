package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

type DbService interface {
	CreateDb() error
	DropDb() error
}

type MessageService struct {
	db *sql.DB
}

func NewDbService() (*MessageService, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(127.0.0.1:3307)/", os.Getenv("DBUSER"), os.Getenv("DBPASS")))
	if err != nil {
		return nil, err
	}
	return &MessageService{db}, nil
}

func (svc *MessageService) Exec(query string) error {
	tx, err := svc.db.Begin()

	_, err = tx.Exec(query)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (svc *MessageService) CreateDb() error {
	return svc.Exec("CREATE DATABASE test")
}

func (svc *MessageService) DropDb() error {
	return svc.Exec("DROP DATABASE test")
}
