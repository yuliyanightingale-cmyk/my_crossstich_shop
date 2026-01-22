package repository

import (
	"database/sql"
	"fmt"
	"my_crossstich_shop/pkg/models"

	_ "github.com/lib/pq" // драйвер PostgreSQL
)

type DB interface {
	initDatabase() error
	seedCatalogData() error
	showDatabaseStats() error
	GetAllWorks() ([]models.CrossStitch, error)
	GetFeaturedWorks() ([]models.CrossStitch, error)
	SaveFeedback(name, email, description string) error
	GetAllFeedback() ([]models.Feedback, error)
	GetFeedbackStats() (map[string]interface{}, error)
}

type db struct {
	conn *sql.DB
}

func New(cfg *models.Config) (*db, error) {
	// ==================== СТРОКА ПОДКЛЮЧЕНИЯ К POSTGRESQL ====================
	pgConnStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DbHost, cfg.DbPort, cfg.DbUser, cfg.DbPassword, cfg.DbName, cfg.DbSslmode,
	)
	// ==================== ИНИЦИАЛИЗАЦИЯ СОЕДИНЕНИЯ С POSTGRESQL ====================
	conn, err := sql.Open("postgres", pgConnStr)
	if err != nil {
		return nil, err
	}

	// Проверяем соединение
	err = conn.Ping()
	if err != nil {
		return nil, err
	}

	db := &db{
		conn: conn,
	}

	err = db.initDatabase()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (r *db) initDatabase() error {
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

	_, err := r.conn.Exec(createWorksTable)
	if err != nil {
		return fmt.Errorf("Ошибка создания таблицы cross_stitch_works: %w", err)
	}

	// 2. Таблица для отзывов
	createFeedbackTable := `
    CREATE TABLE IF NOT EXISTS feedback (
        id SERIAL PRIMARY KEY,
        name VARCHAR(100) NOT NULL,
        email VARCHAR(150) NOT NULL,
        description TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT now()
    );`

	_, err = r.conn.Exec(createFeedbackTable)
	if err != nil {
		return fmt.Errorf("Ошибка создания таблицы feedback: %w", err)
	}

	// 3. Заполняем каталог начальными данными если он пустой
	err = r.seedCatalogData()
	if err != nil {
		return fmt.Errorf("Ошибка заполнения каталога: %w", err)
	}
	// 4. Показываем статистику
	err = r.showDatabaseStats()
	if err != nil {
		return fmt.Errorf("Ошибка показа статистики: %w", err)
	}

	return nil
}

func (r *db) seedCatalogData() error {
	// Проверяем, есть ли данные в каталоге
	var count int
	err := r.conn.QueryRow("SELECT COUNT(*) FROM cross_stitch_works").Scan(&count)
	if err != nil {
		return fmt.Errorf("Ошибка проверки данных каталога: %w", err)
	}

	// Если таблица пустая - заполняем
	if count == 0 {
		fmt.Println("Заполняем каталог начальными данными...")

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
			_, err := r.conn.Exec(`
                INSERT INTO cross_stitch_works (title, size, price, description, image_url) 
                VALUES ($1, $2, $3, $4, $5)`,
				work.title, work.size, work.price, work.description, work.imageURL)
			if err != nil {
				return fmt.Errorf("Ошибка вставки работы %w", err)
			} else {
				fmt.Printf("Добавлено: %s\n", work.title)
			}
		}
		fmt.Println("Начальные данные добавлены в каталог")
	} else {
		fmt.Printf("В каталоге уже есть %d работ\n", count)
	}

	return nil
}

func (r *db) showDatabaseStats() error {
	// Статистика каталога
	var worksCount int
	err := r.conn.QueryRow("SELECT COUNT(*) FROM cross_stitch_works").Scan(&worksCount)
	if err != nil {
		return fmt.Errorf("Ошибка проверки статистики: %w", err)
	}

	// Статистика отзывов
	var feedbackCount int
	err = r.conn.QueryRow("SELECT COUNT(*) FROM feedback").Scan(&feedbackCount)
	if err != nil {
		return fmt.Errorf("Ошибка проверки запросов: %w", err)
	}

	fmt.Printf("Статистика базы данных:\n")
	fmt.Printf("Товаров в каталоге: %d\n", worksCount)
	fmt.Printf("Отзывов: %d\n", feedbackCount)

	return nil
}

func (r *db) GetAllWorks() ([]models.CrossStitch, error) {
	rows, err := r.conn.Query(`
        SELECT id, title, size, price, description, image_url 
        FROM cross_stitch_works 
        ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса каталога: %v", err)
	}
	defer rows.Close()

	var works []models.CrossStitch
	for rows.Next() {
		var w models.CrossStitch
		err := rows.Scan(&w.ID, &w.Title, &w.Size, &w.Price, &w.Description, &w.ImageURL)
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения данных каталога: %v", err)
		}
		works = append(works, w)
	}

	return works, nil
}

func (r *db) GetFeaturedWorks() ([]models.CrossStitch, error) {
	rows, err := r.conn.Query(`
        SELECT id, title, size, price, description, image_url 
        FROM cross_stitch_works 
        ORDER BY id 
        LIMIT 3`)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса избранных работ: %v", err)
	}
	defer rows.Close()

	var works []models.CrossStitch
	for rows.Next() {
		var w models.CrossStitch
		err := rows.Scan(&w.ID, &w.Title, &w.Size, &w.Price, &w.Description, &w.ImageURL)
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения данных: %v", err)
		}
		works = append(works, w)
	}

	return works, nil
}

// ==================== ФУНКЦИИ ДЛЯ РАБОТЫ С ОТЗЫВАМИ ====================

func (r *db) SaveFeedback(name, email, description string) error {
	fmt.Printf("Сохранение отзыва в PostgreSQL: %s, %s\n", name, email)

	// Явно указываем created_at для совместимости
	_, err := r.conn.Exec(`
        INSERT INTO feedback (name, email, description, created_at) 
        VALUES ($1, $2, $3, now())`,
		name, email, description)

	if err != nil {
		fmt.Printf("Ошибка сохранения отзыва: %v\n", err)
		return err
	}

	fmt.Printf("Отзыв успешно сохранен\n")
	return nil
}

func (r *db) GetAllFeedback() ([]models.Feedback, error) {
	rows, err := r.conn.Query(`
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

	var feedbacks []models.Feedback
	for rows.Next() {
		var f models.Feedback
		err := rows.Scan(&f.ID, &f.Name, &f.Email, &f.Description, &f.CreatedAt)
		if err != nil {
			return nil, err
		}
		feedbacks = append(feedbacks, f)
	}

	return feedbacks, nil
}

func (r *db) GetFeedbackStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Общее количество отзывов
	var total int
	err := r.conn.QueryRow("SELECT COUNT(*) FROM feedback").Scan(&total)
	if err != nil {
		return nil, err
	}
	stats["total"] = total

	// Отзывы за последние 7 дней
	var last7days int
	err = r.conn.QueryRow(`
        SELECT COUNT(*) FROM feedback 
        WHERE created_at >= now() - interval '7 days' 
           OR created_at IS NULL`).Scan(&last7days)
	if err != nil {
		return nil, err
	}
	stats["last7days"] = last7days

	// Последний отзыв
	var lastFeedback string
	err = r.conn.QueryRow(`
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
