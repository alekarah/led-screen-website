package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"ledsite/internal/models"

	"github.com/gin-gonic/gin"
)

// coordsDelta — допуск при сравнении координат (~11 метров)
const coordsDelta = 0.0001

// isDuplicatePoint проверяет, существует ли точка с такими же координатами
func (h *Handlers) isDuplicatePoint(lat, lng float64) bool {
	var count int64
	h.db.Model(&models.MapPoint{}).
		Where("latitude BETWEEN ? AND ? AND longitude BETWEEN ? AND ?",
			lat-coordsDelta, lat+coordsDelta,
			lng-coordsDelta, lng+coordsDelta).
		Count(&count)
	return count > 0
}

// AdminMapPointsPage — страница управления точками на карте
func (h *Handlers) AdminMapPointsPage(c *gin.Context) {
	var points []models.MapPoint
	if err := h.db.Order("sort_order ASC, id DESC").Find(&points).Error; err != nil {
		log.Printf("Ошибка загрузки точек на карте: %v", err)
		c.HTML(http.StatusInternalServerError, "admin_base.html", gin.H{
			"PageID": "admin-error",
			"error":  "Ошибка загрузки данных",
		})
		return
	}

	c.HTML(http.StatusOK, "admin_base.html", gin.H{
		"title":     "Точки на карте",
		"PageID":    "admin-map-points",
		"MapPoints": points,
	})
}

// CreateMapPoint — создание новой точки на карте
func (h *Handlers) CreateMapPoint(c *gin.Context) {
	title := strings.TrimSpace(c.PostForm("title"))
	description := strings.TrimSpace(c.PostForm("description"))
	latStr := c.PostForm("latitude")
	lngStr := c.PostForm("longitude")
	panoramaURL := strings.TrimSpace(c.PostForm("panorama_url"))

	if title == "" {
		jsonErr(c, http.StatusBadRequest, "Название не может быть пустым")
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		jsonErr(c, http.StatusBadRequest, "Некорректная широта")
		return
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		jsonErr(c, http.StatusBadRequest, "Некорректная долгота")
		return
	}

	isActive := c.PostForm("is_active") != "false"

	// Проверка дубликата по координатам
	if h.isDuplicatePoint(lat, lng) {
		jsonErr(c, http.StatusConflict, "Точка с такими координатами уже существует")
		return
	}

	point := models.MapPoint{
		Title:       title,
		Description: description,
		Latitude:    lat,
		Longitude:   lng,
		PanoramaURL: panoramaURL,
		IsActive:    isActive,
	}

	if err := h.db.Create(&point).Error; err != nil {
		log.Printf("Ошибка создания точки '%s': %v", title, err)
		jsonErr(c, http.StatusInternalServerError, "Ошибка создания точки")
		return
	}

	jsonOK(c, gin.H{
		"message":   "Точка успешно создана",
		"map_point": point,
	})
}

// GetMapPoint — получение точки для редактирования (JSON)
func (h *Handlers) GetMapPoint(c *gin.Context) {
	id, ok := mustID(c)
	if !ok {
		return
	}

	var point models.MapPoint
	if err := h.db.First(&point, id).Error; err != nil {
		jsonErr(c, http.StatusNotFound, "Точка не найдена")
		return
	}

	jsonOK(c, gin.H{"map_point": point})
}

// UpdateMapPoint — обновление точки на карте
func (h *Handlers) UpdateMapPoint(c *gin.Context) {
	id, ok := mustID(c)
	if !ok {
		return
	}

	var point models.MapPoint
	if err := h.db.First(&point, id).Error; err != nil {
		jsonErr(c, http.StatusNotFound, "Точка не найдена")
		return
	}

	title := strings.TrimSpace(c.PostForm("title"))
	if title == "" {
		jsonErr(c, http.StatusBadRequest, "Название не может быть пустым")
		return
	}

	lat, err := strconv.ParseFloat(c.PostForm("latitude"), 64)
	if err != nil {
		jsonErr(c, http.StatusBadRequest, "Некорректная широта")
		return
	}

	lng, err := strconv.ParseFloat(c.PostForm("longitude"), 64)
	if err != nil {
		jsonErr(c, http.StatusBadRequest, "Некорректная долгота")
		return
	}

	point.Title = title
	point.Description = strings.TrimSpace(c.PostForm("description"))
	point.Latitude = lat
	point.Longitude = lng
	point.PanoramaURL = strings.TrimSpace(c.PostForm("panorama_url"))
	point.IsActive = c.PostForm("is_active") != "false"

	if err := h.db.Save(&point).Error; err != nil {
		log.Printf("Ошибка обновления точки ID=%d: %v", id, err)
		jsonErr(c, http.StatusInternalServerError, "Ошибка обновления точки")
		return
	}

	jsonOK(c, gin.H{"message": "Точка успешно обновлена"})
}

// DeleteMapPoint — удаление точки с карты
func (h *Handlers) DeleteMapPoint(c *gin.Context) {
	id, ok := mustID(c)
	if !ok {
		return
	}

	if err := h.db.Delete(&models.MapPoint{}, id).Error; err != nil {
		log.Printf("Ошибка удаления точки ID=%d: %v", id, err)
		jsonErr(c, http.StatusInternalServerError, "Ошибка удаления точки")
		return
	}

	jsonOK(c, gin.H{"message": "Точка успешно удалена"})
}

// UpdateMapPointsSorting — обновление порядка отображения точек (drag & drop)
func (h *Handlers) UpdateMapPointsSorting(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		jsonErr(c, http.StatusBadRequest, "Некорректные данные")
		return
	}

	for i, id := range req.IDs {
		if err := h.db.Model(&models.MapPoint{}).Where("id = ?", id).Update("sort_order", i).Error; err != nil {
			log.Printf("Ошибка обновления порядка для точки ID=%d: %v", id, err)
		}
	}

	jsonOK(c, gin.H{"message": "Порядок успешно обновлен"})
}

// BulkImportMapPoints — массовый импорт точек из ссылок Яндекс.Карт
func (h *Handlers) BulkImportMapPoints(c *gin.Context) {
	var req struct {
		Links []string `json:"links" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		jsonErr(c, http.StatusBadRequest, "Некорректные данные")
		return
	}

	var created int
	var errors []string

	for i, link := range req.Links {
		link = strings.TrimSpace(link)
		if link == "" {
			continue
		}

		lat, lng, err := parseCoordsFromYandexURL(link)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Строка %d: %v", i+1, err))
			continue
		}

		// Проверка дубликата по координатам
		if h.isDuplicatePoint(lat, lng) {
			errors = append(errors, fmt.Sprintf("Строка %d: точка с такими координатами уже существует", i+1))
			continue
		}

		// Извлекаем адрес из URL (между /house/ или последний сегмент пути)
		title := extractTitleFromURL(link, i+1)

		point := models.MapPoint{
			Title:       title,
			Latitude:    lat,
			Longitude:   lng,
			PanoramaURL: link,
			IsActive:    true,
		}

		if err := h.db.Create(&point).Error; err != nil {
			errors = append(errors, fmt.Sprintf("Строка %d: ошибка БД", i+1))
			continue
		}
		created++
	}

	msg := fmt.Sprintf("Импортировано %d точек", created)
	if len(errors) > 0 {
		msg += fmt.Sprintf(", ошибок: %d", len(errors))
	}

	jsonOK(c, gin.H{
		"message": msg,
		"created": created,
		"errors":  errors,
	})
}

// parseCoordsFromYandexURL извлекает координаты из ссылки Яндекс.Карт
// Формат параметра ll: ll=долгота,широта (lng,lat)
func parseCoordsFromYandexURL(rawURL string) (lat, lng float64, err error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return 0, 0, fmt.Errorf("некорректный URL")
	}

	llParam := u.Query().Get("ll")
	if llParam == "" {
		return 0, 0, fmt.Errorf("параметр ll не найден в URL")
	}

	parts := strings.Split(llParam, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("некорректный формат ll")
	}

	// В Яндекс.Картах ll = долгота,широта
	lng, err = strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("некорректная долгота")
	}

	lat, err = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("некорректная широта")
	}

	return lat, lng, nil
}

// extractTitleFromURL пытается извлечь читаемое название из URL Яндекс.Карт
func extractTitleFromURL(rawURL string, index int) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Sprintf("Точка %d", index)
	}

	// Ищем паттерн /house/название_улицы_номер/
	segments := strings.Split(u.Path, "/")
	for i, seg := range segments {
		if seg == "house" && i+1 < len(segments) {
			// Следующий сегмент — адрес (slug: varshavskaya_ulitsa_26)
			addr := segments[i+1]
			// Декодируем URL-encoded символы
			if decoded, err := url.PathUnescape(addr); err == nil {
				addr = decoded
			}
			// Заменяем _ на пробелы
			addr = strings.ReplaceAll(addr, "_", " ")
			return addr
		}
	}

	return fmt.Sprintf("Точка %d", index)
}
