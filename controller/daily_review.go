package controller

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

func GetDailyReviewReport(c *gin.Context) {
	report, err := service.GetDailyReviewReport(c.Query("date"))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, os.ErrNotExist) {
			status = http.StatusNotFound
		} else if errors.Is(err, service.ErrInvalidDailyReviewDate) {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"success": false, "message": err.Error()})
		return
	}
	common.ApiSuccess(c, report)
}

func RunDailyReviewNow(c *gin.Context) {
	task, _, err := service.StartDailyReviewTask(time.Now().Format("20060102"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	recordManageAudit(c, "daily_review.run", map[string]any{"date": time.Now().Format("20060102")})
	common.ApiSuccess(c, task.ToResponse())
}
