package handlers

import (
	"net/http"

	"ledsite/internal/models"

	"github.com/gin-gonic/gin"
)

// /admin/contacts — страница всех заявок
func (h *Handlers) AdminContactsPage(c *gin.Context) {
	var contacts []models.ContactForm
	h.db.Order("created_at DESC").Find(&contacts)

	c.HTML(http.StatusOK, "admin_base.html", gin.H{
		"title":       "Заявки",
		"PageID":      "admin-contacts",
		"contactsAll": contacts,
	})
}
