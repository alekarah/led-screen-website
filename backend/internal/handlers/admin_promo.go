package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"ledsite/internal/models"

	"github.com/gin-gonic/gin"
)

// AdminPromoPage отображает страницу управления popup акцией
func (h *Handlers) AdminPromoPage(c *gin.Context) {
	var promo models.PromoPopup
	if err := h.db.First(&promo).Error; err != nil {
		// Создаем запись если её нет
		promo = models.PromoPopup{
			Title:            "",
			Content:          "",
			IsActive:         false,
			Pages:            `["home"]`,
			TTLHours:         24,
			ShowDelaySeconds: 0,
		}
		h.db.Create(&promo)
	}

	// Парсим JSON массив страниц для шаблона
	var pages []string
	if err := json.Unmarshal([]byte(promo.Pages), &pages); err != nil {
		pages = []string{"home"}
	}

	c.HTML(http.StatusOK, "admin_base.html", gin.H{
		"PageID":    "admin-promo",
		"title":     "Управление акцией",
		"promo":     promo,
		"pages":     pages,
		"allPages":  getAllPages(),
		"csrfToken": c.GetString("csrf_token"),
	})
}

// AdminPromoUpdate обновляет настройки popup акции
func (h *Handlers) AdminPromoUpdate(c *gin.Context) {
	var promo models.PromoPopup
	if err := h.db.First(&promo).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Promo popup not found"})
		return
	}

	// Получаем данные из формы
	promo.Title = c.PostForm("title")
	promo.Content = c.PostForm("content")
	promo.IsActive = c.PostForm("is_active") == "on" || c.PostForm("is_active") == "true"

	// TTL
	ttl := 24
	if v, err := strconv.Atoi(c.PostForm("ttl_hours")); err == nil && v >= 0 {
		ttl = v
	}
	promo.TTLHours = ttl

	// Delay
	delay := 0
	if v, err := strconv.Atoi(c.PostForm("show_delay_seconds")); err == nil && v >= 0 {
		delay = v
	}
	promo.ShowDelaySeconds = delay

	// Pages - получаем массив чекбоксов
	selectedPages := c.PostFormArray("pages[]")
	if len(selectedPages) == 0 {
		selectedPages = []string{"home"}
	}
	pagesJSON, err := json.Marshal(selectedPages)
	if err != nil {
		pagesJSON = []byte(`["home"]`)
	}
	promo.Pages = string(pagesJSON)

	if err := h.db.Save(&promo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update promo popup"})
		return
	}

	// Редирект обратно на страницу с сообщением об успехе
	c.Redirect(http.StatusFound, "/admin/promo?success=1")
}

// GetActivePromo возвращает активный popup для публичной части (API)
func (h *Handlers) GetActivePromo(c *gin.Context) {
	pageID := c.Query("page")
	if pageID == "" {
		pageID = "home"
	}

	var promo models.PromoPopup
	if err := h.db.First(&promo).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"active": false})
		return
	}

	if !promo.IsActive {
		c.JSON(http.StatusOK, gin.H{"active": false})
		return
	}

	// Проверяем, включена ли текущая страница
	var pages []string
	if err := json.Unmarshal([]byte(promo.Pages), &pages); err != nil {
		c.JSON(http.StatusOK, gin.H{"active": false})
		return
	}

	pageEnabled := false
	for _, p := range pages {
		if p == pageID {
			pageEnabled = true
			break
		}
	}

	if !pageEnabled {
		c.JSON(http.StatusOK, gin.H{"active": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"active":             true,
		"title":              promo.Title,
		"content":            promo.Content,
		"ttl_hours":          promo.TTLHours,
		"show_delay_seconds": promo.ShowDelaySeconds,
	})
}

// PageOption представляет опцию страницы для выбора в админке
type PageOption struct {
	ID   string
	Name string
}

// getAllPages возвращает список всех доступных страниц для popup
func getAllPages() []PageOption {
	return []PageOption{
		{ID: "home", Name: "Главная"},
		{ID: "prices", Name: "Цены"},
		{ID: "projects", Name: "Портфолио"},
		{ID: "services", Name: "Услуги"},
		{ID: "led-guide", Name: "О LED-экранах"},
		{ID: "contact", Name: "Контакты"},
	}
}
