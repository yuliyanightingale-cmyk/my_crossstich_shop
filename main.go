package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"my_crossstich_shop/pkg/config"
	"net/http"
	"time"

	_ "github.com/lib/pq" // драйвер PostgreSQL
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

// Структура для отзывов
type Feedback struct {
	ID          int
	Name        string
	Email       string
	Description string
	CreatedAt   string
}

// Глобальная переменная для базы данных
var db *sql.DB

func main() {
	fmt.Println("🚀 Запуск сервера...")

	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	// ==================== ПОДКЛЮЧЕНИЕ К POSTGRESQL ====================
	pgConnStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DbHost, cfg.DbPort, cfg.DbUser, cfg.DbPassword, cfg.DbName, cfg.DbSslmode,
	)

	db, err = sql.Open("postgres", pgConnStr)
	if err != nil {
		log.Fatal("❌ Ошибка подключения к PostgreSQL:", err)
	}
	defer db.Close()

	// Проверяем подключение
	err = db.Ping()
	if err != nil {
		log.Fatal("❌ Не удалось подключиться к PostgreSQL:", err)
	}
	fmt.Println("✅ Подключение к PostgreSQL установлено")

	// ==================== ИНИЦИАЛИЗАЦИЯ БАЗЫ ====================
	initDatabase()

	// ==================== НАСТРОЙКА СЕРВЕРА ====================
	// Статические файлы
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Основные маршруты сайта
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/catalog", catalogHandler)
	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/contacts", contactsHandler)
	http.HandleFunc("/feedback", feedbackHandler)

	// Админские маршруты
	http.HandleFunc("/admin/feedback", adminFeedbackHandler)
	http.HandleFunc("/admin/stats", adminStatsHandler)

	// ==================== ЗАПУСК СЕРВЕРА ====================
	port := ":8080"
	fmt.Printf("\n✅ Сервер запущен по адресу: http://localhost%s\n", port)
	fmt.Println("📊 Админка отзывов: http://localhost:8080/admin/feedback")
	fmt.Println("📈 Статистика: http://localhost:8080/admin/stats")
	fmt.Println("\n🛑 Для остановки нажмите Ctrl+C")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("❌ Ошибка запуска сервера:", err)
	}
}

// ==================== ИНИЦИАЛИЗАЦИЯ БАЗЫ ДАННЫХ ====================

func initDatabase() {
	// 1. Таблица для вышивок (каталог)
	createWorksTable := `
    CREATE TABLE IF NOT EXISTS cross_stitch_works (
        id SERIAL PRIMARY KEY,
        title VARCHAR(200) NOT NULL,
        size VARCHAR(100),
        price INTEGER NOT NULL,
        description TEXT,
        image_url VARCHAR(500) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`

	_, err := db.Exec(createWorksTable)
	if err != nil {
		log.Fatal("❌ Ошибка создания таблицы cross_stitch_works:", err)
	}
	fmt.Println("✅ Таблица cross_stitch_works создана/проверена")

	// 2. Таблица для отзывов
	createFeedbackTable := `
    CREATE TABLE IF NOT EXISTS feedback (
        id SERIAL PRIMARY KEY,
        name VARCHAR(100) NOT NULL,
        email VARCHAR(150) NOT NULL,
        description TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT now()
    );`

	_, err = db.Exec(createFeedbackTable)
	if err != nil {
		log.Fatal("❌ Ошибка создания таблицы feedback:", err)
	}
	fmt.Println("✅ Таблица feedback создана/проверена")

	// 3. Заполняем каталог начальными данными если он пустой
	seedCatalogData()

	// 4. Показываем статистику
	showDatabaseStats()
}

func seedCatalogData() {
	// Проверяем, есть ли данные в каталоге
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM cross_stitch_works").Scan(&count)
	if err != nil {
		log.Printf("⚠️ Ошибка проверки данных каталога: %v", err)
		return
	}

	// Если таблица пустая - заполняем
	if count == 0 {
		fmt.Println("📝 Заполняем каталог начальными данными...")

		works := []struct {
			title, size, description, imageURL string
			price                              int
		}{
			{
				"«Шаман»",
				"57x75 см (1170x1560 крестиков)",
				"Яркая композиция, выполненная нитками DMC. Идеально для гостиной или стилизации интерьера.",
				"/static/images/shaman.jpg",
				4500,
			},
			{
				"«Фантазия»",
				"60x75 см (1560x1960 крестиков)",
				"Хранитель снов и фантазий. Использованы оттенки синего и фиолетового.",
				"/static/images/fantasy.jpg",
				6800,
			},
			{
				"«Золотая рыбка»",
				"50x75 см (980x1170 крестиков)",
				"Портрет девушки у моря в окружении золотых рыбок. Подходит для подарка.",
				"/static/images/gold_fish.jpg",
				3800,
			},
			{
				"«Не нужно слов»",
				"50x75 см (980x1170 крестиков)",
				"Влюбленная пара. Создаёт уютную атмосферу в интерьере.",
				"/static/images/no_words.jpg",
				5200,
			},
		}

		for _, work := range works {
			_, err := db.Exec(`
                INSERT INTO cross_stitch_works (title, size, price, description, image_url) 
                VALUES ($1, $2, $3, $4, $5)`,
				work.title, work.size, work.price, work.description, work.imageURL)
			if err != nil {
				log.Printf("⚠️ Ошибка вставки работы '%s': %v", work.title, err)
			} else {
				fmt.Printf("  ✅ Добавлено: %s\n", work.title)
			}
		}
		fmt.Println("✅ Начальные данные добавлены в каталог")
	} else {
		fmt.Printf("📊 В каталоге уже есть %d работ\n", count)
	}
}

func showDatabaseStats() {
	// Статистика каталога
	var worksCount int
	db.QueryRow("SELECT COUNT(*) FROM cross_stitch_works").Scan(&worksCount)

	// Статистика отзывов
	var feedbackCount int
	db.QueryRow("SELECT COUNT(*) FROM feedback").Scan(&feedbackCount)

	fmt.Printf("📊 Статистика базы данных:\n")
	fmt.Printf("  🛍️  Товаров в каталоге: %d\n", worksCount)
	fmt.Printf("  💬 Отзывов: %d\n", feedbackCount)
}

// ==================== ФУНКЦИИ ДЛЯ РАБОТЫ С КАТАЛОГОМ ====================

func getAllWorks() ([]CrossStitch, error) {
	rows, err := db.Query(`
        SELECT id, title, size, price, description, image_url 
        FROM cross_stitch_works 
        ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса каталога: %v", err)
	}
	defer rows.Close()

	var works []CrossStitch
	for rows.Next() {
		var w CrossStitch
		err := rows.Scan(&w.ID, &w.Title, &w.Size, &w.Price, &w.Description, &w.ImageURL)
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения данных каталога: %v", err)
		}
		works = append(works, w)
	}

	return works, nil
}

func getFeaturedWorks() ([]CrossStitch, error) {
	rows, err := db.Query(`
        SELECT id, title, size, price, description, image_url 
        FROM cross_stitch_works 
        ORDER BY id 
        LIMIT 3`)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса избранных работ: %v", err)
	}
	defer rows.Close()

	var works []CrossStitch
	for rows.Next() {
		var w CrossStitch
		err := rows.Scan(&w.ID, &w.Title, &w.Size, &w.Price, &w.Description, &w.ImageURL)
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения данных: %v", err)
		}
		works = append(works, w)
	}

	return works, nil
}

// ==================== ФУНКЦИИ ДЛЯ РАБОТЫ С ОТЗЫВАМИ ====================

func saveFeedback(name, email, description string) error {
	log.Printf("💾 Сохранение отзыва в PostgreSQL: %s, %s", name, email)

	// Явно указываем created_at для совместимости
	_, err := db.Exec(`
        INSERT INTO feedback (name, email, description, created_at) 
        VALUES ($1, $2, $3, now())`,
		name, email, description)

	if err != nil {
		log.Printf("❌ Ошибка сохранения отзыва: %v", err)
		return err
	}

	log.Printf("✅ Отзыв успешно сохранен")
	return nil
}

func getAllFeedback() ([]Feedback, error) {
	rows, err := db.Query(`
        SELECT id, name, email, description, 
               COALESCE(
                   TO_CHAR(created_at, 'DD.MM.YYYY HH24:MI'),
                   TO_CHAR(now(), 'DD.MM.YYYY HH24:MI')
               ) as created_at
        FROM feedback 
        ORDER BY COALESCE(created_at, now()) DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feedbacks []Feedback
	for rows.Next() {
		var f Feedback
		err := rows.Scan(&f.ID, &f.Name, &f.Email, &f.Description, &f.CreatedAt)
		if err != nil {
			return nil, err
		}
		feedbacks = append(feedbacks, f)
	}

	return feedbacks, nil
}

func getFeedbackStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Общее количество отзывов
	var total int
	err := db.QueryRow("SELECT COUNT(*) FROM feedback").Scan(&total)
	if err != nil {
		return nil, err
	}
	stats["total"] = total

	// Отзывы за последние 7 дней
	var last7days int
	err = db.QueryRow(`
        SELECT COUNT(*) FROM feedback 
        WHERE created_at >= now() - interval '7 days' 
           OR created_at IS NULL`).Scan(&last7days)
	if err != nil {
		return nil, err
	}
	stats["last7days"] = last7days

	// Последний отзыв
	var lastFeedback string
	err = db.QueryRow(`
        SELECT COALESCE(
            TO_CHAR(created_at, 'DD.MM.YYYY HH24:MI'),
            TO_CHAR(now(), 'DD.MM.YYYY HH24:MI')
        )
        FROM feedback 
        ORDER BY COALESCE(created_at, now()) DESC LIMIT 1`).Scan(&lastFeedback)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	stats["lastFeedback"] = lastFeedback

	return stats, nil
}

// ==================== ОСНОВНЫЕ ОБРАБОТЧИКИ СТРАНИЦ ====================

func homeHandler(w http.ResponseWriter, r *http.Request) {
	works, err := getFeaturedWorks()
	if err != nil {
		log.Printf("❌ Ошибка загрузки данных для главной: %v", err)
		http.Error(w, "Ошибка загрузки данных", http.StatusInternalServerError)
		return
	}

	data := struct {
		PageTitle string
		Works     []CrossStitch
	}{
		PageTitle: "Главная",
		Works:     works,
	}

	renderTemplate(w, "index.html", data)
}

func catalogHandler(w http.ResponseWriter, r *http.Request) {
	works, err := getAllWorks()
	if err != nil {
		log.Printf("❌ Ошибка загрузки каталога: %v", err)
		http.Error(w, "Ошибка загрузки каталога", http.StatusInternalServerError)
		return
	}

	data := struct {
		PageTitle string
		Works     []CrossStitch
	}{
		PageTitle: "Каталог",
		Works:     works,
	}

	renderTemplate(w, "catalog.html", data)
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		PageTitle string
	}{
		PageTitle: "О нас",
	}

	renderTemplate(w, "about.html", data)
}

func contactsHandler(w http.ResponseWriter, r *http.Request) {
	success := r.URL.Query().Get("success") == "true"

	data := struct {
		PageTitle string
		Success   bool
	}{
		PageTitle: "Контакты",
		Success:   success,
	}

	renderTemplate(w, "contacts.html", data)
}

func feedbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Получаем данные из формы
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Ошибка обработки формы", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	description := r.FormValue("description")

	// Валидация
	if name == "" || email == "" || description == "" {
		http.Error(w, "Все поля обязательны для заполнения", http.StatusBadRequest)
		return
	}

	// Сохраняем в PostgreSQL
	err = saveFeedback(name, email, description)
	if err != nil {
		log.Printf("❌ Ошибка сохранения отзыва: %v", err)
		http.Error(w, "Ошибка сохранения отзыва", http.StatusInternalServerError)
		return
	}

	// Редирект с флагом успеха
	http.Redirect(w, r, "/contacts?success=true", http.StatusSeeOther)
}

// ==================== АДМИНСКИЕ ОБРАБОТЧИКИ ====================

func adminFeedbackHandler(w http.ResponseWriter, r *http.Request) {
	feedbacks, err := getAllFeedback()
	if err != nil {
		log.Printf("❌ Ошибка загрузки отзывов: %v", err)
		http.Error(w, "Ошибка загрузки отзывов", http.StatusInternalServerError)
		return
	}

	data := struct {
		PageTitle string
		Feedbacks []Feedback
		Count     int
	}{
		PageTitle: "Админка - Отзывы",
		Feedbacks: feedbacks,
		Count:     len(feedbacks),
	}

	tmpl, err := template.ParseFiles("templates/base.html", "templates/admin_feedback.html")
	if err != nil {
		log.Printf("❌ Ошибка загрузки шаблона админки: %v", err)
		http.Error(w, "Ошибка загрузки шаблона", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("❌ Ошибка отрисовки шаблона админки: %v", err)
		http.Error(w, "Ошибка отрисовки", http.StatusInternalServerError)
	}
}

func adminStatsHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем статистику отзывов
	feedbackStats, err := getFeedbackStats()
	if err != nil {
		log.Printf("❌ Ошибка получения статистики: %v", err)
		http.Error(w, "Ошибка получения статистики", http.StatusInternalServerError)
		return
	}

	// Получаем статистику каталога
	var catalogCount int
	err = db.QueryRow("SELECT COUNT(*) FROM cross_stitch_works").Scan(&catalogCount)
	if err != nil {
		log.Printf("❌ Ошибка получения статистики каталога: %v", err)
	}

	data := struct {
		PageTitle     string
		FeedbackStats map[string]interface{}
		CatalogCount  int
		ServerTime    string
	}{
		PageTitle:     "Админка - Статистика",
		FeedbackStats: feedbackStats,
		CatalogCount:  catalogCount,
		ServerTime:    time.Now().Format("02.01.2006 15:04:05"),
	}

	tmpl, err := template.ParseFiles("templates/base.html", "templates/admin_stats.html")
	if err != nil {
		log.Printf("❌ Ошибка загрузки шаблона статистики: %v", err)
		http.Error(w, "Ошибка загрузки шаблона", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("❌ Ошибка отрисовки шаблона статистики: %v", err)
		http.Error(w, "Ошибка отрисовки", http.StatusInternalServerError)
	}
}

// ==================== ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ====================

func renderTemplate(w http.ResponseWriter, tmplName string, data interface{}) {
	tmpl, err := template.ParseFiles("templates/base.html", "templates/"+tmplName)
	if err != nil {
		log.Printf("❌ Ошибка загрузки шаблона %s: %v", tmplName, err)
		http.Error(w, "Ошибка загрузки шаблона", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("❌ Ошибка отрисовки шаблона %s: %v", tmplName, err)
		http.Error(w, "Ошибка отрисовки", http.StatusInternalServerError)
	}
}
