package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

var ctx = context.Background()

func menu() {
	var userInputMenu int
	fmt.Print(`
	1 - Показать все контакты
	2 - Добавить контакт
	3 - Изменить контакт
	4 - Удалить контакт
	0 - Выход
	`)
	fmt.Print("Выберите действие: ")

	fmt.Scan(&userInputMenu)
	switch userInputMenu {
	case '1':

	case `2`:
	case `3`:
	case `4`:
	case `5`:

	}

}

func main() {
	menu()
	// Строка подключения
	connStr := "postgres://postgres:1234@localhost:5432/mydb"

	// Подключение к базе
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Fatalf("Ошибка подключения: %v", err)
	}
	defer conn.Close(ctx)

	fmt.Println("Успешное подключение к базе!")

	// Создаем таблицу, если она не существует

	_, err = conn.Exec(ctx, `
CREATE TABLE IF NOT EXISTS users(
id SERIAL PRIMARY KEY,
name TEXT NOT NULL,
age INT
);
`)
	if err != nil {
		log.Fatalf("Ошибка при создании таблицы: %v", err)
	}
	fmt.Println("Таблица users успешно создана!")

	menu()

	// ввод данных от пользователя
	var name string
	var age int
	fmt.Print("Введите ваше имя и возраст: ")
	fmt.Scanln(&name, &age)

	_, err = conn.Exec(ctx, `
	INSERT INTO users (name, age) VALUES($1, $2);
	`, name, age)
	if err != nil {
		log.Fatalf("Ошибка при добавлении данных пользователя %v", err)
	}
	fmt.Println("Ваши данные успешно добавлены!")

	rows, err := conn.Query(ctx, `
	SELECT * FROM users;
	`)
	if err != nil {
		log.Fatalf("Ошибка при запросе данных %v", err)
	}
	defer rows.Close()

	// вывод всех пользователей
	fmt.Println("Список всех пользователей:")
	for rows.Next() {
		var id int
		var userName string
		var userAge int
		err := rows.Scan(&id, &userName, &userAge)
		if err != nil {
			log.Fatalf("Ошибка при считывании строки: %v", err)
		}
		fmt.Printf("ID: %d | Имя: %s | Возраст: %d\n", id, userName, userAge)
	}

}
