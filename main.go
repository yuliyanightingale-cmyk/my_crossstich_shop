package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	_ "modernc.org/sqlite"
)

// Структура "Вышивка"
type CrossStitch struct {
	ID          int
	Title       string
	Size        string
	Price       int
	Description string
	ImageURL    string
}

// Данные для сайта
var stitchWorks = []CrossStitch{
	{
		ID:          1,
		Title:       "«Шаман»",
		Size:        "57x75 см (1170x1560 крестиков)",
		Price:       4500,
		Description: "Яркая композиция, выполненная нитками DMC. Идеально гостиной или стилизации интерьера.",
		ImageURL:    "/static/images/shaman.jpg",
	},
	{
		ID:          2,
		Title:       "«Фантазия»",
		Size:        "60x75 см (1560x1960 крестиков)",
		Price:       6800,
		Description: "Хранитель снов и фантазий. Использованы оттенки синего и фиолетового.",
		ImageURL:    "/static/images/fantasy.jpg",
	},
	{
		ID:          3,
		Title:       "«Золотая рыбка»",
		Size:        "50x75 см (980x1170 крестиков)",
		Price:       3800,
		Description: "Портрет девушки у моря в окружении золотых рыбок. Подходит для подарка.",
		ImageURL:    "/static/images/gold_fish.jpg",
	},
	{
		ID:          4,
		Title:       "«Не нужно слов»",
		Size:        "50x75 см (980x1170 крестиков)",
		Price:       5200,
		Description: "Влюбленная пара. Создаёт уютную атмосферу в интерьере.",
		ImageURL:    "/static/images/no_words.jpg",
	},
}

var (
	db  *sql.DB // объвляем глобальную переменную базы данных, чтобы иметь доступ из любых функций
	err error
)

func main() {
	db, err = sql.Open("sqlite", "my_store.db") // запускаем базу данных
	if err != nil {
		panic(err)
	}

	defer db.Close()

	// создаем таблицу если она еще не создана
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS feedback (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		email TEXT,
		description TEXT
	);`)
	if err != nil {
		panic(err)
	}

	// выполняем запрос на получение данных из таблицы
	rows, err := db.Query(`SELECT id, name, email, description FROM feedback;`)
	if err != nil {
		panic(err)
	}
	fmt.Println("Существующие запросы: ")
	// читаем полученные запросом данные
	for rows.Next() {
		var (
			id          int
			name        string
			email       string
			description string
		)
		err := rows.Scan(&id, &name, &email, &description)
		if err != nil {
			panic(err)
		}
		fmt.Println(id, name, email, description) // выводим данные в терминал
	}
	// Статические файлы
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Маршруты
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/catalog", catalogHandler)
	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/contacts", contactsHandler)
	http.HandleFunc("/feedback", feedbackHandler)

	// Запуск сервера
	port := ":8080"
	fmt.Printf("Сервер запущен: http://localhost%s\n", port)
	fmt.Println("Остановите сервер: Ctrl+C")
	http.ListenAndServe(port, nil)
}

// Главная страница
func homeHandler(w http.ResponseWriter, r *http.Request) {

	var featuredWorks []CrossStitch
	if len(stitchWorks) > 3 {
		featuredWorks = stitchWorks[:3]
	} else {
		featuredWorks = stitchWorks
	}

	data := struct {
		PageTitle string
		Works     []CrossStitch
	}{
		PageTitle: "Главная",
		Works:     featuredWorks,
	}

	tmpl, err := template.ParseFiles("templates/base.html", "templates/index.html")
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка отрисовки: "+err.Error(), http.StatusInternalServerError)
	}
}

// Каталог
func catalogHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		PageTitle string
		Works     []CrossStitch
	}{
		PageTitle: "Каталог",
		Works:     stitchWorks,
	}

	tmpl, err := template.ParseFiles("templates/base.html", "templates/catalog.html")
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка отрисовки: "+err.Error(), http.StatusInternalServerError)
	}
}

// О нас
func aboutHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		PageTitle string
		Works     []CrossStitch
	}{
		PageTitle: "О нас",
		Works:     []CrossStitch{},
	}

	tmpl, err := template.ParseFiles("templates/base.html", "templates/about.html")
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка отрисовки: "+err.Error(), http.StatusInternalServerError)
	}
}

// Контакты
func contactsHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		PageTitle string
		Works     []CrossStitch
	}{
		PageTitle: "Контакты",
		Works:     []CrossStitch{},
	}

	tmpl, err := template.ParseFiles("templates/base.html", "templates/contacts.html")
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка отрисовки: "+err.Error(), http.StatusInternalServerError)
	}
}

type feedbackRequest struct {
	Name        string
	Email       string
	Description string
}

// отправка формы
func feedbackHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/contacts", http.StatusPermanentRedirect)

	// парсим запрос
	fmt.Println(r.URL.RawQuery)
	rawFields := strings.Split(r.URL.RawQuery, "&")
	fmt.Println(rawFields)

	fbReq := feedbackRequest{}
	for i := range rawFields {
		rawField := strings.Split(rawFields[i], "=")
		switch rawField[0] {
		case "name":
			value, err := url.QueryUnescape(rawField[1])
			if err != nil {
				fmt.Println(err)
				return
			}
			fbReq.Name = strings.TrimSpace(value)
		case "email":
			value, err := url.QueryUnescape(rawField[1])
			if err != nil {
				fmt.Println(err)
				return
			}
			fbReq.Email = strings.TrimSpace(value)
		case "description":
			value, err := url.QueryUnescape(rawField[1])
			if err != nil {
				fmt.Println(err)
				return
			}
			fbReq.Description = strings.TrimSpace(value)
		}
	}
	fmt.Println(fbReq)

	// сохраняем в базу данных
	result, err := db.Exec(
		`INSERT INTO feedback (
			name, email, description
		) VALUES ($1, $2, $3)`,
		fbReq.Name, fbReq.Email, fbReq.Description,
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(result.LastInsertId()) // выводим в терминал идентификатор созданной записи
}
