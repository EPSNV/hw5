package main

import (
	"database/sql"
	"net/http"
)

func NewMessager(db *sql.DB) *Messager {
	// вы можете добавлять сюда логику, если вам это требуется
	// например, создать тут sqlx-коннект, а не обычный database/sql
	return &Messager{
		DB: db,
	}
}

type Messager struct {
	// вы можете добавлять сюда поля, если вам это требуется
	DB *sql.DB
}

// тут вы должны уже писать ваши хендлеры
func (m *Messager) List(http.ResponseWriter, *http.Request) {

}

func (m *Messager) Mark(http.ResponseWriter, *http.Request) {

}

func (m *Messager) Delete(http.ResponseWriter, *http.Request) {

}

func (m *Messager) Create(http.ResponseWriter, *http.Request) {

}
