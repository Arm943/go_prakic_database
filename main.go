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
	5 - Поиск контакта        |  6 - Поиск по тегу 
	--------------------------------------------------
	0 - Выход                 | 
	`)
		fmt.Print("Выберите действие: ")

		fmt.Scan(&userInputMenu)
		switch userInputMenu {
		case 1:
			showContactsWithTags(conn)
		case 2:
			addNumber(conn)
		case 3:
			update(conn)
		case 4:
			delete(conn)
		case 5:
			searchContact(conn)
		case 6:
			searchByTag(conn)
		case 0:
			return
		default:
			fmt.Println("Неверный ввод, попробуйте снова.")
		}
	}

}

// Показать все контакты с их тегами
func showContactsWithTags(conn *pgx.Conn) {
	// Запрос для получения всех контактов с их тегами
	rows, err := conn.Query(ctx, `
	SELECT 
  users.id, 
  users.name, 
  users.phone_number, 
  STRING_AGG(tags.tag, ', ')
FROM users
JOIN users_tags ON users.id = users_tags.user_id
JOIN tags ON tags.id = users_tags.tag_id
GROUP BY users.id, users.name, users.phone_number;
	`)
	if err != nil {
		log.Fatalf("❌ Ошибка при выполнении запроса: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name, phone, tags string
		err := rows.Scan(&id, &name, &phone, &tags)
		if err != nil {
			log.Fatalf("❌ Ошибка при чтении строки: %v", err)
		}
		fmt.Printf("👤 %s 📱 %s 🏷️  %s\n", name, phone, tags)
	}
}

// создание новой записи в книге
func addNumber(conn *pgx.Conn) {
	var name string
	var phoneNumber string
	var tag string
	var userID int
	var tagID int

	// Ввод имени и номера телефона
	fmt.Print("Введите ваше имя: ")
	fmt.Scan(&name)
	fmt.Scanln() // очистка буфера \n
	fmt.Print("Введите ваш номер телефона: ")
	fmt.Scan(&phoneNumber)
	fmt.Scanln() // очистка буфера \n

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

	//вставляем новые данные в таблицу и при этом получаем данные ID
	err = tx.QueryRow(ctx, `INSERT INTO users (name, phone_number) VALUES($1, $2) RETURNING id;`, name, phoneNumber).Scan(&userID)
	if err != nil {
		tx.Rollback(ctx) // если ошибка — откат
		log.Fatalf("❌ Ошибка при добавлении данных пользователя: %v", err)
	}

	err = tx.Commit(ctx) // фиксируем изменения
	if err != nil {
		log.Fatalf("❌ Ошибка при подтверждении транзакции: %v", err)
	}

	// функционал добавления тегов
	fmt.Print("Пропишите теги к контакту: ")
	fmt.Scan(&tag)

	// Вставка данных в таблицу
	err = conn.QueryRow(ctx, `
	INSERT INTO tags (tag) VALUES($1) ON CONFLICT (tag) DO NOTHING RETURNING id;
	`, tag).Scan(&tagID)
	if err == pgx.ErrNoRows {
		//тег уже есть, просто получаем его ID
		err = conn.QueryRow(ctx, `SELECT id FROM tags WHERE tag = $1`, tag).Scan(&tagID)
		if err != nil {
			log.Fatalf("❌ Ошибка при получении id существующего тега: %v", err)
		}
	}

	// Связать userID и tagID в таблице users_tags

	_, err = conn.Exec(ctx, `
	INSERT INTO users_tags (user_id, tag_id) VALUES($1,$2);
	`, userID, tagID)

	fmt.Println("Ваши данные контакта успешно добавлены!")
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

// поиск контактов по тегам

func searchByTag(conn *pgx.Conn) {
	var tag string
	fmt.Print("Введите тег для поиска всех контактов: ")
	fmt.Scan(&tag)

	// поиск по тегу с использованием индекса
	rows, err := conn.Query(ctx, `
SELECT u.name, u.phone_number
		FROM users u
		JOIN users_tags ut ON u.id = ut.user_id
		JOIN tags t ON t.id = ut.tag_id
		WHERE t.tag = $1;
`, tag)
	if err != nil {
		log.Fatalf("❌ Ошибка при выполнении запроса: %v", err)
	}
	defer rows.Close()

	// выводим данные контактов по тегу
	fmt.Println("Контакты с тегом:", tag)
	for rows.Next() {
		var name, phoneNumber string
		err := rows.Scan(&name, &phoneNumber)
		if err != nil {
			log.Fatalf("❌ Ошибка при считывании данных: %v", err)
		}
		fmt.Printf("%s,  %s\n", name, phoneNumber)
	}
	// Если ошибок нет, выводим сообщение
	if err := rows.Err(); err != nil {
		log.Fatalf("❌ Ошибка при обработке строк: %v", err)
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
	// индекс по имени
	_, err = conn.Exec(ctx, `
CREATE INDEX IF NOT EXISTS index_name ON users(name);
`)
	if err != nil {
		log.Fatalf("❌ Ошибка при создании индекса по name: %v", err)
	}

	// уникальный индекс по тегу
	_, err = conn.Exec(ctx, `
CREATE UNIQUE INDEX IF NOT EXISTS idx_tag_unique ON tags(tag);
`)
	if err != nil {
		log.Fatalf("❌ Ошибка при создании уникального индекса по tag: %v", err)
	}
	fmt.Println("✅ Индексы успешно созданы!")

	menu(conn)
	defer conn.Close(ctx)

}
