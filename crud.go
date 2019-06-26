package main

import (
	"database/sql"
	"encoding/json"
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

type Item struct {
	id       int
	user_id  int
	subject  int
	read     int
	priority int
}

// тут вы должны уже писать ваши хендлеры
func (m *Messager) List(w http.ResponseWriter, r *http.Request) {
	items := []*Item{}
	rows, err := m.DB.Query("SELECT id, name, price, cnt, sum, store_count FROM products ORDER BY id ASC")
	if err != nil {
		w.Write([]byte(`{"status": 500, "error": "db error"}`))
		return
	}
	defer rows.Close()
	for rows.Next() {
		it := &Item{}
		err := rows.Scan(&it.id, &it.user_id, &it.subject, &it.read, &it.priority)
		if err != nil {
			w.Write([]byte(`{"status": 500, "error": "db error"}`))
			return
		}
		items = append(items, it)
	}
	result, _ := json.Marshal(items)
	w.Write(result)
}

func (m *Messager) Mark(w http.ResponseWriter, r *http.Request) {

}

func (m *Messager) Delete(w http.ResponseWriter, r *http.Request) {

}

func (m *Messager) Create(w http.ResponseWriter, r *http.Request) {

}
