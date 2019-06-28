package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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
func (m *Messager) Auth(r *http.Request) (int, error) {
	session, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		fmt.Println("ErrNoCookie")
		return 0, err
	}
	fmt.Println("session id")
	var userID int
	rows, err := m.DB.Query("SELECT user_id FROM sessions WHERE id = $1", session.Value)
	defer rows.Close()
	if err != nil {
		fmt.Printf("error2 = %s \n", err.Error())
		return 0, err
	}
	fmt.Println("user id")
	for rows.Next() {
		err := rows.Scan(&userID)
		if err != nil {
			fmt.Printf("Scanerror = %s \n", err.Error())
			return 0, err
		}
		fmt.Println("Scan in userID")
	}

	fmt.Printf("Scan in userV = %d\n", userID)
	return userID, nil
}
func (m *Messager) List(w http.ResponseWriter, r *http.Request) {
	var ReadMap = map[int]bool{
		0: false,
		1: true,
	}

	var PriorityMap = map[int]string{
		1: "low",
		3: "normal",
		5: "high",
	}
	userID, err := m.Auth(r)
	if err != nil {
		w.Write([]byte(`{"status": 401, "error": "auth error"}`))
		return
	}
	if userID == 0 {
		w.WriteHeader(401)
		w.Write([]byte(`{"error": "no_auth"}`))
		return
	}
	items := []map[string]interface{}{}
	rows, err := m.DB.Query("SELECT id, subject, read, priority FROM messages WHERE user_id = $1 ORDER BY id ASC", userID)

	if err != nil {
		w.Write([]byte(`{"status": 504, "error": "db error"}`))
		return
	}

	defer rows.Close()
	fmt.Println("message`s data")
	for rows.Next() {

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

	var ReadMap = map[bool]int{
		false: 0,
		true:  1,
	}

	var PriorityMap = map[string]int{
		"low":    1,
		"normal": 3,
		"high":   5,
	}

	userID, _ := m.Auth(r)
	//Тут я пропускаю ошибки, потому что при ошибках userID = 0 и дальше это обрабатывается, надо разобраться с этим костылем при наличии времени
	// if err != nil {
	// 	w.Write([]byte(`{"status": 401, "error": "auth error"}`))
	// 	return
	// }
	if userID == 0 {
		w.WriteHeader(401)
		w.Write([]byte(`{"error": "no_auth"}`))
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(400)
		w.Write([]byte(`{"error": "bad_method"}`))
		return
	}
	ID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(`{"error": "bad_id"}`))
		return
	}
	// READ, err := strconv.Atoi(r.URL.Query().Get("read"))
	// if err != nil {
	// 	w.WriteHeader(400)
	// 	w.Write([]byte(`{"error": "bad_read"}`))
	// 	return
	// }
	// PRIORITY, err := strconv.Atoi(r.URL.Query().Get("priority"))
	// if err != nil {
	// 	w.WriteHeader(400)
	// 	w.Write([]byte(`{"error": "bad_priority"}`))
	// 	return
	// }
	fmt.Printf("ID = %d\n", ID)
	var messageUserID int
	rows, err := m.DB.Query("SELECT user_id FROM messages WHERE id = $1 ", ID)

	if err != nil {
		w.Write([]byte(`{"status": 506, "error": "db error"}`))
		return
	}
	defer rows.Close()
	for rows.Next() {

		err := rows.Scan(&messageUserID)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
	}
	fmt.Printf("messageUserID = %d\n", messageUserID)
	if messageUserID != userID {
		w.WriteHeader(404)
		w.Write([]byte(`{"error": "no_record"}`))
		return
	}
	// query := `UPDATE messages SET read = store_count + $1 WHERE id = $2`
	// _, err = s.db.Exec(query, cnt, productID)
	// _, err = m.DB.Exec("DELETE FROM messages WHERE id = $1", ID)
	// if err != nil {
	// 	w.Write([]byte(`{"status": 500, "error": "db err"}`))
	// 	return
	// }
	// w.Write([]byte(`{"affected":1}`))
}

func (m *Messager) Delete(w http.ResponseWriter, r *http.Request) {

	userID, _ := m.Auth(r)
	//Тут я пропускаю ошибки, потому что при ошибках userID = 0 и дальше это обрабатывается, надо разобраться с этим костылем при наличии времени
	// if err != nil {
	// 	w.Write([]byte(`{"status": 401, "error": "auth error"}`))
	// 	return
	// }
	if userID == 0 {
		w.WriteHeader(401)
		w.Write([]byte(`{"error": "no_auth"}`))
		return
	}
	if r.Method != http.MethodDelete {
		w.WriteHeader(400)
		w.Write([]byte(`{"error": "bad_method"}`))
		return
	}
	ID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(`{"error": "bad_id"}`))
		return
	}
	fmt.Printf("ID = %d\n", ID)
	var messageUserID int
	rows, err := m.DB.Query("SELECT user_id FROM messages WHERE id = $1 ", ID)

	if err != nil {
		w.Write([]byte(`{"status": 505, "error": "db error"}`))
		return
	}
	defer rows.Close()
	for rows.Next() {

		err := rows.Scan(&messageUserID)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
	}
	fmt.Printf("messageUserID = %d\n", messageUserID)
	if messageUserID != userID {
		w.WriteHeader(404)
		w.Write([]byte(`{"error": "no_record"}`))
		return
	}
	_, err = m.DB.Exec("DELETE FROM messages WHERE id = $1", ID)
	if err != nil {
		w.Write([]byte(`{"status": 500, "error": "db err"}`))
		return
	}
	w.Write([]byte(`{"affected":1}`))
}

func (m *Messager) Create(w http.ResponseWriter, r *http.Request) {

}
