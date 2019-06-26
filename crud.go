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
	subject  string
	read     int
	priority int
}

// тут вы должны уже писать ваши хендлеры
func (m *Messager) List(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		w.Write([]byte(`{"status": 501, "error": "db error"}`))
		return
	}

	user, err := m.DB.Query("SELECT user_id FROM sessions WHERE id = $1", session.Value)
	if err != nil {
		w.Write([]byte(`{"status": 502, "error": "db error"}`))
		return
	}
	defer user.Close()

	user_val := 0
	for user.Next() {
		err = user.Scan(&user_val)
		if err != nil {
			w.Write([]byte(`{"status": 503, "error": "db error"}`))
			return
		}
	}

	items := []*Item{}
	rows, err := m.DB.Query("SELECT id, user_id, subject, read, priority FROM messages WHERE user_id = $1 ORDER BY id ASC", user_val)

	if err != nil {
		w.Write([]byte(`{"status": 504, "error": "db error"}`))
		return
	}

	defer rows.Close()

	for rows.Next() {
		it := &Item{}
		err := rows.Scan(&it.id, &it.user_id, &it.subject, &it.read, &it.priority)
		if err != nil {
			w.Write([]byte(`{"status": 501, "error": "db error"}`))
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
