package handlers

import (
	"net/http"
	"strconv"

	"ledsite/internal/models"

	"github.com/gin-gonic/gin"
)

// ReorderProject - изменение порядка одного проекта
func (h *Handlers) ReorderProject(c *gin.Context) {
	id, ok := mustID(c)
	if !ok {
		return
	}

	positionStr := c.PostForm("position")
	position, err := strconv.Atoi(positionStr)
	if err != nil {
		jsonErr(c, http.StatusBadRequest, "Неверная позиция")
		return
	}

	var project models.Project
	if err := h.db.First(&project, id).Error; err != nil {
		jsonErr(c, http.StatusNotFound, "Проект не найден")
		return
	}

	project.SortOrder = position
	if err := h.db.Save(&project).Error; err != nil {
		jsonErr(c, http.StatusInternalServerError, "Ошибка обновления порядка: "+err.Error())
		return
	}
	jsonOK(c, gin.H{"message": "Порядок проекта обновлен"})
}

// BulkReorderProjects - массовое изменение порядка проектов
func (h *Handlers) BulkReorderProjects(c *gin.Context) {
	var requestData BulkReorderRequest

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверные данные запроса",
		})
		return
	}

	// Валидируем данные
	if len(requestData.Projects) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Пустой список проектов",
		})
		return
	}

	// Обновляем порядок всех проектов в транзакции
	if err := h.updateProjectsOrderInTransaction(requestData.Projects); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка обновления порядка проектов: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Порядок всех проектов обновлен",
	})
}

// ResetProjectOrder - сброс порядка проектов к алфавитному
func (h *Handlers) ResetProjectOrder(c *gin.Context) {
	var projects []models.Project
	h.db.Order("title ASC").Find(&projects)

	if len(projects) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "Нет проектов для сортировки",
		})
		return
	}

	// Присваиваем новые номера порядка в транзакции
	if err := h.resetProjectsOrderInTransaction(projects); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка сброса порядка: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Порядок проектов сброшен к алфавитному",
	})
}

// Вспомогательные типы и функции

// BulkReorderRequest структура для массового изменения порядка
type BulkReorderRequest struct {
	Projects []ProjectOrderData `json:"projects"`
}

// ProjectOrderData данные о порядке проекта
type ProjectOrderData struct {
	ID        int `json:"id"`
	SortOrder int `json:"sort_order"`
}

// updateProjectsOrderInTransaction обновляет порядок проектов в транзакции
func (h *Handlers) updateProjectsOrderInTransaction(projects []ProjectOrderData) error {
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, projectData := range projects {
		if err := tx.Model(&models.Project{}).
			Where("id = ?", projectData.ID).
			Update("sort_order", projectData.SortOrder).Error; err != nil {

			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// resetProjectsOrderInTransaction сбрасывает порядок проектов в транзакции
func (h *Handlers) resetProjectsOrderInTransaction(projects []models.Project) error {
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for index, project := range projects {
		if err := tx.Model(&project).Update("sort_order", index).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// GetProjectsOrderStats - получение статистики порядка проектов (дополнительная функция)
func (h *Handlers) GetProjectsOrderStats(c *gin.Context) {
	var stats struct {
		TotalProjects    int64 `json:"total_projects"`
		MaxSortOrder     int   `json:"max_sort_order"`
		MinSortOrder     int   `json:"min_sort_order"`
		DuplicateOrders  int64 `json:"duplicate_orders"`
		ProjectsWithZero int64 `json:"projects_with_zero"`
	}

	// Общее количество проектов
	h.db.Model(&models.Project{}).Count(&stats.TotalProjects)

	// Максимальный и минимальный sort_order
	h.db.Model(&models.Project{}).Select("MAX(sort_order)").Row().Scan(&stats.MaxSortOrder)
	h.db.Model(&models.Project{}).Select("MIN(sort_order)").Row().Scan(&stats.MinSortOrder)

	// Проекты с sort_order = 0
	h.db.Model(&models.Project{}).Where("sort_order = 0").Count(&stats.ProjectsWithZero)

	// Дубликаты sort_order
	h.db.Raw(`
		SELECT COUNT(*) 
		FROM (
			SELECT sort_order 
			FROM projects 
			GROUP BY sort_order 
			HAVING COUNT(*) > 1
		) duplicates
	`).Row().Scan(&stats.DuplicateOrders)

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}
