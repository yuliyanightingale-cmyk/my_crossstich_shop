package server

import (
	"fmt"
	"html/template"
	"log"
	"my_crossstich_shop/pkg/models"
	"my_crossstich_shop/pkg/repository"
	"net/http"
	"time"
)

type srv struct {
	db repository.DB
}

func New(db repository.DB) *srv {
	return &srv{
		db: db,
	}
}

func (s *srv) Run() error {
	// ==================== НАСТРОЙКА СЕРВЕРА ====================
	// Статические файлы
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Основные маршруты сайта
	http.HandleFunc("/", s.homeHandler)
	http.HandleFunc("/catalog", s.catalogHandler)
	http.HandleFunc("/about", s.aboutHandler)
	http.HandleFunc("/contacts", s.contactsHandler)
	http.HandleFunc("/feedback", s.feedbackHandler)

	// Админские маршруты
	http.HandleFunc("/admin/feedback", s.adminFeedbackHandler)
	http.HandleFunc("/admin/stats", s.adminStatsHandler)

	// ==================== ЗАПУСК СЕРВЕРА ====================
	port := ":8080"
	fmt.Printf("\n✅ Сервер запущен по адресу: http://localhost%s\n", port)
	fmt.Println("Админка отзывов: http://localhost:8080/admin/feedback")
	fmt.Println("Статистика: http://localhost:8080/admin/stats")
	fmt.Println("\n🛑 Для остановки нажмите Ctrl+C")

	err := http.ListenAndServe(port, nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *srv) homeHandler(w http.ResponseWriter, r *http.Request) {
	works, err := s.db.GetFeaturedWorks()
	if err != nil {
		log.Printf("❌ Ошибка загрузки данных для главной: %v", err)
		http.Error(w, "Ошибка загрузки данных", http.StatusInternalServerError)
		return
	}

	data := struct {
		PageTitle string
		Works     []models.CrossStitch
	}{
		PageTitle: "Главная",
		Works:     works,
	}

	s.renderTemplate(w, "index.html", data)
}

func (s *srv) catalogHandler(w http.ResponseWriter, r *http.Request) {
	works, err := s.db.GetAllWorks()
	if err != nil {
		log.Printf("❌ Ошибка загрузки каталога: %v", err)
		http.Error(w, "Ошибка загрузки каталога", http.StatusInternalServerError)
		return
	}

	data := struct {
		PageTitle string
		Works     []models.CrossStitch
	}{
		PageTitle: "Каталог",
		Works:     works,
	}

	s.renderTemplate(w, "catalog.html", data)
}

func (s *srv) aboutHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		PageTitle string
	}{
		PageTitle: "О нас",
	}

	s.renderTemplate(w, "about.html", data)
}

func (s *srv) contactsHandler(w http.ResponseWriter, r *http.Request) {
	success := r.URL.Query().Get("success") == "true"

	data := struct {
		PageTitle string
		Success   bool
	}{
		PageTitle: "Контакты",
		Success:   success,
	}

	s.renderTemplate(w, "contacts.html", data)
}

func (s *srv) feedbackHandler(w http.ResponseWriter, r *http.Request) {
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
	err = s.db.SaveFeedback(name, email, description)
	if err != nil {
		log.Printf("❌ Ошибка сохранения отзыва: %v", err)
		http.Error(w, "Ошибка сохранения отзыва", http.StatusInternalServerError)
		return
	}

	// Редирект с флагом успеха
	http.Redirect(w, r, "/contacts?success=true", http.StatusSeeOther)
}

// ==================== АДМИНСКИЕ ОБРАБОТЧИКИ ====================

func (s *srv) adminFeedbackHandler(w http.ResponseWriter, r *http.Request) {
	feedbacks, err := s.db.GetAllFeedback()
	if err != nil {
		log.Printf("❌ Ошибка загрузки отзывов: %v", err)
		http.Error(w, "Ошибка загрузки отзывов", http.StatusInternalServerError)
		return
	}

	data := struct {
		PageTitle string
		Feedbacks []models.Feedback
		Count     int
	}{
		PageTitle: "Админка - Отзывы",
		Feedbacks: feedbacks,
		Count:     len(feedbacks),
	}

	tmpl, err := template.ParseFiles("templates/base.html", "templates/admin_feedback.html")
	if err != nil {
		log.Printf("Ошибка загрузки шаблона админки: %v", err)
		http.Error(w, "Ошибка загрузки шаблона", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Ошибка отрисовки шаблона админки: %v", err)
		http.Error(w, "Ошибка отрисовки", http.StatusInternalServerError)
	}
}

func (s *srv) adminStatsHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем статистику отзывов
	feedbackStats, err := s.db.GetFeedbackStats()
	if err != nil {
		log.Printf("Ошибка получения статистики: %v", err)
		http.Error(w, "Ошибка получения статистики", http.StatusInternalServerError)
		return
	}

	data := struct {
		PageTitle     string
		FeedbackStats map[string]interface{}
		ServerTime    string
	}{
		PageTitle:     "Админка - Статистика",
		FeedbackStats: feedbackStats,
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

func (s *srv) renderTemplate(w http.ResponseWriter, tmplName string, data interface{}) {
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
