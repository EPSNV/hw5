package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

// CaseResponse
type CR map[string]interface{}

type Case struct {
	Method string // GET по-умолчанию в http.NewRequest если передали пустую строку
	Path   string
	Query  string
	Status int
	Result interface{}
	Body   interface{}
	Cookie *http.Cookie
}

var (
	client = &http.Client{Timeout: time.Second}
)

func TestQueries(t *testing.T) {

	// параметры доступа в connStr находятся в cfg.go
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatal("cant open db:", err)
	}

	err = db.Ping()
	if err != nil {
		t.Fatal("cant connect to db:", err)
	}

	_, err = db.Exec(dbDump)
	if err != nil {
		t.Fatal("cant import db dump:", err)
	}

	msgrg := NewMessager(db)

	mux := &http.ServeMux{}
	mux.HandleFunc("/list", msgrg.List)
	mux.HandleFunc("/create", msgrg.Create)
	mux.HandleFunc("/mark", msgrg.Mark)
	mux.HandleFunc("/delete", msgrg.Delete)

	userFIRSTCookie := &http.Cookie{Name: "session_id", Value: "9289be379007b2d17039ac3e10db194e"}
	userSECONDCookie := &http.Cookie{Name: "session_id", Value: "009c1a3246d029855beeb3d25890d446"}
	userUNKNOWNCookie := &http.Cookie{Name: "session_id", Value: "eb236b830a450847aadb166b1353d9c0"}

	ts := httptest.NewServer(mux)

	cases := []Case{
		// список писем первого пользователя
		Case{
			Path:   "/list",
			Cookie: userFIRSTCookie,
			Result: CR{
				"messages": []CR{
					CR{
						"id":       2,
						"subject":  "second letter",
						"read":     false,
						"priority": "high",
					},
					CR{
						"id":       3,
						"subject":  "third letter",
						"read":     false,
						"priority": "normal",
					},
					CR{
						"id":       4,
						"subject":  "fourth letter",
						"read":     true,
						"priority": "low",
					},
				},
			},
		},
		// список писем второго пользователя
		Case{
			Path:   "/list",
			Cookie: userSECONDCookie,
			Result: CR{
				"messages": []CR{
					CR{
						"id":       7,
						"subject":  "2/2",
						"read":     false,
						"priority": "normal",
					},
					CR{
						"id":       8,
						"subject":  "2/3",
						"read":     false,
						"priority": "low",
					},
				},
			},
		},
		// список писем с несуществующей авторизацией
		Case{
			Path:   "/list",
			Cookie: userUNKNOWNCookie,
			Status: http.StatusUnauthorized,
			Result: CR{
				"error": "no_auth",
			},
		},

		// удаление письма, юзер без авторизации
		Case{
			Path:   "/delete",
			Method: http.MethodDelete,
			Query:  "id=8",
			Status: http.StatusUnauthorized,
			Result: CR{
				"error": "no_auth",
			},
		},
		// удаление письма, не тот юзер ( письмо 8 принадлежит юзеру 2)
		Case{
			Path:   "/delete",
			Method: http.MethodDelete,
			Query:  "id=8",
			Cookie: userFIRSTCookie,
			Status: http.StatusNotFound,
			Result: CR{
				"error": "no_record",
			},
		},
		// удаление письма, письмо не найдено
		Case{
			Path:   "/delete",
			Method: http.MethodDelete,
			Query:  "id=100000",
			Cookie: userSECONDCookie,
			Status: http.StatusNotFound,
			Result: CR{
				"error": "no_record",
			},
		},
		// удаление письма
		Case{
			Path:   "/delete",
			Method: http.MethodDelete,
			Query:  "id=8",
			Cookie: userSECONDCookie,
			Result: CR{
				"affected": 1,
			},
		},

		// удаление письма, не тот метод
		Case{
			Path:   "/delete",
			Method: http.MethodPost,
			Status: http.StatusBadRequest,
			Query:  "id=8",
			Cookie: userSECONDCookie,
			Result: CR{
				"error": "bad_method",
			},
		},

		// помните про sql-инъекции? входящие параматреы необходимо проверять и передавать через плейсхолдеры ($1, $2, $3)
		Case{
			Path:   "/delete",
			Method: http.MethodDelete,
			Status: http.StatusBadRequest,
			Query:  "id=8+or+id%3D7",
			Cookie: userSECONDCookie,
			Result: CR{
				"error": "bad_id",
			},
		},

		// список писем второго пользователя, теперь тут только 1 письмо
		Case{
			Path:   "/list",
			Cookie: userSECONDCookie,
			Result: CR{
				"messages": []CR{
					CR{
						"id":       7,
						"subject":  "2/2",
						"read":     false,
						"priority": "normal",
					},
				},
			},
		},

		// для начала проверяем авторизацию
		Case{
			Path:   "/mark",
			Method: http.MethodPost,
			Status: http.StatusUnauthorized,
			Query:  "id=7",
			Result: CR{
				"error": "no_auth",
			},
		},

		// id 7 принадлежит 2-му юзеру, поэтому StatusNotFound
		Case{
			Path:   "/mark",
			Method: http.MethodPost,
			Status: http.StatusNotFound,
			Query:  "id=7&read=true",
			Cookie: userFIRSTCookie,
			Result: CR{
				"error": "no_record",
			},
		},

		// если не передано что менять - должна быть ошибка
		Case{
			Path:   "/mark",
			Method: http.MethodPost,
			Status: http.StatusBadRequest,
			Query:  "id=7",
			Cookie: userSECONDCookie,
			Result: CR{
				"error": "nothing_to_change",
			},
		},

		// read может принимать только значения true/false
		Case{
			Path:   "/mark",
			Method: http.MethodPost,
			Status: http.StatusBadRequest,
			Query:  "id=7&read=invalid",
			Cookie: userSECONDCookie,
			Result: CR{
				"error": "bad_read",
			},
		},

		// priority может принимать только значения high/low/normal
		Case{
			Path:   "/mark",
			Method: http.MethodPost,
			Status: http.StatusBadRequest,
			Query:  "id=7&priority=invalid",
			Cookie: userSECONDCookie,
			Result: CR{
				"error": "bad_priority",
			},
		},

		// нормальное обновление, 2 поля
		Case{
			Path:   "/mark",
			Method: http.MethodPost,
			Query:  "id=7&priority=high&read=true",
			Cookie: userSECONDCookie,
			Result: CR{
				"affected": 1,
			},
		},

		Case{
			Path:   "/list",
			Cookie: userSECONDCookie,
			Result: CR{
				"messages": []CR{
					CR{
						"id":       7,
						"subject":  "2/2",
						"read":     true,
						"priority": "high",
					},
				},
			},
		},

		// нормальное обновление, priority
		Case{
			Path:   "/mark",
			Method: http.MethodPost,
			Query:  "id=7&priority=low",
			Cookie: userSECONDCookie,
			Result: CR{
				"affected": 1,
			},
		},

		// нормальное обновление, read
		Case{
			Path:   "/mark",
			Method: http.MethodPost,
			Query:  "id=7&read=false",
			Cookie: userSECONDCookie,
			Result: CR{
				"affected": 1,
			},
		},

		// проверяем что все изменилось
		Case{
			Path:   "/list",
			Cookie: userSECONDCookie,
			Result: CR{
				"messages": []CR{
					CR{
						"id":       7,
						"subject":  "2/2",
						"read":     false,
						"priority": "low",
					},
				},
			},
		},

		//  создание, без авторизации
		Case{
			Path:   "/create",
			Method: http.MethodPut,
			Status: http.StatusUnauthorized,
			Query:  `id=11&priority=invalid&subject=test"`,
			Result: CR{
				"error": "no_auth",
			},
		},

		//  создание, неправильный priority
		Case{
			Path:   "/create",
			Method: http.MethodPut,
			Status: http.StatusBadRequest,
			Query:  `id=11&priority=invalid&subject=test"`,
			Cookie: userSECONDCookie,
			Result: CR{
				"error": "bad_priority",
			},
		},

		// создание, все ок
		// обратите внимание, id игнорируется, а в subject кавычка ( используйте плейсхолдеры, никогда не подставляйте напрямую в запрос! )
		Case{
			Path:   "/create",
			Method: http.MethodPut,
			Query:  `id=11&priority=high&subject=test'`,
			Cookie: userSECONDCookie,
			Result: CR{
				"id": 9,
			},
		},

		// проверяем что все записалось
		Case{
			Path:   "/list",
			Cookie: userSECONDCookie,
			Result: CR{
				"messages": []CR{
					CR{
						"id":       7,
						"subject":  "2/2",
						"read":     false,
						"priority": "low",
					},
					CR{
						"id":       9,
						"subject":  "test'",
						"read":     false,
						"priority": "high",
					},
				},
			},
		},
	}

	for idx, item := range cases {
		var (
			err      error
			result   interface{}
			expected interface{}
			req      *http.Request
		)

		caseName := fmt.Sprintf("case %d: [%s] %s %s", idx, item.Method, item.Path, item.Query)

		// если у вас случилась это ошибка - значит вы не делаете где-то rows.Close и у вас текут соединения с базой
		if db.Stats().OpenConnections != 1 {
			t.Fatalf("[%s] you have %d open connections, must be 1", caseName, db.Stats().OpenConnections)
		}

		if item.Method == "" || item.Method == http.MethodGet || item.Method == http.MethodDelete {
			req, err = http.NewRequest(item.Method, ts.URL+item.Path+"?"+item.Query, nil)
		} else {
			req, err = http.NewRequest(item.Method, ts.URL+item.Path, strings.NewReader(item.Query))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		}

		if item.Cookie != nil {
			req.AddCookie(item.Cookie)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("[%s] request error: %v", caseName, err)
			continue
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		// fmt.Printf("[%s] body: %s\n", caseName, string(body))
		if item.Status == 0 {
			item.Status = http.StatusOK
		}

		if resp.StatusCode != item.Status {
			t.Fatalf("[%s] expected http status %v, got %v\n\tresponse: %s", caseName, item.Status, resp.StatusCode, body)
			continue
		}

		err = json.Unmarshal(body, &result)
		if err != nil {
			t.Fatalf("[%s] cant unpack json: %v\n\trecieved %s", caseName, err, body)
			continue
		}

		// reflect.DeepEqual не работает если нам приходят разные типы
		// а там приходят разные типы (string VS interface{}) по сравнению с тем что в ожидаемом результате
		// этот маленький грязный хак конвертит данные сначала в json, а потом обратно в interface - получаем совместимые результаты
		// не используйте это в продакшен-коде - надо явно писать что ожидается интерфейс или использовать другой подход с точным форматом ответа
		data, err := json.Marshal(item.Result)
		json.Unmarshal(data, &expected)

		if !reflect.DeepEqual(result, expected) {
			t.Fatalf("[%s] results not match\nGot : %#v\nWant: %#v", caseName, result, expected)
			continue
		}
	}
}

var dbDump = `
DROP TABLE IF EXISTS "messages";
DROP SEQUENCE IF EXISTS messages_id_seq;
CREATE SEQUENCE messages_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 2147483647 START 9 CACHE 1;

CREATE TABLE "public"."messages" (
    "id" integer DEFAULT nextval('messages_id_seq') NOT NULL,
    "user_id" integer DEFAULT '0' NOT NULL,
    "subject" character varying(255) DEFAULT '' NOT NULL,
    "read" smallint DEFAULT '0' NOT NULL,
    "priority" smallint DEFAULT '0' NOT NULL
) WITH (oids = false);

CREATE INDEX "messages_user_id" ON "public"."messages" USING btree ("user_id");

INSERT INTO "messages" ("id", "user_id", "subject", "read", "priority") VALUES
(4,	1,	'fourth letter',	1,	1),
(3,	1,	'third letter',	'0',	3),
(2,	1,	'second letter',	'0',	5),
(7,	2,	'2/2',	'0',	3),
(8,	2,	'2/3',	'0',	1);

DROP TABLE IF EXISTS "sessions";
CREATE TABLE "public"."sessions" (
    "id" character varying(255) NOT NULL,
    "user_id" integer NOT NULL
) WITH (oids = false);

CREATE INDEX "sessions_user_id" ON "public"."sessions" USING btree ("user_id");

INSERT INTO "sessions" ("id", "user_id") VALUES
('9289be379007b2d17039ac3e10db194e',	1),
('009c1a3246d029855beeb3d25890d446',	2);

`
