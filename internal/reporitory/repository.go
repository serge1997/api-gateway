package reporitory

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/serge1997/apigateway/internal/stream"
	"gorm.io/gorm"
)

type repo struct {
	db *gorm.DB
}

func New(db *gorm.DB) *repo {
	return &repo{db}
}

func (r repo) AvgLatency(ctx context.Context) (interface{}, error) {
	type LatencyResult struct {
		Service  string  `json:"service"`
		Duration float64 `json:"duration"`
	}
	var results []LatencyResult
	if err := r.db.WithContext(ctx).
		Raw("select service, avg(duration) as duration from request_streams group by service").
		Scan(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

func (r repo) ErrRate(ctx context.Context) (interface{}, error) {
	type ErrRateResult struct {
		Service string  `json:"service"`
		Rate    float64 `json:"rate"`
	}
	var results []ErrRateResult
	if err := r.db.WithContext(ctx).
		Raw(`select 
			service,
			(count(status) * 100) / (select count(*) from request_streams) as Rate
		from request_streams
		where status > 299
		GROUP BY service;`).Scan(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

func (r repo) ReqInMinute(ctx context.Context, minutes int) (interface{}, error) {
	var minuteFilter = "-1 minute"
	if minutes > 1 {
		minuteFilter = fmt.Sprintf("-%d minutes", minutes)
	}
	type ReqInMinute struct {
		Service string `gorm:"service" json:"service"`
		Total   int64  `gorm:"total" json:"total"`
	}
	var result []ReqInMinute
	if err := r.db.WithContext(ctx).Raw(`
		select 
    		service,
    		count(*) total
		from request_streams
		where created_at >= datetime('now', 'localtime',  @minutes)
		GROUP BY service;
	`, sql.Named("minutes", minuteFilter)).Scan(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (r repo) ReqInMinuteGroupedByMinute(ctx context.Context, minutes int) (interface{}, error) {
	var minuteFilter = "-2 minutes"
	if minutes > 1 {
		minuteFilter = fmt.Sprintf("-%d minutes", minutes)
	}
	type ReqInMinutesGrouped struct {
		Service string `json:"service"`
		Total   int64  `json:"total"`
		Min     string `json:"min"`
	}
	var result []ReqInMinutesGrouped
	if err := r.db.WithContext(ctx).
		Raw(`
		select 
			strftime('%H:%M', datetime(created_at, 'localtime')) as min,
			service,
			count(*) as total
		from request_streams
		where created_at >= datetime('now', 'localtime', @minutes) 
		GROUP BY strftime('%H:%M', created_at), service`,
			sql.Named("minutes", minuteFilter),
		).Scan(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (r repo) HttpStatusMetricsOfService(ctx context.Context, service string) (interface{}, error) {
	type StatusMetrics struct {
		Status int     `json:"status"`
		Perc   float64 `json:"perc"`
	}
	var result []StatusMetrics
	if err := r.db.WithContext(ctx).
		Raw(`
		select 
			status,
			count(*) * 100 / (SELECT count(*) from request_streams) as perc
		from request_streams
		where service = @service
		GROUP BY status`, sql.Named("service", service)).Scan(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (r repo) HttpMethodsMetricsOfService(ctx context.Context, service string) (interface{}, error) {
	type StatusMethod struct {
		Method string  `json:"method"`
		Perc   float64 `json:"perc"`
	}
	var result []StatusMethod
	if err := r.db.WithContext(ctx).
		Raw(`
		select 
			method,
			count(*) * 100 / (SELECT count(*) from request_streams) as perc
		from request_streams
		where service = @service
		GROUP BY method;`, sql.Named("service", service)).Scan(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (r repo) ReqInMinuteGroupedByMinuteOfService(ctx context.Context, minutes int, service string) (interface{}, error) {
	var minuteFilter = "-60 minutes"
	if minutes > 1 {
		minuteFilter = fmt.Sprintf("-%d minutes", minutes)
	}
	type ReqInMinutesGrouped struct {
		Service string `json:"service"`
		Total   int64  `json:"total"`
		Min     string `json:"min"`
	}
	var result []ReqInMinutesGrouped
	if err := r.db.WithContext(ctx).
		Raw(`
		select 
			strftime('%H:%M', datetime(created_at, 'localtime')) as min,
			service,
			count(*) as total
		from request_streams
		where created_at >= datetime('now', 'localtime', @minutes)
		and service = @service 
		GROUP BY strftime('%H:%M', created_at), service`,
			sql.Named("minutes", minuteFilter), sql.Named("service", service),
		).Scan(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (r repo) AllInLastHourOfService(ctx context.Context, service string) (interface{}, error) {
	var result []stream.RequestStream
	if err := r.db.WithContext(ctx).
		Raw(`select * from request_streams 
			where created_at >= datetime('now', 'localtime',  '-1 hour')
			order by id desc
			`,
		).Scan(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (r repo) DurationMetricsOfService(ctx context.Context, service string) (interface{}, error) {
	type DurationMetric struct {
		Min         string  `json:"min"`
		AvgDuration float64 `json:"avgDuration"`
		MaxDuration float64 `json:"maxDuration"`
	}
	var result []DurationMetric
	if err := r.db.WithContext(ctx).
		Raw(`SELECT 
				strftime('%H:%M', datetime(created_at, 'localtime')) AS min,
				AVG(duration) AS avg_duration,
				MAX(duration) AS max_duration
			FROM request_streams
			WHERE service = @service
			and created_at >= datetime('now', 'localtime',  '-1 hour')
			GROUP BY min
			ORDER BY min ASC`, sql.Named("service", service),
		).Scan(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}
