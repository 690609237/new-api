package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
)

const (
	contactEmailDailyLimit = 3
	contactEmailKeyPrefix  = "contact_email_daily:"
)

var contactEmailCounts = struct {
	sync.Mutex
	counts map[string]int
}{counts: make(map[string]int)}

const contactEmailTakeScript = `
local count = redis.call('INCR', KEYS[1])
if count == 1 then
  redis.call('EXPIRE', KEYS[1], ARGV[1])
end
if count > tonumber(ARGV[2]) then
  redis.call('DECR', KEYS[1])
  return 0
end
return 1
`

const contactEmailReleaseScript = `
local count = redis.call('DECR', KEYS[1])
if count <= 0 then
  redis.call('DEL', KEYS[1])
end
return count
`

func contactEmailKey(userID int, now time.Time) string {
	return fmt.Sprintf("%s%d:%s", contactEmailKeyPrefix, userID, now.Format("20060102"))
}

// ReserveContactEmail reserves one of the user's three daily contact messages.
// The reservation can be released when SMTP delivery fails, so failed sends do
// not consume the daily allowance.
func ReserveContactEmail(ctx context.Context, userID int) (bool, error) {
	if userID <= 0 {
		return false, fmt.Errorf("invalid user id")
	}
	now := time.Now()
	key := contactEmailKey(userID, now)
	if common.RedisEnabled && common.RDB != nil {
		result, err := common.RDB.Eval(ctx, contactEmailTakeScript, []string{key}, 48*60*60, contactEmailDailyLimit).Int()
		if err != nil {
			return false, fmt.Errorf("contact email rate limit failed: %w", err)
		}
		return result == 1, nil
	}

	contactEmailCounts.Lock()
	defer contactEmailCounts.Unlock()
	if contactEmailCounts.counts[key] >= contactEmailDailyLimit {
		return false, nil
	}
	contactEmailCounts.counts[key]++
	return true, nil
}

func ReleaseContactEmail(ctx context.Context, userID int) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user id")
	}
	key := contactEmailKey(userID, time.Now())
	if common.RedisEnabled && common.RDB != nil {
		if _, err := common.RDB.Eval(ctx, contactEmailReleaseScript, []string{key}).Result(); err != nil {
			return fmt.Errorf("contact email rate limit release failed: %w", err)
		}
		return nil
	}

	contactEmailCounts.Lock()
	defer contactEmailCounts.Unlock()
	if count := contactEmailCounts.counts[key]; count <= 1 {
		delete(contactEmailCounts.counts, key)
	} else {
		contactEmailCounts.counts[key] = count - 1
	}
	return nil
}
