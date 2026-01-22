package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"ledsite/internal/models"

	"github.com/gin-gonic/gin"
)

// getBaseURL определяет базовый URL с правильным протоколом (http/https).
// Учитывает работу за reverse proxy (nginx) через заголовки X-Forwarded-Proto и X-Forwarded-Host.
func getBaseURL(c *gin.Context) string {
	// Production домен всегда использует HTTPS
	host := c.Request.Host
	if host == "s-n-r.ru" || host == "www.s-n-r.ru" {
		return "https://s-n-r.ru"
	}

	// Проверяем заголовки от reverse proxy (nginx)
	proto := c.GetHeader("X-Forwarded-Proto")
	if proto == "" {
		// Если нет заголовка от proxy, проверяем TLS
		if c.Request.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}

	// Используем X-Forwarded-Host если есть, иначе Host
	forwardedHost := c.GetHeader("X-Forwarded-Host")
	if forwardedHost != "" {
		host = forwardedHost
	}

	// Убираем www. если есть
	host = strings.TrimPrefix(host, "www.")

	return proto + "://" + host
}

// Sitemap генерирует XML карту сайта для поисковых систем.
//
// Включает:
//   - Статические страницы (главная, услуги, контакты и т.д.)
//   - Динамические страницы проектов из БД
//
// Формат соответствует спецификации sitemap.org
// Частота обновления и приоритет настроены для оптимальной индексации.
//
// GET /sitemap.xml
func (h *Handlers) Sitemap(c *gin.Context) {
	// Определяем базовый URL с учетом SSL и reverse proxy
	baseURL := getBaseURL(c)

	// Получаем все проекты для динамических URL
	var projects []models.Project
	h.db.Select("id, slug, updated_at").Order("updated_at DESC").Find(&projects)

	// Текущая дата для lastmod
	now := time.Now().Format("2006-01-02")

	// Начало XML
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
`

	// Статические страницы
	staticPages := []struct {
		loc        string
		changefreq string
		priority   string
	}{
		{baseURL + "/", "weekly", "1.0"},                   // Главная - максимальный приоритет
		{baseURL + "/projects", "weekly", "0.9"},           // Портфолио
		{baseURL + "/prices", "weekly", "0.9"},             // Цены
		{baseURL + "/led-screens-guide", "monthly", "0.8"}, // О LED-экранах (информационная страница)
		{baseURL + "/services", "monthly", "0.8"},          // Услуги
		{baseURL + "/contact", "monthly", "0.7"},           // Контакты
		{baseURL + "/privacy", "yearly", "0.3"},            // Политика конфиденциальности
	}

	for _, page := range staticPages {
		xml += fmt.Sprintf(`  <url>
    <loc>%s</loc>
    <lastmod>%s</lastmod>
    <changefreq>%s</changefreq>
    <priority>%s</priority>
  </url>
`, page.loc, now, page.changefreq, page.priority)
	}

	// Динамические страницы проектов
	for _, project := range projects {
		lastmod := project.UpdatedAt.Format("2006-01-02")
		xml += fmt.Sprintf(`  <url>
    <loc>%s/projects/%s</loc>
    <lastmod>%s</lastmod>
    <changefreq>monthly</changefreq>
    <priority>0.6</priority>
  </url>
`, baseURL, project.Slug, lastmod)
	}

	// Закрываем XML
	xml += `</urlset>
`

	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.String(http.StatusOK, xml)
}

// RobotsTxt генерирует файл robots.txt с правилами для поисковых ботов.
//
// Включает:
//   - Разрешения для всех ботов
//   - Запрет на индексацию админ-панели
//   - Ссылку на sitemap.xml
//
// GET /robots.txt
func (h *Handlers) RobotsTxt(c *gin.Context) {
	// Определяем базовый URL с учетом SSL и reverse proxy
	baseURL := getBaseURL(c)

	robots := fmt.Sprintf(`# robots.txt для s-n-r.ru
# LED экраны в Санкт-Петербурге

# Разрешаем всем поисковым ботам индексировать сайт
User-agent: *
Allow: /

# Запрещаем индексацию админ-панели
Disallow: /admin/
Disallow: /api/admin/

# Запрещаем индексацию служебных файлов
Disallow: /static/uploads/

# Карта сайта для поисковых систем
Sitemap: %s/sitemap.xml

# Специальные правила для Яндекса
User-agent: Yandex
Allow: /

# Специальные правила для Google
User-agent: Googlebot
Allow: /
`, baseURL)

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, robots)
}
