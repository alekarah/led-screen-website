package handlers

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"time"

	"ledsite/internal/models"

	"github.com/gin-gonic/gin"
)

var moscowLoc, _ = time.LoadLocation("Europe/Moscow")

func nowMoscow() time.Time {
	return time.Now().In(moscowLoc)
}

// /admin/contacts — страница всех заявок (с пагинацией)
func (h *Handlers) AdminContactsPage(c *gin.Context) {
	var contacts []models.ContactForm

	// --- Базовый запрос ---
	qb := h.db.Model(&models.ContactForm{})

	// --- Поиск (поддержим search и q) ---
	search := c.Query("search")
	if search == "" {
		search = c.Query("q")
	}
	if search != "" {
		q := "%" + search + "%"
		qb = qb.Where("name ILIKE ? OR phone ILIKE ? OR email ILIKE ? OR message ILIKE ?", q, q, q, q)
	}

	// --- Фильтр по статусу ---
	status := c.Query("status")
	if status != "" {
		qb = qb.Where("status = ?", status)
	}

	// --- Интервал дат ---
	dateRange := c.Query("date")
	now := nowMoscow()

	switch dateRange {
	case "today":
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, moscowLoc)
		end := start.Add(24 * time.Hour)
		qb = qb.Where("created_at >= ? AND created_at < ?", start.UTC(), end.UTC())
	case "7d":
		from := now.AddDate(0, 0, -7)
		qb = qb.Where("created_at >= ?", from.UTC())
	case "month":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, moscowLoc)
		end := start.AddDate(0, 1, 0)
		qb = qb.Where("created_at >= ? AND created_at < ?", start.UTC(), end.UTC())
	}

	// --- Пагинация ---
	// page >= 1, limit ∈ {25,50,100} (по умолчанию 50)
	page := 1
	if v, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && v > 0 {
		page = v
	}
	limit := 50
	if v, err := strconv.Atoi(c.DefaultQuery("limit", "50")); err == nil {
		if v == 25 || v == 50 || v == 100 {
			limit = v
		}
	}
	offset := (page - 1) * limit

	// --- Счётчик total (до LIMIT) ---
	var total int64
	if err := qb.Count(&total).Error; err != nil {
		c.String(http.StatusInternalServerError, "DB error")
		return
	}

	// --- Данные текущей страницы ---
	if err := qb.Order("created_at DESC").Limit(limit).Offset(offset).Find(&contacts).Error; err != nil {
		c.String(http.StatusInternalServerError, "DB error")
		return
	}

	// --- Кол-во страниц (минимум 1) ---
	pages := int((total + int64(limit) - 1) / int64(limit))
	if pages < 1 {
		pages = 1
	}

	// --- prev / next с ограничениями ---
	prevPage := page - 1
	if prevPage < 1 {
		prevPage = 1
	}
	nextPage := page + 1
	if nextPage > pages {
		nextPage = pages
	}

	// --- Набор номеров страниц (с троеточиями как -1) ---
	pageNumbers := buildPageNumbers(page, pages)

	// --- Рендер ---
	c.HTML(http.StatusOK, "admin_base.html", gin.H{
		"title":       "Заявки",
		"PageID":      "admin-contacts",
		"contactsAll": contacts, // текущая страница
		"total":       total,    // всего записей
		"page":        page,
		"pages":       pages,
		"prevPage":    prevPage,
		"nextPage":    nextPage,
		"pageNumbers": pageNumbers,
		"limit":       limit,
		"search":      search,
		"status":      status,
		"dateRange":   dateRange,
	})
}

// Смена статуса заявки
func (h *Handlers) UpdateContactStatus(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	// Проверка допустимых статусов
	if body.Status != "new" && body.Status != "processed" && body.Status != "archived" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Недопустимый статус"})
		return
	}

	if err := h.db.Model(&models.ContactForm{}).
		Where("id = ?", id).
		Update("status", body.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить заявку"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Статус изменён", "status": body.Status})
}

// buildPageNumbers строит компактный список страниц:
// всегда показываем 1,2 ... (окно вокруг текущей) ... N-1,N
// троеточия помечаем как -1
func buildPageNumbers(current, total int) []int {
	if total <= 7 {
		out := make([]int, total)
		for i := 0; i < total; i++ {
			out[i] = i + 1
		}
		return out
	}

	var res []int
	// первые две
	res = append(res, 1, 2)

	// левое троеточие
	if current > 4 {
		res = append(res, -1)
	}

	// окно вокруг текущей (±1), внутри (3 .. total-2)
	start := current - 1
	if start < 3 {
		start = 3
	}
	end := current + 1
	if end > total-2 {
		end = total - 2
	}
	for i := start; i <= end; i++ {
		res = append(res, i)
	}

	// правое троеточие
	if current < total-3 {
		res = append(res, -1)
	}

	// последние две
	res = append(res, total-1, total)
	return res
}

// BulkUpdateContacts — массовая смена статуса: processed | new | archived
func (h *Handlers) BulkUpdateContacts(c *gin.Context) {
	var req struct {
		Action string `json:"action"`
		IDs    []uint `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Неверные данные"})
		return
	}
	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Не выбрано ни одной заявки"})
		return
	}

	switch req.Action {
	case "processed", "new", "archived":
		// ok
	default:
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Недопустимое действие"})
		return
	}

	// Выполним массовый апдейт
	if err := h.db.Model(&models.ContactForm{}).
		Where("id IN ?", req.IDs).
		Update("status", req.Action).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Не удалось обновить заявки"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"action":  req.Action,
		"ids":     req.IDs,
	})
}

// AdminContactsExportCSV — экспорт заявок в CSV с учётом фильтров (UTF-8 + BOM)
func (h *Handlers) AdminContactsExportCSV(c *gin.Context) {
	var contacts []models.ContactForm

	// --- Базовый запрос как на странице ---
	qb := h.db.Model(&models.ContactForm{})

	// Поиск: q или search
	q := c.Query("q")
	if q == "" {
		q = c.Query("search")
	}
	if q != "" {
		like := "%" + q + "%"
		qb = qb.Where("name ILIKE ? OR phone ILIKE ? OR email ILIKE ? OR message ILIKE ?", like, like, like, like)
	}

	// Статус
	status := c.Query("status")
	if status != "" {
		qb = qb.Where("status = ?", status)
	}

	// Даты
	now := nowMoscow()

	switch c.Query("date") {
	case "today":
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, moscowLoc)
		end := start.Add(24 * time.Hour)
		qb = qb.Where("created_at >= ? AND created_at < ?", start.UTC(), end.UTC())
	case "7d":
		from := now.AddDate(0, 0, -7)
		qb = qb.Where("created_at >= ?", from.UTC())
	case "month":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, moscowLoc)
		end := start.AddDate(0, 1, 0)
		qb = qb.Where("created_at >= ? AND created_at < ?", start.UTC(), end.UTC())
	}

	// Вытащим все подходящие (без пагинации), отсортированные по дате
	if err := qb.Order("created_at DESC").Find(&contacts).Error; err != nil {
		c.String(http.StatusInternalServerError, "DB error")
		return
	}

	// --- Заголовки ответа ---
	filename := "contacts_export_" + time.Now().Format("20060102_150405") + ".csv"
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)

	// --- Пишем BOM, чтобы Excel открыл UTF-8 корректно ---
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(c.Writer)
	w.Comma = ';' // точка с запятой — безопаснее из-за запятых в тексте

	// Шапка
	_ = w.Write([]string{
		"Имя", "Телефон", "Email", "Компания", "Тип проекта", "Сообщение", "Статус", "Дата",
	})

	// Данные
	for _, cf := range contacts {
		row := []string{
			cf.Name,
			cf.Phone,
			cf.Email,
			cf.Company,
			cf.ProjectType,
			cf.Message,
			cf.Status,
			cf.CreatedAt.In(moscowLoc).Format("02.01.2006 15:04"),
		}
		_ = w.Write(row)
	}

	w.Flush()
	if err := w.Error(); err != nil {
		// если вдруг ошибка записи — отдадим 500
		c.Error(err)
	}
}
