package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSitemap проверяет генерацию sitemap.xml
func TestSitemap(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/sitemap.xml", h.Sitemap)

	req, _ := http.NewRequest("GET", "/sitemap.xml", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/xml; charset=utf-8", w.Header().Get("Content-Type"))

	body := w.Body.String()

	// Проверяем структуру XML
	assert.Contains(t, body, `<?xml version="1.0" encoding="UTF-8"?>`)
	assert.Contains(t, body, `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)
	assert.Contains(t, body, `</urlset>`)

	// Проверяем наличие основных страниц (http или https)
	assert.True(t, strings.Contains(body, `<loc>http://`) || strings.Contains(body, `<loc>https://`), "Should contain http:// or https://")
	assert.Contains(t, body, `/</loc>`) // Главная
	assert.Contains(t, body, `/projects</loc>`)
	assert.Contains(t, body, `/services</loc>`)
	assert.Contains(t, body, `/contact</loc>`)
	assert.Contains(t, body, `/privacy</loc>`)

	// Проверяем наличие обязательных тегов
	assert.Contains(t, body, `<lastmod>`)
	assert.Contains(t, body, `<changefreq>`)
	assert.Contains(t, body, `<priority>`)
}

// TestSitemap_Priority проверяет правильность приоритетов
func TestSitemap_Priority(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/sitemap.xml", h.Sitemap)

	req, _ := http.NewRequest("GET", "/sitemap.xml", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	body := w.Body.String()

	// Главная страница должна иметь приоритет 1.0
	assert.Contains(t, body, "<priority>1.0</priority>")

	// Портфолио - 0.9
	assert.Contains(t, body, "<priority>0.9</priority>")

	// Политика конфиденциальности - самый низкий приоритет
	assert.Contains(t, body, "<priority>0.3</priority>")
}

// TestRobotsTxt проверяет генерацию robots.txt
func TestRobotsTxt(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/robots.txt", h.RobotsTxt)

	req, _ := http.NewRequest("GET", "/robots.txt", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))

	body := w.Body.String()

	// Проверяем основные директивы
	assert.Contains(t, body, "User-agent: *")
	assert.Contains(t, body, "Allow: /")

	// Проверяем блокировку админки
	assert.Contains(t, body, "Disallow: /admin/")
	assert.Contains(t, body, "Disallow: /api/admin/")
	assert.Contains(t, body, "Disallow: /static/uploads/")

	// Проверяем ссылку на sitemap
	assert.Contains(t, body, "Sitemap:")
	assert.Contains(t, body, "/sitemap.xml")

	// Проверяем правила для Яндекс и Google
	assert.Contains(t, body, "User-agent: Yandex")
	assert.Contains(t, body, "User-agent: Googlebot")
}

// TestRobotsTxt_Format проверяет формат robots.txt
func TestRobotsTxt_Format(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/robots.txt", h.RobotsTxt)

	req, _ := http.NewRequest("GET", "/robots.txt", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	body := w.Body.String()
	lines := strings.Split(body, "\n")

	// Должно быть больше 10 строк
	assert.Greater(t, len(lines), 10)

	// Не должно быть пустых User-agent директив
	for i, line := range lines {
		if strings.HasPrefix(line, "User-agent:") {
			// После User-agent должна быть директива Allow или Disallow
			if i+1 < len(lines) {
				nextLine := lines[i+1]
				assert.True(t,
					strings.HasPrefix(nextLine, "Allow:") ||
						strings.HasPrefix(nextLine, "Disallow:") ||
						nextLine == "",
					"После User-agent должна быть директива или пустая строка")
			}
		}
	}
}

// TestSitemap_HTTPS проверяет что production домен использует HTTPS
func TestSitemap_HTTPS(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/sitemap.xml", h.Sitemap)

	// Эмулируем запрос к production домену
	req, _ := http.NewRequest("GET", "/sitemap.xml", nil)
	req.Host = "s-n-r.ru"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	body := w.Body.String()

	// Production домен должен использовать HTTPS
	assert.Contains(t, body, "<loc>https://s-n-r.ru/</loc>")
	assert.NotContains(t, body, "<loc>http://s-n-r.ru/</loc>")
}

// TestSitemap_ForwardedProto проверяет поддержку X-Forwarded-Proto заголовка
func TestSitemap_ForwardedProto(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/sitemap.xml", h.Sitemap)

	// Эмулируем запрос через nginx reverse proxy с SSL
	req, _ := http.NewRequest("GET", "/sitemap.xml", nil)
	req.Host = "example.com"
	req.Header.Set("X-Forwarded-Proto", "https")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	body := w.Body.String()

	// Должен использовать HTTPS из заголовка
	assert.Contains(t, body, "<loc>https://example.com/</loc>")
}

// TestRobotsTxt_HTTPS проверяет что robots.txt использует HTTPS для sitemap
func TestRobotsTxt_HTTPS(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/robots.txt", h.RobotsTxt)

	// Эмулируем запрос к production домену
	req, _ := http.NewRequest("GET", "/robots.txt", nil)
	req.Host = "s-n-r.ru"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	body := w.Body.String()

	// Sitemap должен быть по HTTPS
	assert.Contains(t, body, "Sitemap: https://s-n-r.ru/sitemap.xml")
	assert.NotContains(t, body, "Sitemap: http://s-n-r.ru/sitemap.xml")
}
