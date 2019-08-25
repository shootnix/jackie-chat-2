package io

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/shootnix/jackie-chat-2/config"
	"log"
	"sync"
)

type Pg struct {
	Conn *sql.DB
}

var once sync.Once
var pg *Pg

func GetPg() *Pg {
	once.Do(func() {
		pg = &Pg{pgConnect()}
	})
	return pg
}

func pgConnect() *sql.DB {
	pgCfg := config.GetConfig().Database.Pg
	connectString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s",
		pgCfg.Host,
		pgCfg.Port,
		pgCfg.Username,
		pgCfg.Password,
		pgCfg.DBName,
	)

	db, err := sql.Open("postgres", connectString)
	if err != nil {
		log.Fatal("Can't connect to postgres: ", err.Error())
	}

	if err := db.Ping(); err != nil {
		log.Fatal("Can't ping database:", err.Error())
	}

	return db
}
