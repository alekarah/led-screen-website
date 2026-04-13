package handlers

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/text/encoding/charmap"
	"gorm.io/gorm"

	"ledsite/internal/models"
)

// cbValute — структура для парсинга XML ответа ЦБ РФ
type cbValute struct {
	CharCode string `xml:"CharCode"`
	Value    string `xml:"Value"`
}

// cbValCurs — корневой элемент XML ответа ЦБ РФ
type cbValCurs struct {
	Valutes []cbValute `xml:"Valute"`
}

// fetchUSDRateFromCB запрашивает курс доллара с сайта ЦБ РФ.
func fetchUSDRateFromCB() (float64, error) {
	return fetchUSDRateFromURL("https://www.cbr.ru/scripts/XML_daily.asp")
}

// fetchUSDRateFromURL запрашивает курс доллара с указанного URL (используется в тестах).
func fetchUSDRateFromURL(url string) (float64, error) {
	req, reqErr := http.NewRequest("GET", url, nil)
	if reqErr != nil {
		return 0, fmt.Errorf("cbr new request: %w", reqErr)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ledsite/1.0)")

	client := &http.Client{}
	resp, httpErr := client.Do(req)
	if httpErr != nil {
		return 0, fmt.Errorf("cbr request: %w", httpErr)
	}
	defer resp.Body.Close()

	// ЦБ отдаёт XML в windows-1251, читаем сырые байты
	raw, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return 0, fmt.Errorf("cbr read body: %w", readErr)
	}

	// Убираем XML-декларацию (<?xml...?>) ДО декодирования —
	// после декодирования парсер будет читать без объявления кодировки
	if idx := bytes.Index(raw, []byte("?>")); idx != -1 {
		raw = raw[idx+2:]
	}

	// Декодируем оставшийся контент windows-1251 → UTF-8
	decoder := charmap.Windows1251.NewDecoder()
	utf8Body, decodeErr := io.ReadAll(decoder.Reader(bytes.NewReader(raw)))
	if decodeErr != nil {
		return 0, fmt.Errorf("cbr decode: %w", decodeErr)
	}

	var curs cbValCurs
	if parseErr := xml.Unmarshal(utf8Body, &curs); parseErr != nil {
		return 0, fmt.Errorf("cbr xml parse: %w", parseErr)
	}

	for _, v := range curs.Valutes {
		if v.CharCode == "USD" {
			// ЦБ использует запятую как десятичный разделитель
			var rate float64
			if _, scanErr := fmt.Sscanf(replaceComma(v.Value), "%f", &rate); scanErr != nil {
				return 0, fmt.Errorf("cbr parse rate: %w", scanErr)
			}
			return rate, nil
		}
	}
	return 0, fmt.Errorf("USD not found in CBR response")
}

// replaceComma заменяет запятую на точку — нужно для парсинга чисел от ЦБ РФ.
func replaceComma(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			result[i] = '.'
		} else {
			result[i] = s[i]
		}
	}
	return string(result)
}

// getOrRefreshUSDRate возвращает актуальный курс доллара из кэша или обновляет его.
// Кэш обновляется раз в час.
func getOrRefreshUSDRate(db *gorm.DB) (float64, error) {
	var settings models.CalculatorSettings
	if dbErr := db.First(&settings).Error; dbErr != nil {
		return 0, fmt.Errorf("getOrRefreshUSDRate db: %w", dbErr)
	}

	// Если курс свежее 24 часов — возвращаем кэш
	if settings.UsdRate > 0 && time.Since(settings.UsdRateAt) < 24*time.Hour {
		return settings.UsdRate * (1 + settings.UsdMarkupPct/100), nil
	}

	// Обновляем курс из ЦБ
	rate, fetchErr := fetchUSDRateFromCB()
	if fetchErr != nil {
		// Если не удалось получить — используем кэш если есть
		if settings.UsdRate > 0 {
			return settings.UsdRate * (1 + settings.UsdMarkupPct/100), nil
		}
		return 0, fmt.Errorf("fetchUSDRateFromCB: %w", fetchErr)
	}

	db.Model(&settings).Updates(map[string]interface{}{
		"usd_rate":    rate,
		"usd_rate_at": time.Now(),
	})

	return rate * (1 + settings.UsdMarkupPct/100), nil
}

// GetCalculatorData возвращает данные для калькулятора в JSON.
// Используется фронтендом для пересчёта при изменении параметров.
//
// GET /api/calculator
func (h *Handlers) GetCalculatorData(c *gin.Context) {
	var settings models.CalculatorSettings
	if dbErr := h.db.First(&settings).Error; dbErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "calculator settings not found"})
		return
	}

	var pitches []models.CalculatorPixelPitch
	h.db.Where("is_active = ?", true).Order("screen_type, sort_order").Find(&pitches)

	rate, rateErr := getOrRefreshUSDRate(h.db)
	if rateErr != nil {
		rate = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"usd_rate": rate,
		"settings": settings,
		"pitches":  pitches,
	})
}

// getCalculatorTemplateData возвращает данные калькулятора для рендера шаблона публичной страницы цен.
// Включает настройки, шаги пикселя по типу экрана и итоговый курс доллара с надбавкой.
func getCalculatorTemplateData(db *gorm.DB) map[string]interface{} {
	var settings models.CalculatorSettings
	db.First(&settings)

	var indoorPitches []models.CalculatorPixelPitch
	db.Where("screen_type = ? AND is_active = ?", "indoor", true).Order("sort_order").Find(&indoorPitches)

	var outdoorPitches []models.CalculatorPixelPitch
	db.Where("screen_type = ? AND is_active = ?", "outdoor", true).Order("sort_order").Find(&outdoorPitches)

	rate, rateErr := getOrRefreshUSDRate(db)
	if rateErr != nil {
		rate = 0
	}

	return map[string]interface{}{
		"calcSettings":   settings,
		"indoorPitches":  indoorPitches,
		"outdoorPitches": outdoorPitches,
		"usdRate":        rate,
	}
}

// AdminCalculatorPage отображает страницу настроек калькулятора в админке.
//
// GET /admin/calculator
func (h *Handlers) AdminCalculatorPage(c *gin.Context) {
	var settings models.CalculatorSettings
	h.db.First(&settings)

	var pitches []models.CalculatorPixelPitch
	h.db.Order("screen_type, sort_order").Find(&pitches)

	c.HTML(http.StatusOK, "admin_base.html", gin.H{
		"title":    "Калькулятор",
		"PageID":   "admin-calculator",
		"settings": settings,
		"pitches":  pitches,
	})
}

// AdminCalculatorUpdateSettings сохраняет константы калькулятора.
//
// POST /admin/calculator/settings
func (h *Handlers) AdminCalculatorUpdateSettings(c *gin.Context) {
	var settings models.CalculatorSettings
	if dbErr := h.db.First(&settings).Error; dbErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "not found"})
		return
	}

	if bindErr := c.ShouldBindJSON(&settings); bindErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": bindErr.Error()})
		return
	}

	if saveErr := h.db.Save(&settings).Error; saveErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// AdminCalculatorCreatePitch создаёт новый шаг пикселя.
//
// POST /admin/calculator/pitches
func (h *Handlers) AdminCalculatorCreatePitch(c *gin.Context) {
	var pitch models.CalculatorPixelPitch
	if bindErr := c.ShouldBindJSON(&pitch); bindErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": bindErr.Error()})
		return
	}
	if dbErr := h.db.Create(&pitch).Error; dbErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create failed"})
		return
	}
	c.JSON(http.StatusOK, pitch)
}

// AdminCalculatorUpdatePitch обновляет шаг пикселя.
//
// POST /admin/calculator/pitches/:id
func (h *Handlers) AdminCalculatorUpdatePitch(c *gin.Context) {
	id := c.Param("id")
	var pitch models.CalculatorPixelPitch
	if dbErr := h.db.First(&pitch, id).Error; dbErr != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if bindErr := c.ShouldBindJSON(&pitch); bindErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": bindErr.Error()})
		return
	}
	if saveErr := h.db.Save(&pitch).Error; saveErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save failed"})
		return
	}
	c.JSON(http.StatusOK, pitch)
}

// AdminCalculatorDeletePitch удаляет шаг пикселя.
//
// DELETE /admin/calculator/pitches/:id
func (h *Handlers) AdminCalculatorDeletePitch(c *gin.Context) {
	id := c.Param("id")
	if dbErr := h.db.Delete(&models.CalculatorPixelPitch{}, id).Error; dbErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
