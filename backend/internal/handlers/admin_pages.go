package handlers

import (
	"encoding/csv"
	"log"
	"net/http"
	"strings"
	"time"

	"ledsite/internal/models"

	"github.com/gin-gonic/gin"
)

var csvHeadersContacts = []string{
	"Имя", "Телефон", "Email", "Компания", "Тип проекта", "Сообщение", "Статус", "Дата",
}

// /admin/contacts — страница всех заявок (без архива)
func (h *Handlers) AdminContactsPage(c *gin.Context) {
	var contacts []models.ContactForm

	// Базовый запрос + выключаем архив
	qb := h.baseContactsQB(c).
		Where("archived_at IS NULL").
		Where("status <> ?", "archived")

	// Фильтр по статусу (только среди неархивных)
	status := c.Query("status")
	if status == "new" || status == "processed" {
		qb = qb.Where("status = ?", status)
	}

	reminder := strings.ToLower(strings.TrimSpace(c.Query("reminder")))
	now := NowMSK()
	startToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, moscowLoc)
	endToday := startToday.Add(24 * time.Hour)

	switch reminder {
	case "today":
		qb = qb.Where(
			"remind_flag = ? AND remind_at IS NOT NULL AND remind_at >= ? AND remind_at < ?",
			true, startToday.UTC(), endToday.UTC(),
		)
	case "overdue":
		qb = qb.Where(
			"remind_flag = ? AND remind_at IS NOT NULL AND remind_at < ?",
			true, now.UTC(),
		)
	case "upcoming":
		qb = qb.Where(
			"remind_flag = ? AND remind_at IS NOT NULL AND remind_at >= ?",
			true, now.UTC(),
		)
	}

	// Пагинация
	page, limit, offset := h.getPageQuery(c)

	// Подсчёт total
	var total int64
	if err := qb.Count(&total).Error; err != nil {
		c.String(http.StatusInternalServerError, "DB error")
		return
	}

	// Данные текущей страницы
	if err := qb.Order("created_at DESC").Limit(limit).Offset(offset).Find(&contacts).Error; err != nil {
		c.String(http.StatusInternalServerError, "DB error")
		return
	}

	// Метаданные пагинации
	pages, prevPage, nextPage, pageNumbers := h.pageMeta(total, page, limit)

	// Параметры для рендера
	search := c.Query("search")
	if search == "" {
		search = c.Query("q")
	}
	dateRange := c.Query("date")

	c.HTML(http.StatusOK, "admin_base.html", gin.H{
		"title":       "Заявки",
		"PageID":      "admin-contacts",
		"contactsAll": contacts,
		"total":       total,
		"page":        page,
		"pages":       pages,
		"prevPage":    prevPage,
		"nextPage":    nextPage,
		"pageNumbers": pageNumbers,
		"limit":       limit,
		"search":      search,
		"status":      status,
		"dateRange":   dateRange,
		"reminder":    reminder,
	})
}

// /admin/contacts/archive — архив заявок
func (h *Handlers) AdminContactsArchivePage(c *gin.Context) {
	var contacts []models.ContactForm

	// Базовый запрос + только архив
	qb := h.baseContactsQB(c).Where("archived_at IS NOT NULL")

	// Пагинация
	page, limit, offset := h.getPageQuery(c)

	// Подсчёт total
	var total int64
	if err := qb.Count(&total).Error; err != nil {
		c.String(http.StatusInternalServerError, "DB error")
		return
	}

	// Данные текущей страницы
	if err := qb.Order("created_at DESC").Limit(limit).Offset(offset).Find(&contacts).Error; err != nil {
		c.String(http.StatusInternalServerError, "DB error")
		return
	}

	// Метаданные пагинации
	pages, prevPage, nextPage, pageNumbers := h.pageMeta(total, page, limit)

	// Параметры для рендера (как у обычной страницы)
	search := c.Query("search")
	if search == "" {
		search = c.Query("q")
	}
	dateRange := c.Query("date")

	c.HTML(http.StatusOK, "admin_base.html", gin.H{
		"title":       "Архив заявок",
		"PageID":      "admin-contacts-archive",
		"contactsAll": contacts,
		"total":       total,
		"page":        page,
		"pages":       pages,
		"prevPage":    prevPage,
		"nextPage":    nextPage,
		"pageNumbers": pageNumbers,
		"limit":       limit,
		"search":      search,
		"status":      "archived",
		"dateRange":   dateRange,
	})
}

// Экспорт CSV (учитывает фильтры, UTF-8 + BOM)
func (h *Handlers) AdminContactsExportCSV(c *gin.Context) {
	var contacts []models.ContactForm

	// Базовый запрос как на страницах (поиск + даты)
	qb := h.baseContactsQB(c)

	// Архив / не архив + статус
	status := c.Query("status")
	if status == "archived" {
		qb = qb.Where("archived_at IS NOT NULL")
	} else {
		qb = qb.Where("archived_at IS NULL").
			Where("status <> ?", "archived")
		if status != "" {
			qb = qb.Where("status = ?", status)
		}
	}

	// Выборка всех подходящих без пагинации
	if err := qb.Order("created_at DESC").Find(&contacts).Error; err != nil {
		c.String(http.StatusInternalServerError, "DB error")
		return
	}

	// Заголовки и выдача CSV
	filename := "contacts_export_" + time.Now().Format("20060102_150405") + ".csv"
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)

	// UTF-8 BOM для корректного открытия в Excel
	if _, err := c.Writer.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		c.String(http.StatusInternalServerError, "Error writing BOM")
		return
	}
	w := csv.NewWriter(c.Writer)
	w.Comma = ';'
	if err := w.Write(csvHeadersContacts); err != nil {
		c.String(http.StatusInternalServerError, "Error writing CSV header")
		return
	}
	for _, cf := range contacts {
		if err := w.Write([]string{
			cf.Name, cf.Phone, cf.Email, cf.Company, cf.ProjectType, cf.Message, cf.Status,
			cf.CreatedAt.In(moscowLoc).Format("02.01.2006 15:04"),
		}); err != nil {
			c.String(http.StatusInternalServerError, "Error writing CSV row")
			return
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		log.Printf("CSV flush error: %v", err)
	}
}
