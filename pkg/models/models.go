package models

type Config struct {
	DbHost     string
	DbPort     string
	DbUser     string
	DbPassword string
	DbName     string
	DbSslmode  string
}

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
