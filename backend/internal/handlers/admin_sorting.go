package handlers

import (
	"log"
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
		log.Printf("Ошибка обновления порядка для проекта ID=%d: %v", id, err)
		jsonErr(c, http.StatusInternalServerError, "Ошибка обновления порядка")
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
		log.Printf("Ошибка обновления порядка проектов (bulk): %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка обновления порядка проектов",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Порядок всех проектов обновлен",
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
