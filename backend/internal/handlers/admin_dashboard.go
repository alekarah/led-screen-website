package handlers

import (
	"log"
	"net/http"
	"os"
	"time"

	"ledsite/internal/models"

	"github.com/gin-gonic/gin"
)

// AdminDashboard - главная страница админки
func (h *Handlers) AdminDashboard(c *gin.Context) {
	var stats struct {
		ProjectsCount int64
		ImagesCount   int64
		ContactsCount int64
	}

	h.db.Model(&models.Project{}).Count(&stats.ProjectsCount)
	h.db.Model(&models.Image{}).Count(&stats.ImagesCount)
	h.db.Model(&models.ContactForm{}).Count(&stats.ContactsCount)

	// заявки за 7 дней
	var newContacts7 int
	h.db.Raw(`SELECT COUNT(*) FROM contact_forms WHERE created_at >= NOW() - INTERVAL '7 days'`).Scan(&newContacts7)

	// --- ПЕРЕЗВОНЫ (сегодня / просрочено / ближайшие) ---
	now := NowMSK()
	startToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, moscowLoc)
	endToday := startToday.Add(24 * time.Hour)

	var remindToday int64
	h.db.Model(&models.ContactForm{}).
		Where("remind_flag = ? AND remind_at IS NOT NULL AND remind_at >= ? AND remind_at < ?",
			true, startToday.UTC(), endToday.UTC()).
		Count(&remindToday)

	var remindOverdue int64
	h.db.Model(&models.ContactForm{}).
		Where("remind_flag = ? AND remind_at IS NOT NULL AND remind_at < ?",
			true, now.UTC()).
		Count(&remindOverdue)

	// Топ-5 проектов
	type TopProject struct {
		ProjectID   uint   `gorm:"column:project_id"`
		Title       string `gorm:"column:title"`
		Views       int64  `gorm:"column:views"`
		Description string `gorm:"column:description"`
	}

	var topProjects []TopProject
	if err := h.db.Table("project_view_dailies AS v").
		Select(`
			p.id    AS project_id,
			p.title AS title,
			SUM(v.views) AS views,
			p.description AS description
		`).
		Joins("JOIN projects p ON p.id = v.project_id").
		Where("v.day >= CURRENT_DATE - INTERVAL '29 days'").
		Group("p.id, p.title, p.description").
		Order("views DESC").
		Limit(5).
		Scan(&topProjects).Error; err != nil {
		topProjects = nil
	}

	// Сумма просмотров по дням (для простого графика/спарклайна)
	type DailyViews struct {
		Day   time.Time
		Views int64
	}
	var views30 []DailyViews
	if err := h.db.Table("project_view_dailies").
		Select("day, SUM(views) AS views").
		Where("day >= CURRENT_DATE - INTERVAL '29 days'").
		Group("day").
		Order("day ASC").
		Scan(&views30).Error; err != nil {
		views30 = nil
	}

	// --- СИСТЕМА ---
	var dbOK bool
	if sqlDB, err := h.db.DB(); err == nil {
		if err := sqlDB.Ping(); err == nil {
			dbOK = true
		}
	}

	sys := gin.H{
		"dbOK":      dbOK,
		"env":       os.Getenv("ENVIRONMENT"),
		"version":   os.Getenv("APP_VERSION"),
		"serverNow": time.Now(),
	}

	c.HTML(http.StatusOK, "admin_base.html", gin.H{
		"title":  "Админ панель",
		"PageID": "admin-dashboard",

		"stats":        stats,
		"newContacts7": newContacts7,

		"reminders": gin.H{
			"today":   remindToday,
			"overdue": remindOverdue,
		},

		"analytics": gin.H{
			"topProjects": topProjects,
			"views30":     views30,
		},

		"sys": sys,
	})

}

// AdminProjects - управление проектами
func (h *Handlers) AdminProjects(c *gin.Context) {
	var projects []models.Project
	h.db.Preload("Categories").Preload("Images").Order("sort_order ASC, created_at DESC").Find(&projects)

	var categories []models.Category
	h.db.Find(&categories)

	c.HTML(http.StatusOK, "admin_base.html", gin.H{
		"title":      "Управление проектами",
		"projects":   projects,
		"categories": categories,
		"PageID":     "admin-projects",
	})
}

// Возвращает количество заявок по дням за последние 7 дней (включая сегодня)
func (h *Handlers) AdminContacts7Days(c *gin.Context) {
	type row struct {
		Day   string `json:"day"`   // YYYY-MM-DD
		Count int    `json:"count"` // сколько заявок в этот день
	}
	var out []row

	// Postgres: generate_series + left join по contact_forms.created_at
	h.db.Raw(`
		WITH d AS (
		  SELECT generate_series(CURRENT_DATE - INTERVAL '6 days', CURRENT_DATE, INTERVAL '1 day')::date AS day
		)
		SELECT to_char(d.day, 'YYYY-MM-DD') AS day,
		       COALESCE(COUNT(cf.id), 0)::int AS count
		FROM d
		LEFT JOIN contact_forms cf ON DATE(cf.created_at) = d.day
		GROUP BY d.day
		ORDER BY d.day;
	`).Scan(&out)

	c.JSON(200, out)
}

// ResetAllViews — глобальный сброс всей статистики просмотров
func (h *Handlers) ResetAllViews(c *gin.Context) {
	if err := h.db.Exec(`TRUNCATE TABLE project_view_dailies RESTART IDENTITY`).Error; err != nil {
		log.Printf("Ошибка сброса статистики: %v", err)
		jsonErr(c, http.StatusInternalServerError, "Ошибка сброса статистики")
		return
	}
	jsonOK(c, gin.H{"ok": true})
}

// ResetProjectViews — сброс просмотров только для конкретного проекта
func (h *Handlers) ResetProjectViews(c *gin.Context) {
	id, ok := mustID(c)
	if !ok {
		return
	}
	if err := h.db.Where("project_id = ?", id).Delete(&models.ProjectViewDaily{}).Error; err != nil {
		log.Printf("Ошибка сброса статистики проекта %d: %v", id, err)
		jsonErr(c, http.StatusInternalServerError, "Ошибка сброса статистики проекта")
		return
	}
	jsonOK(c, gin.H{"ok": true})
}
