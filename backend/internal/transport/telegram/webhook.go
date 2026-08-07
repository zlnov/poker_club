package telegram

import (
	"net/http"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// WebhookHandler returns a gin handler function that receives Telegram
// updates via webhook and passes them to the bot for processing.
func (b *Bot) WebhookHandler() func(c *gin.Context) {
	return func(c *gin.Context) {
		update := tgbotapi.Update{}
		if err := c.ShouldBindJSON(&update); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid update payload"})
			return
		}
		b.ProcessUpdate(update)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}
