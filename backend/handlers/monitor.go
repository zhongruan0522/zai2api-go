package handlers

import (
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
	"zai2api-go/database"
	"zai2api-go/models"

	"github.com/gin-gonic/gin"
)

type ChannelAvailability struct {
	Total        int64   `json:"total"`
	Success      int64   `json:"success"`
	Failed       int64   `json:"failed"`
	Availability float64 `json:"availability"`
}

type MonitorChannels struct {
	OCR   ChannelAvailability `json:"ocr"`
	Chat  ChannelAvailability `json:"chat"`
	Image ChannelAvailability `json:"image"`
}

type MonitorSummary struct {
	RecentHour MonitorChannels `json:"recent_hour"`
	Today      MonitorChannels `json:"today"`
}

type TrendPoint struct {
	Label        string  `json:"label"`
	Total        int64   `json:"total"`
	Success      int64   `json:"success"`
	Failed       int64   `json:"failed"`
	Availability float64 `json:"availability"`
}

type ChannelTrends struct {
	OCR   []TrendPoint `json:"ocr"`
	Chat  []TrendPoint `json:"chat"`
	Image []TrendPoint `json:"image"`
}

func GetMonitorSummary(c *gin.Context) {
	now := monitorNow()
	currentHourStart := startOfHour(now)
	recentHourStart := currentHourStart.Add(-time.Hour)
	dayStart := startOfDay(now)

	summary := MonitorSummary{
		RecentHour: MonitorChannels{
			OCR:   channelAvailability(&models.OCRLog{}, recentHourStart, currentHourStart),
			Chat:  channelAvailability(&models.ChatLog{}, recentHourStart, currentHourStart),
			Image: channelAvailability(&models.ImageLog{}, recentHourStart, currentHourStart),
		},
		Today: MonitorChannels{
			OCR:   channelAvailability(&models.OCRLog{}, dayStart, now),
			Chat:  channelAvailability(&models.ChatLog{}, dayStart, now),
			Image: channelAvailability(&models.ImageLog{}, dayStart, now),
		},
	}
	c.JSON(http.StatusOK, summary)
}

func GetMonitorDaily(c *gin.Context) {
	days := 7
	if d := c.Query("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v >= 1 && v <= 90 {
			days = v
		}
	}

	now := monitorNow()
	start := startOfDay(now).AddDate(0, 0, -days+1)
	end := start.AddDate(0, 0, days)

	trends := ChannelTrends{
		OCR:   channelDailyTrends(&models.OCRLog{}, start, end, days),
		Chat:  channelDailyTrends(&models.ChatLog{}, start, end, days),
		Image: channelDailyTrends(&models.ImageLog{}, start, end, days),
	}
	c.JSON(http.StatusOK, trends)
}

func GetMonitorHourly(c *gin.Context) {
	now := monitorNow()
	start := startOfDay(now)
	end := startOfHour(now).Add(time.Hour)
	hours := int(end.Sub(start) / time.Hour)

	trends := ChannelTrends{
		OCR:   channelHourlyTrends(&models.OCRLog{}, start, end, hours),
		Chat:  channelHourlyTrends(&models.ChatLog{}, start, end, hours),
		Image: channelHourlyTrends(&models.ImageLog{}, start, end, hours),
	}
	c.JSON(http.StatusOK, trends)
}

func channelAvailability(sample interface{}, start time.Time, end time.Time) ChannelAvailability {
	var ca ChannelAvailability

	database.DB.Model(sample).
		Select(`
			COUNT(*) as total,
			COALESCE(SUM(CASE WHEN success THEN 1 ELSE 0 END), 0) as success,
			COALESCE(SUM(CASE WHEN NOT success THEN 1 ELSE 0 END), 0) as failed
		`).
		Where("created_at >= ? AND created_at < ?", start, end).
		Scan(&ca)

	if ca.Total > 0 {
		ca.Availability = math.Round(float64(ca.Success)/float64(ca.Total)*10000) / 100
	}

	return ca
}

func channelDailyTrends(sample interface{}, start time.Time, end time.Time, days int) []TrendPoint {
	type aggRow struct {
		Bucket  string
		Total   int64
		Success int64
		Failed  int64
	}

	rows := make([]aggRow, 0, days)
	dateExpr := "DATE(created_at)"
	database.DB.Model(sample).
		Select("TO_CHAR("+dateExpr+", 'YYYY-MM-DD') as bucket, COUNT(*) as total, COALESCE(SUM(CASE WHEN success THEN 1 ELSE 0 END), 0) as success, COALESCE(SUM(CASE WHEN NOT success THEN 1 ELSE 0 END), 0) as failed").
		Where("created_at >= ? AND created_at < ?", start, end).
		Group(dateExpr).
		Order("bucket ASC").
		Scan(&rows)

	index := make(map[string]aggRow)
	for _, r := range rows {
		index[r.Bucket] = r
	}

	points := make([]TrendPoint, 0, days)
	for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		pt := TrendPoint{Label: d.Format("01-02")}
		if r, ok := index[key]; ok {
			pt.Total = r.Total
			pt.Success = r.Success
			pt.Failed = r.Failed
			if r.Total > 0 {
				pt.Availability = math.Round(float64(r.Success)/float64(r.Total)*10000) / 100
			}
		}
		points = append(points, pt)
	}

	return points
}

func channelHourlyTrends(sample interface{}, start time.Time, end time.Time, hours int) []TrendPoint {
	type aggRow struct {
		Bucket  string
		Total   int64
		Success int64
		Failed  int64
	}

	rows := make([]aggRow, 0, hours)
	hourExpr := "DATE_TRUNC('hour', created_at)"
	database.DB.Model(sample).
		Select("TO_CHAR("+hourExpr+", 'YYYY-MM-DD HH24:MI') as bucket, COUNT(*) as total, COALESCE(SUM(CASE WHEN success THEN 1 ELSE 0 END), 0) as success, COALESCE(SUM(CASE WHEN NOT success THEN 1 ELSE 0 END), 0) as failed").
		Where("created_at >= ? AND created_at < ?", start, end).
		Group(hourExpr).
		Order("bucket ASC").
		Scan(&rows)

	index := make(map[string]aggRow)
	for _, r := range rows {
		index[r.Bucket] = r
	}

	points := make([]TrendPoint, 0, hours)
	for h := start; h.Before(end); h = h.Add(time.Hour) {
		key := h.Format("2006-01-02 15:04")
		pt := TrendPoint{Label: h.Format("15")}
		if r, ok := index[key]; ok {
			pt.Total = r.Total
			pt.Success = r.Success
			pt.Failed = r.Failed
			if r.Total > 0 {
				pt.Availability = math.Round(float64(r.Success)/float64(r.Total)*10000) / 100
			}
		}
		points = append(points, pt)
	}

	return points
}

func monitorNow() time.Time {
	return time.Now().In(monitorLocation())
}

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func startOfHour(t time.Time) time.Time {
	return t.Truncate(time.Hour)
}

func monitorLocation() *time.Location {
	tz := os.Getenv("DB_TIMEZONE")
	if tz == "" {
		tz = "Asia/Shanghai"
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.FixedZone("UTC+8", 8*60*60)
	}
	return loc
}
