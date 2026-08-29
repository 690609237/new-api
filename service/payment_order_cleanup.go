package service

import (
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

const (
	pendingPaymentOrderRetention = 24 * time.Hour
	paymentOrderCleanupInterval  = time.Hour
)

// StartPaymentOrderCleanupTask periodically removes payment and subscription
// orders that were never completed. Only the master node performs the cleanup
// so a multi-node deployment does not run duplicate delete scans.
func StartPaymentOrderCleanupTask() {
	if !common.IsMasterNode {
		return
	}
	go func() {
		cleanupPendingPaymentOrders()
		ticker := time.NewTicker(paymentOrderCleanupInterval)
		defer ticker.Stop()
		for range ticker.C {
			cleanupPendingPaymentOrders()
		}
	}()
}

func cleanupPendingPaymentOrders() {
	cutoff := time.Now().Add(-pendingPaymentOrderRetention).Unix()
	topUps, subscriptionOrders, err := model.DeleteExpiredPendingPaymentOrders(cutoff)
	if err != nil {
		common.SysError("failed to delete expired pending payment orders: " + err.Error())
		return
	}
	if topUps > 0 || subscriptionOrders > 0 {
		common.SysLog(fmt.Sprintf("deleted expired pending payment orders: topups=%d subscription_orders=%d", topUps, subscriptionOrders))
	}
}
