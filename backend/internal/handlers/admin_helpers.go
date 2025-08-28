package handlers

import (
	"net/http"
	"strconv"
	"time"

	"ledsite/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var moscowLoc, _ = time.LoadLocation("Europe/Moscow")

func NowMSK() time.Time {
	return time.Now().In(moscowLoc)
}

// applyDateFilter — применяет фильтр дат к запросу
func applyDateFilter(qb *gorm.DB, dateRange string) *gorm.DB {
	if dateRange == "" {
		return qb
	}
	now := NowMSK()
	switch dateRange {
	case "today":
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, moscowLoc)
		return qb.Where("created_at >= ? AND created_at < ?", start.UTC(), start.Add(24*time.Hour).UTC())
	case "7d":
		return qb.Where("created_at >= ?", now.AddDate(0, 0, -7).UTC())
	case "month":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, moscowLoc)
		return qb.Where("created_at >= ? AND created_at < ?", start.UTC(), start.AddDate(0, 1, 0).UTC())
	default:
		return qb
	}
}

// buildPageNumbers строит компактный список страниц.
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
	res = append(res, 1, 2)
	if current > 4 {
		res = append(res, -1)
	}
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
	if current < total-3 {
		res = append(res, -1)
	}
	res = append(res, total-1, total)
	return res
}

// Единые JSON-ответы
func jsonOK(c *gin.Context, payload any) {
	c.JSON(http.StatusOK, payload)
}

func jsonErr(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{"error": msg})
}

// Парсинг :id из URL с валидацией
func mustID(c *gin.Context) (uint, bool) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		jsonErr(c, http.StatusBadRequest, "Некорректный id")
		return 0, false
	}
	return uint(id64), true
}

// Валидация статуса
func parseStatus(s string) (string, bool) {
	switch s {
	case "new", "processed", "archived":
		return s, true
	default:
		return "", false
	}
}

// Базовый запрос: поиск + фильтры по датам
func (h *Handlers) baseContactsQB(c *gin.Context) *gorm.DB {
	qb := h.db.Model(&models.ContactForm{})

	// поиск
	search := c.Query("search")
	if search == "" {
		search = c.Query("q")
	}
	if search != "" {
		q := "%" + search + "%"
		qb = qb.Where("name ILIKE ? OR phone ILIKE ? OR email ILIKE ? OR message ILIKE ?", q, q, q, q)
	}

	// даты
	qb = applyDateFilter(qb, c.Query("date"))
	return qb

}

// getPageQuery — читает page/limit из query и возвращает page, limit, offset
func (h *Handlers) getPageQuery(c *gin.Context) (int, int, int) {
	page := 1
	if v, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && v > 0 {
		page = v
	}
	limit := 50
	if v, err := strconv.Atoi(c.DefaultQuery("limit", "50")); err == nil {
		switch v {
		case 25, 50, 100:
			limit = v
		}
	}
	offset := (page - 1) * limit
	return page, limit, offset
}

// pageMeta — считает pages/prev/next + номера страниц
func (h *Handlers) pageMeta(total int64, page, limit int) (int, int, int, []int) {
	pages := int((total + int64(limit) - 1) / int64(limit))
	if pages < 1 {
		pages = 1
	}
	prevPage := page - 1
	if prevPage < 1 {
		prevPage = 1
	}
	nextPage := page + 1
	if nextPage > pages {
		nextPage = pages
	}
	return pages, prevPage, nextPage, buildPageNumbers(page, pages)
}
