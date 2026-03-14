package common

import (
	"log"
	"time"
	"zai2api-go/services"
)

func StartDailyResetScheduler() {
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
			time.Sleep(next.Sub(now))
			if err := services.ResetDailyCallCount(); err != nil {
				log.Printf("Reset daily call count failed: %v", err)
			} else {
				log.Println("Daily call count reset successfully")
			}
		}
	}()
}
