package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// ---------- mustID ----------

func TestMustID_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test/:id", func(c *gin.Context) {
		id, ok := mustID(c)
		if ok {
			c.JSON(http.StatusOK, gin.H{"id": id})
		}
	})

	req, _ := http.NewRequest("GET", "/test/42", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(42), resp["id"])
}

func TestMustID_Zero(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test/:id", func(c *gin.Context) {
		mustID(c)
	})

	req, _ := http.NewRequest("GET", "/test/0", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMustID_NonNumeric(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test/:id", func(c *gin.Context) {
		mustID(c)
	})

	req, _ := http.NewRequest("GET", "/test/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMustID_Negative(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test/:id", func(c *gin.Context) {
		mustID(c)
	})

	req, _ := http.NewRequest("GET", "/test/-1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------- parseStatus ----------

func TestParseStatus_Valid(t *testing.T) {
	tests := []string{"new", "processed", "archived"}
	for _, s := range tests {
		result, ok := parseStatus(s)
		assert.True(t, ok, "статус %q должен быть валидным", s)
		assert.Equal(t, s, result)
	}
}

func TestParseStatus_Invalid(t *testing.T) {
	tests := []string{"", "invalid", "deleted", "NEW", "Processed"}
	for _, s := range tests {
		_, ok := parseStatus(s)
		assert.False(t, ok, "статус %q не должен быть валидным", s)
	}
}

// ---------- buildPageNumbers ----------

func TestBuildPageNumbers_SmallTotal(t *testing.T) {
	// <= 7 страниц — показываем все
	result := buildPageNumbers(1, 5)
	assert.Equal(t, []int{1, 2, 3, 4, 5}, result)
}

func TestBuildPageNumbers_ExactlySeven(t *testing.T) {
	result := buildPageNumbers(4, 7)
	assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 7}, result)
}

func TestBuildPageNumbers_SinglePage(t *testing.T) {
	result := buildPageNumbers(1, 1)
	assert.Equal(t, []int{1}, result)
}

func TestBuildPageNumbers_LargeTotal_CurrentAtStart(t *testing.T) {
	result := buildPageNumbers(1, 20)
	// Начало: 1, 2, (нет троеточия т.к. current<=4), 3... потом троеточие, 19, 20
	assert.Equal(t, 1, result[0])
	assert.Equal(t, 2, result[1])
	assert.Equal(t, 20, result[len(result)-1])
	assert.Equal(t, 19, result[len(result)-2])
	// Должно содержать троеточие (-1)
	assert.Contains(t, result, -1)
}

func TestBuildPageNumbers_LargeTotal_CurrentInMiddle(t *testing.T) {
	result := buildPageNumbers(10, 20)
	// Должно: 1, 2, -1, 9, 10, 11, -1, 19, 20
	assert.Equal(t, 1, result[0])
	assert.Equal(t, 20, result[len(result)-1])
	assert.Contains(t, result, 10) // текущая страница
	assert.Contains(t, result, 9)  // prev
	assert.Contains(t, result, 11) // next

	// Два троеточия
	count := 0
	for _, v := range result {
		if v == -1 {
			count++
		}
	}
	assert.Equal(t, 2, count)
}

func TestBuildPageNumbers_LargeTotal_CurrentAtEnd(t *testing.T) {
	result := buildPageNumbers(20, 20)
	assert.Equal(t, 1, result[0])
	assert.Equal(t, 20, result[len(result)-1])
	assert.Contains(t, result, 19)
	// Одно троеточие (слева)
	count := 0
	for _, v := range result {
		if v == -1 {
			count++
		}
	}
	assert.Equal(t, 1, count)
}

// ---------- jsonOK ----------

func TestJsonOK_GinH(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		jsonOK(c, gin.H{"message": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, "ok", resp["message"])
}

func TestJsonOK_Struct(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	type result struct {
		Count int `json:"count"`
	}

	router.GET("/test", func(c *gin.Context) {
		jsonOK(c, result{Count: 5})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	// struct не добавляет success
	assert.Equal(t, float64(5), resp["count"])
	assert.Nil(t, resp["success"])
}

// ---------- jsonErr ----------

func TestJsonErr(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		jsonErr(c, http.StatusNotFound, "Не найдено")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Не найдено", resp["error"])
}

// ---------- pageMeta ----------

func TestPageMeta_FirstPage(t *testing.T) {
	_, h := setupTestRouter(t)
	pages, prev, next, numbers := h.pageMeta(100, 1, 50)
	assert.Equal(t, 2, pages)
	assert.Equal(t, 1, prev) // не может быть меньше 1
	assert.Equal(t, 2, next)
	assert.Equal(t, []int{1, 2}, numbers)
}

func TestPageMeta_LastPage(t *testing.T) {
	_, h := setupTestRouter(t)
	pages, prev, next, _ := h.pageMeta(100, 2, 50)
	assert.Equal(t, 2, pages)
	assert.Equal(t, 1, prev)
	assert.Equal(t, 2, next) // не может быть больше pages
}

func TestPageMeta_ZeroTotal(t *testing.T) {
	_, h := setupTestRouter(t)
	pages, prev, next, numbers := h.pageMeta(0, 1, 50)
	assert.Equal(t, 1, pages) // минимум 1
	assert.Equal(t, 1, prev)
	assert.Equal(t, 1, next)
	assert.Equal(t, []int{1}, numbers)
}

func TestPageMeta_LargeDataset(t *testing.T) {
	_, h := setupTestRouter(t)
	pages, _, _, _ := h.pageMeta(1000, 5, 25)
	assert.Equal(t, 40, pages)
}

// ---------- getPageQuery ----------

func TestGetPageQuery_Defaults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	_, h := setupTestRouter(t)

	router.GET("/test", func(c *gin.Context) {
		page, limit, offset := h.getPageQuery(c)
		c.JSON(http.StatusOK, gin.H{"page": page, "limit": limit, "offset": offset})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(1), resp["page"])
	assert.Equal(t, float64(50), resp["limit"])
	assert.Equal(t, float64(0), resp["offset"])
}

func TestGetPageQuery_CustomValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	_, h := setupTestRouter(t)

	router.GET("/test", func(c *gin.Context) {
		page, limit, offset := h.getPageQuery(c)
		c.JSON(http.StatusOK, gin.H{"page": page, "limit": limit, "offset": offset})
	})

	req, _ := http.NewRequest("GET", "/test?page=3&limit=25", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(3), resp["page"])
	assert.Equal(t, float64(25), resp["limit"])
	assert.Equal(t, float64(50), resp["offset"]) // (3-1)*25
}

func TestGetPageQuery_InvalidLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	_, h := setupTestRouter(t)

	router.GET("/test", func(c *gin.Context) {
		_, limit, _ := h.getPageQuery(c)
		c.JSON(http.StatusOK, gin.H{"limit": limit})
	})

	// limit=30 не в списке [25, 50, 100] — должен fallback на 50
	req, _ := http.NewRequest("GET", "/test?limit=30", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(50), resp["limit"])
}

func TestGetPageQuery_InvalidPage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	_, h := setupTestRouter(t)

	router.GET("/test", func(c *gin.Context) {
		page, _, _ := h.getPageQuery(c)
		c.JSON(http.StatusOK, gin.H{"page": page})
	})

	req, _ := http.NewRequest("GET", "/test?page=-1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(1), resp["page"]) // fallback на 1
}

// ---------- NowMSK ----------

func TestNowMSK_Timezone(t *testing.T) {
	now := NowMSK()
	assert.Equal(t, "Europe/Moscow", now.Location().String())
}

func TestNowMSKUTC_Timezone(t *testing.T) {
	now := NowMSKUTC()
	assert.Equal(t, "UTC", now.Location().String())
}
