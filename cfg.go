package main

var (
	//  если у вас другие параметры доступа - вы можете поменять их тут
	// этот файл можно менять чтобы работало у вас, но эту переменную никуда не переносить ( иначе на сервере будет ошибка из-за дублирования )
	connStr = "user=root dbname=mydb password=root host=127.0.0.1 port=5432 sslmode=disable"
)
