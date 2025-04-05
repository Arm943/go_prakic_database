package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

var ctx = context.Background()

func menu(conn *pgx.Conn) {
	var userInputMenu int

	for {
		fmt.Print(`
	1 - Показать все контакты |  2 - Добавить контакт
	--------------------------------------------------
	3 - Изменить контакт      |  4 - Удалить контакт
	--------------------------------------------------
	5 - Поиск контакта        |  0 - Выход     
	`)
		fmt.Print("Выберите действие: ")

		fmt.Scan(&userInputMenu)
		switch userInputMenu {
		case 1:
			outpute(conn)
		case 2:
			addNumber(conn)
		case 3:
			update(conn)
		case 4:
			delete(conn)
		case 5:
			searchContact(conn)
		case 0:
			return
		default:
			fmt.Println("Неверный ввод, попробуйте снова.")
		}
	}

}

func outpute(conn *pgx.Conn) {
	// вывод всех пользователей
	rows, err := conn.Query(ctx, `
	SELECT * FROM users ORDER BY id ASC;
	`)
	if err != nil {
		log.Fatalf("❌ Ошибка при запросе данных %v", err)
	}
	defer rows.Close()

	// проверка на пустую таблицу
	if !rows.Next() {
		fmt.Println("Телефонная книга пуста!")
		return
	}

	fmt.Println("Список всех пользователей:")
	for rows.Next() {
		var id int
		var userName string
		var phoneNumber string
		err := rows.Scan(&id, &userName, &phoneNumber)
		if err != nil {
			log.Fatalf("❌ Ошибка при считывании строки: %v", err)
		}
		fmt.Printf("ID: %d | Имя: %s | Телефон: %s\n", id, userName, phoneNumber)
	}

}

// ввод данных от пользователя
func addNumber(conn *pgx.Conn) {
	var name string
	var phoneNumber string

	// Ввод имени и номера телефона
	fmt.Print("Введите ваше имя: ")
	fmt.Scan(&name)
	fmt.Scanln() // очистка буфера \n
	fmt.Print("Введите ваш номер телефона: ")
	fmt.Scan(&phoneNumber)

	// Проверка на пустые данные
	if name == "" || phoneNumber == "" {
		fmt.Println("Имя или номер телефона не могут быть пустыми!")
		return
	}

	// Вставка данных в таблицу через транзакцию
	tx, err := conn.Begin(ctx)
	if err != nil {
		log.Fatalf("❌ Ошибка при начале транзакции: %v", err)
	}

	_, err = tx.Exec(ctx, `INSERT INTO users (name, phone_number) VALUES($1, $2);`, name, phoneNumber)
	if err != nil {
		tx.Rollback(ctx) // если ошибка — откат
		log.Fatalf("❌ Ошибка при добавлении данных пользователя: %v", err)
	}

	err = tx.Commit(ctx) // фиксируем изменения
	if err != nil {
		log.Fatalf("❌ Ошибка при подтверждении транзакции: %v", err)
	}

	fmt.Println("Ваши данные успешно добавлены!")

}

// изменение данных
func update(conn *pgx.Conn) {
	var id int
	var userName string
	var phoneNumber string
	fmt.Print("Введите ID контакта, который нужный изменить: ")
	fmt.Scan(&id)
	fmt.Scanln() // очистка буфера \n
	fmt.Print("Введите имя: ")
	fmt.Scan(&userName)
	fmt.Scanln() // очистка буфера \n
	fmt.Print("Введите ваш номер телефона: ")
	fmt.Scan(&phoneNumber)

	tx, err := conn.Begin(ctx)
	if err != nil {
		log.Fatalf("❌ Ошибка при начале транзакции: %v", err)
	}
	_, err = tx.Exec(ctx, `
	UPDATE users SET name = $1, phone_number = $2 WHERE id = $3;
`, userName, phoneNumber, id)

	if err != nil {
		tx.Rollback(ctx)
		log.Fatalf("❌ Ошибка при добавлении данных пользователя: %v", err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		log.Fatalf("❌ Ошибка при сохранении измененных данных: %v", err)
	}

	fmt.Println("Ваши данные успешно добавлены!")

}

// удаление записи
func delete(conn *pgx.Conn) {
	var id int
	fmt.Print("❌ Введите ID контакта, который нужно удалить: ")
	fmt.Scan(&id)

	_, err := conn.Exec(ctx, `
	DELETE FROM users WHERE id =$1;
	`, id)
	if err != nil {
		log.Fatalf("❌ Ошибка при удалении данных пользователя: %v", err)
	}
	fmt.Println("Пользователь успешно удален!")
}

// поиск номера телефона по имени
func searchContact(conn *pgx.Conn) {
	var name string
	var id int
	var phoneNumber string

	fmt.Print("Введите имя для поиска: ")
	fmt.Scan(&name)

	//Поиск с использованием индекса
	rows, err := conn.Query(ctx, `
	SELECT id, name, phone_number FROM users WHERE name = $1
	`, name)
	if err != nil {
		log.Fatalf("❌ Ошибка при выполнении запроса для поиска: %v", err)
	}
	defer rows.Close()

	// Если есть результаты, то выводим их

	found := false
	for rows.Next() {
		err := rows.Scan(&id, &name, &phoneNumber)
		if err != nil {
			log.Fatalf("❌ Ошибка при сканировании строки: %v", err)
		}
		fmt.Printf("Найден контакт: %s, %s, ID: %d\n", name, phoneNumber, id)
		found = true
	}

	if !found {
		fmt.Println("Контакт с таким именем не найден.")
	}
	if rows.Err() != nil {
		log.Fatalf("❌ Ошибка при переборе строки: %v", err)
	}

}

func main() {

	// Строка подключения
	connStr := "postgres://postgres:1234@localhost:5432/mydb"

	// Подключение к базе
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Fatalf("❌ Ошибка подключения: %v", err)
	}
	defer conn.Close(ctx)

	fmt.Println("✅ Успешное подключение к базе!")

	// Создаем таблицу, если она не существует

	_, err = conn.Exec(ctx, `
CREATE TABLE IF NOT EXISTS users(
id SERIAL PRIMARY KEY,
name TEXT NOT NULL,
phone_number TEXT NOT NULL
);
`)
	if err != nil {
		log.Fatalf("❌ Ошибка при создании таблицы: %v", err)
	}
	fmt.Println("✅ Таблица users успешно создана!")

	// создаем таблицу с тегами

	_, err = conn.Exec(ctx, `
CREATE TABLE IF NOT EXISTS tags(
id SERIAL PRIMARY KEY,
tag TEXT NOT NULL
);
`)
	if err != nil {
		log.Fatalf("❌ Ошибка при создании таблицы тегов: %v", err)
	}
	fmt.Println("✅ Таблица tags успешно создана!")

	// создаем промежуточную таблицу для "многие-ко-многим"

	_, err = conn.Exec(ctx, `
CREATE TABLE IF NOT EXISTS users_tags(
user_id INT REFERENCES users(id),
tag_id INT REFERENCES tags(id),
PRIMARY KEY (user_id,tag_id)
);
`)
	if err != nil {
		log.Fatalf("❌ Ошибка при создании таблицы users-tags: %v", err)
	}
	fmt.Println("✅ Таблица users-tags тегов успешно создана!")

	// создаем индекс по столбу name
	_, err = conn.Exec(ctx, `
CREATE INDEX IF NOT EXISTS index_name ON users(name);
`)
	if err != nil {
		log.Fatalf("❌ Ошибка при создании индекса index_name: %v", err)
	}
	fmt.Println("✅ Индексы для поля name успешно созданы!")

	menu(conn)

}
