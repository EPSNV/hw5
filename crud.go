package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
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
	ReadMap := map[int]bool{
		0: false,
		1: true,
	}

	PriorityMap := map[int]string{
		1: "low",
		3: "normal",
		5: "high",
	}

	session, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		w.Write([]byte(`{"status": 501, "error": "db error"}`))
		return
	}
	fmt.Println("session id")
	var userV int
	user, err := m.DB.Query("SELECT user_id FROM sessions WHERE id = $1", session.Value)
	defer user.Close()
	if err != nil {
		fmt.Printf("error2 = %s \n", err.Error())
		w.Write([]byte(`{"status": 503, "error": "db error"}`))
		return
	}
	fmt.Println("user id")
	for user.Next() {
		user.Scan(&userV)
		fmt.Println("Scan in userV")
	}
	// } else {
	// 	fmt.Printf("error1 = %s \n", err.Error())
	// 	w.Write([]byte(`{"status": 502, "error": "ErrNoRows"}`))
	// 	return
	// }

	// user, err := m.DB.Query("SELECT user_id FROM sessions WHERE id = $1", session.Value)

	// if err != nil {
	// 	w.Write([]byte(`{"status": 502, "error": "no auth"}`))
	// 	return
	// }
	// defer user.Close()

	// user_val := 0
	// i := 0
	// if user.Next() {
	// 	err = user.Scan(&user_val)
	// 	if err != nil {
	// 		fmt.Printf("error2 = %s \n", err.Error())
	// 		w.Write([]byte(`{"status": 503, "error": "db error"}`))
	// 		return
	// 	}
	// 	i = i + 1
	// } else {
	// 	fmt.Printf("12341234243244")
	// }
	// if i == 0 {
	// 	fmt.Printf("i not changed")
	// }

	items := []map[string]interface{}{}
	rows, err := m.DB.Query("SELECT id, subject, read, priority FROM messages WHERE user_id = $1 ORDER BY id ASC", userV)

	if err != nil {
		w.Write([]byte(`{"status": 504, "error": "db error"}`))
		return
	}

	defer rows.Close()
	fmt.Println("message`s data")
	for rows.Next() {
		//it := map[string]interface{}{}
		var id, subject interface{}
		var read, priority int
		err := rows.Scan(&id, &subject, &read, &priority)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		fmt.Printf("id=%d, subject=%s, read=%d, prio=%d, priomap=%s\n", id, subject, read, priority, PriorityMap[priority])
		items = append(items, map[string]interface{}{"id": id, "subject": subject, "read": ReadMap[read], "priority": PriorityMap[priority]})
	}
	fmt.Println("message`s data in JSON")
	result, _ := json.Marshal(map[string]interface{}{"messages": items})
	w.Write(result)
}

func (m *Messager) Mark(w http.ResponseWriter, r *http.Request) {

}

func (m *Messager) Delete(w http.ResponseWriter, r *http.Request) {

}

func (m *Messager) Create(w http.ResponseWriter, r *http.Request) {

}
