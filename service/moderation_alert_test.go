package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/require"
)

func TestRecordModerationAlertSendsEmailAfterThreshold(t *testing.T) {
	t.Setenv("MODERATION_ALERT_EMAIL", "alert@example.com")
	t.Setenv("MODERATION_ALERT_THRESHOLD", "2")

	resetModerationAlertStateForTest()

	previousSender := moderationAlertEmailSender
	defer func() { moderationAlertEmailSender = previousSender }()

	var sendCount int
	var gotSubject string
	var gotReceiver string
	var gotContent string
	moderationAlertEmailSender = func(subject string, receiver string, content string) error {
		sendCount++
		gotSubject = subject
		gotReceiver = receiver
		gotContent = content
		return nil
	}

	RecordModerationAlert(context.Background(), errors.New("moderation upstream returned status 429"))
	require.Equal(t, 0, sendCount)

	RecordModerationAlert(context.Background(), errors.New("send moderation request: dial tcp: connect: connection refused"))
	require.Equal(t, 1, sendCount)
	require.Contains(t, gotSubject, common.SystemName+" moderation alert")
	require.Equal(t, "alert@example.com", gotReceiver)
	require.Contains(t, gotContent, "System: "+common.SystemName)
	require.Contains(t, gotContent, "Node:")
	require.Contains(t, gotContent, "Window: last 30 minutes")
	require.Contains(t, gotContent, "Failure count: 2")
	require.Contains(t, gotContent, "Threshold: 2")
	require.Contains(t, gotContent, "dial tcp: connect: connection refused")

	RecordModerationAlert(context.Background(), errors.New("send moderation request: dial tcp: connection refused again"))
	require.Equal(t, 1, sendCount)
	require.True(t, strings.Contains(gotContent, "No user prompt content is included"))
}

func TestRecordModerationAlertSkipsWithoutRecipient(t *testing.T) {
	t.Setenv("MODERATION_ALERT_EMAIL", "")
	t.Setenv("MODERATION_ALERT_THRESHOLD", "1")

	resetModerationAlertStateForTest()

	previousSender := moderationAlertEmailSender
	defer func() { moderationAlertEmailSender = previousSender }()
	moderationAlertEmailSender = func(subject string, receiver string, content string) error {
		t.Fatal("email sender should not be called")
		return nil
	}

	RecordModerationAlert(context.Background(), errors.New("send moderation request: dial tcp"))
}
