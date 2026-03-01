package handlers

import (
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
)

// formatPhone форматирует номер +79991234567 → +7 999 123 45 67
func formatPhone(phone string) string {
	digits := regexp.MustCompile(`\D`).ReplaceAllString(phone, "")
	if len(digits) == 11 && digits[0] == '7' {
		return "+" + digits[0:1] + " " + digits[1:4] + " " + digits[4:7] + " " + digits[7:9] + " " + digits[9:11]
	}
	return phone
}

// AdminSettingsPage отображает страницу настроек сайта (телефон, email)
func (h *Handlers) AdminSettingsPage(c *gin.Context) {
	settings := getSettings(h.db)

	c.HTML(http.StatusOK, "admin_base.html", gin.H{
		"PageID":    "admin-settings",
		"title":     "Настройки сайта",
		"settings":  settings,
		"csrfToken": c.GetString("csrf_token"),
	})
}

// AdminSettingsUpdate сохраняет настройки сайта
func (h *Handlers) AdminSettingsUpdate(c *gin.Context) {
	settings := getSettings(h.db)

	settings.Phone = c.PostForm("phone")
	settings.PhoneDisplay = formatPhone(settings.Phone)
	settings.PhoneNote = c.PostForm("phone_note")
	settings.Email = c.PostForm("email")
	settings.EmailNote = c.PostForm("email_note")
	settings.Address = c.PostForm("address")
	settings.AddressNote = c.PostForm("address_note")
	settings.WorkHours = c.PostForm("work_hours")
	settings.WorkHoursNote = c.PostForm("work_hours_note")
	statsProjects, _ := strconv.Atoi(c.PostForm("stats_projects"))
	settings.StatsProjects = statsProjects
	statsYears, _ := strconv.Atoi(c.PostForm("stats_years"))
	settings.StatsYears = statsYears

	if err := h.db.Save(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения настроек"})
		return
	}

	c.Redirect(http.StatusFound, "/admin/settings?success=1")
}
