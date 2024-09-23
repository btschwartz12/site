package slack

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
)

var (
	slackWebhookUrl string
)

func init() {
	slackWebhookUrl = os.Getenv("SLACK_WEBHOOK_URL")
}

func SendAlert(logger *zap.SugaredLogger, title string, blocks []Block) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		done := make(chan struct{})
		errChan := make(chan error, 1)

		go func() {
			defer close(done)
			err := sendToSlack(slackWebhookUrl, title, blocks)
			if err != nil {
				errChan <- fmt.Errorf("failed to send message to Slack: %w", err)
				return
			}
			errChan <- nil
		}()

		select {
		case err := <-errChan:
			if err != nil {
				logger.Errorw("error occurred while sending slack message", "error", err, "blocks", blocks)
			} else {
				logger.Infow("sent message to slack")
			}
		case <-ctx.Done():
			logger.Errorw("timed out sending message to slack", "error", ctx.Err())
		}
	}()
}
