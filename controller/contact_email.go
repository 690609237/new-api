package controller

import (
	"html"
	"net/http"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

const (
	contactEmailReceiver = "1549277597@qq.com"
	contactEmailMaxTitle = 120
	contactEmailMaxBody  = 5000
)

type contactEmailRequest struct {
	Subject string `json:"subject"`
	Content string `json:"content"`
}

func SendContactEmail(c *gin.Context) {
	var request contactEmailRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request"})
		return
	}

	subject := strings.TrimSpace(request.Subject)
	content := strings.TrimSpace(request.Content)
	if subject == "" || len([]rune(subject)) > contactEmailMaxTitle {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Please enter a valid email subject"})
		return
	}
	if content == "" || len([]rune(content)) > contactEmailMaxBody {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Please enter valid email content"})
		return
	}

	userID := c.GetInt("id")
	user, err := model.GetSelfUserById(userID)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	reserved, err := service.ReserveContactEmail(c.Request.Context(), userID)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if !reserved {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"success": false,
			"code":    "CONTACT_EMAIL_DAILY_LIMIT",
			"message": "You can send up to 3 contact emails per day",
		})
		return
	}

	body := "<p><strong>User ID:</strong> " + html.EscapeString(strconv.Itoa(user.Id)) + "</p>" +
		"<p><strong>Username:</strong> " + html.EscapeString(user.Username) + "</p>" +
		"<p><strong>Registered email:</strong> " + html.EscapeString(user.Email) + "</p>" +
		"<hr><p>" + html.EscapeString(content) + "</p>"
	if err := common.SendEmail(subject, contactEmailReceiver, body); err != nil {
		_ = service.ReleaseContactEmail(c.Request.Context(), userID)
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Email sent successfully"})
}
