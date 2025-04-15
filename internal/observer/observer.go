package observer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/yourusername/payment-monitor/internal/websocket"
	"github.com/yourusername/payment-monitor/pkg/models"
	"gorm.io/gorm"
)

type Observer struct {
	db           *gorm.DB
	config       *Config
	alertChannel chan<- *models.Alert
	hub          *websocket.Hub
}

type Config struct {
	Interval        time.Duration
	Threshold       float64
	MinTransactions int
	Dimensions      []string
}

func NewObserver(db *gorm.DB, config *Config, alertChannel chan<- *models.Alert, hub *websocket.Hub) *Observer {
	return &Observer{
		db:           db,
		config:       config,
		alertChannel: alertChannel,
		hub:          hub,
	}
}

func (o *Observer) Start(ctx context.Context) {
	ticker := time.NewTicker(o.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fmt.Println("observer observing")
			o.checkDimensions()
		}
	}
}

func (o *Observer) checkDimensions() {
	for _, dimension := range o.config.Dimensions {
		fmt.Println("checking dimension", dimension)
		stats, err := o.getPaymentStats(dimension)
		if err != nil {
			fmt.Printf("Error getting stats for dimension %s: %v\n", dimension, err)
			continue
		}

		for _, stat := range stats {
			// Send metrics through WebSocket
			o.hub.BroadcastMetrics(&websocket.MetricsMessage{
				Type:        "metrics",
				Dimension:   stat.Dimension,
				Value:       stat.Value,
				SuccessRate: stat.SuccessRate,
				Timestamp:   stat.Timestamp,
			})

			if stat.Total < o.config.MinTransactions {
				continue
			}
			fmt.Println("current drop percentage", stat.DropPercentage)
			fmt.Println("threshold ", o.config.Threshold)
			if stat.DropPercentage > o.config.Threshold {
				fmt.Println(stat)
				fmt.Printf("alerting for dimension %s drop %f\n", dimension, stat.DropPercentage)
				alert := &models.Alert{
					ID:             fmt.Sprintf("%s-%s-%d", dimension, stat.Value, time.Now().Unix()),
					Dimension:      dimension,
					Value:          stat.Value,
					CurrentRate:    stat.SuccessRate,
					PreviousRate:   stat.PreviousRate,
					DropPercentage: stat.DropPercentage,
					Timestamp:      time.Now(),
				}

				// Add dimension-specific fields
				switch dimension {
				case "gateway":
					fmt.Println("gateway alert triggered for ", stat.Value)
					alert.Gateway = stat.Value
				case "gateway_method":
					// Split the value into gateway and method
					parts := strings.Split(stat.Value, "_")
					if len(parts) == 2 {
						alert.Gateway = parts[0]
						alert.Method = parts[1]
					}
					fmt.Println("gateway method alert triggered for gateway", alert.Gateway, " method ", alert.Method)
				case "gateway_merchant":
					// Split the value into gateway and merchant_id
					parts := strings.Split(stat.Value, "_")
					if len(parts) == 2 {
						alert.Gateway = parts[0]
						alert.MerchantID = parts[1]
					}
					fmt.Println("gateway merchant alert triggered for gateway", alert.Gateway, " merchantID ", alert.MerchantID)
				}

				o.alertChannel <- alert
			}
		}
	}
}

func (o *Observer) getPaymentStats(dimension string) ([]*models.PaymentStats, error) {
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)
	twoHoursAgo := now.Add(-2 * time.Hour)

	switch dimension {
	case "gateway":
		return o.getGatewayStats(oneHourAgo, twoHoursAgo)
	case "gateway_method":
		return o.getGatewayMethodStats(oneHourAgo, twoHoursAgo)
	case "gateway_merchant":
		return o.getGatewayMerchantStats(oneHourAgo, twoHoursAgo)
	default:
		return nil, fmt.Errorf("unknown dimension: %s", dimension)
	}
}

func (o *Observer) getGatewayStats(oneHourAgo, twoHoursAgo time.Time) ([]*models.PaymentStats, error) {
	var currentStats []struct {
		Gateway     string
		Total       int64
		Successful  int64
		SuccessRate float64
	}

	// Get current hour stats
	if err := o.db.Model(&models.Payment{}).
		Select("gateway, COUNT(*) as total, "+
			"SUM(CASE WHEN status = 'STATUS_CAPTURED' THEN 1 ELSE 0 END) as successful, "+
			"AVG(CASE WHEN status = 'STATUS_CAPTURED' THEN 1.0 ELSE 0.0 END) * 100 as success_rate").
		Where("to_timestamp(created_at) >= to_timestamp(?)", oneHourAgo.Unix()).
		Group("gateway").
		Scan(&currentStats).Error; err != nil {
		return nil, err
	}

	fmt.Println("currentStats", currentStats)

	var previousStats []struct {
		Gateway     string
		SuccessRate float64
	}

	// Get previous hour stats
	if err := o.db.Model(&models.Payment{}).
		Select("gateway, AVG(CASE WHEN status = 'STATUS_CAPTURED' THEN 1.0 ELSE 0.0 END) * 100 as success_rate").
		Where("to_timestamp(created_at) >= to_timestamp(?) AND to_timestamp(created_at) < to_timestamp(?)",
			twoHoursAgo.Unix(), oneHourAgo.Unix()).
		Group("gateway").
		Scan(&previousStats).Error; err != nil {
		return nil, err
	}
	fmt.Println("previousStats", previousStats)

	// Combine stats
	stats := make([]*models.PaymentStats, 0, len(currentStats))
	for _, current := range currentStats {
		previousRate := findPreviousRate(previousStats, current.Gateway)
		dropPercentage := previousRate - current.SuccessRate // Show positive drop percentage

		stats = append(stats, &models.PaymentStats{
			Dimension:      "gateway",
			Value:          current.Gateway,
			Total:          int(current.Total),
			Successful:     int(current.Successful),
			SuccessRate:    current.SuccessRate,
			PreviousRate:   previousRate,
			DropPercentage: dropPercentage,
			Timestamp:      time.Now(),
		})
	}

	return stats, nil
}

func findPreviousRate(stats []struct {
	Gateway     string
	SuccessRate float64
}, gateway string) float64 {
	for _, stat := range stats {
		if stat.Gateway == gateway {
			return stat.SuccessRate
		}
	}
	return 0.0
}

func (o *Observer) getGatewayMethodStats(oneHourAgo, twoHoursAgo time.Time) ([]*models.PaymentStats, error) {
	var currentStats []struct {
		Gateway     string
		Method      string
		Total       int64
		Successful  int64
		SuccessRate float64
	}

	// Get current hour stats
	if err := o.db.Model(&models.Payment{}).
		Select("gateway, method, COUNT(*) as total, "+
			"SUM(CASE WHEN status = 'STATUS_CAPTURED' THEN 1 ELSE 0 END) as successful, "+
			"AVG(CASE WHEN status = 'STATUS_CAPTURED' THEN 1.0 ELSE 0.0 END) * 100 as success_rate").
		Where("to_timestamp(created_at) >= to_timestamp(?)", oneHourAgo.Unix()).
		Group("gateway, method").
		Scan(&currentStats).Error; err != nil {
		return nil, err
	}

	var previousStats []struct {
		Gateway     string
		Method      string
		SuccessRate float64
	}

	// Get previous hour stats
	if err := o.db.Model(&models.Payment{}).
		Select("gateway, method, AVG(CASE WHEN status = 'STATUS_CAPTURED' THEN 1.0 ELSE 0.0 END) * 100 as success_rate").
		Where("to_timestamp(created_at) >= to_timestamp(?) AND to_timestamp(created_at) < to_timestamp(?)",
			twoHoursAgo.Unix(), oneHourAgo.Unix()).
		Group("gateway, method").
		Scan(&previousStats).Error; err != nil {
		return nil, err
	}

	// Combine stats
	stats := make([]*models.PaymentStats, 0, len(currentStats))
	for _, current := range currentStats {
		previousRate := findPreviousRateForGatewayMethod(previousStats, current.Gateway, current.Method)
		dropPercentage := previousRate - current.SuccessRate // Show positive drop percentage

		stats = append(stats, &models.PaymentStats{
			Dimension:      "gateway_method",
			Value:          fmt.Sprintf("%s_%s", current.Gateway, current.Method),
			Total:          int(current.Total),
			Successful:     int(current.Successful),
			SuccessRate:    current.SuccessRate,
			PreviousRate:   previousRate,
			DropPercentage: dropPercentage,
			Timestamp:      time.Now(),
		})
	}

	return stats, nil
}

func (o *Observer) getGatewayMerchantStats(oneHourAgo, twoHoursAgo time.Time) ([]*models.PaymentStats, error) {
	var currentStats []struct {
		Gateway     string
		MerchantID  string
		Total       int64
		Successful  int64
		SuccessRate float64
	}

	// Get current hour stats
	if err := o.db.Model(&models.Payment{}).
		Select("gateway, merchant_id, COUNT(*) as total, "+
			"SUM(CASE WHEN status = 'STATUS_CAPTURED' THEN 1 ELSE 0 END) as successful, "+
			"AVG(CASE WHEN status = 'STATUS_CAPTURED' THEN 1.0 ELSE 0.0 END) * 100 as success_rate").
		Where("to_timestamp(created_at) >= to_timestamp(?)", oneHourAgo.Unix()).
		Group("gateway, merchant_id").
		Scan(&currentStats).Error; err != nil {
		return nil, err
	}

	var previousStats []struct {
		Gateway     string
		MerchantID  string
		SuccessRate float64
	}

	// Get previous hour stats
	if err := o.db.Model(&models.Payment{}).
		Select("gateway, merchant_id, AVG(CASE WHEN status = 'STATUS_CAPTURED' THEN 1.0 ELSE 0.0 END) * 100 as success_rate").
		Where("to_timestamp(created_at) >= to_timestamp(?) AND to_timestamp(created_at) < to_timestamp(?)",
			twoHoursAgo.Unix(), oneHourAgo.Unix()).
		Group("gateway, merchant_id").
		Scan(&previousStats).Error; err != nil {
		return nil, err
	}

	// Combine stats
	stats := make([]*models.PaymentStats, 0, len(currentStats))
	for _, current := range currentStats {
		previousRate := findPreviousRateForGatewayMerchant(previousStats, current.Gateway, current.MerchantID)
		dropPercentage := previousRate - current.SuccessRate // Show positive drop percentage

		stats = append(stats, &models.PaymentStats{
			Dimension:      "gateway_merchant",
			Value:          fmt.Sprintf("%s_%s", current.Gateway, current.MerchantID),
			Total:          int(current.Total),
			Successful:     int(current.Successful),
			SuccessRate:    current.SuccessRate,
			PreviousRate:   previousRate,
			DropPercentage: dropPercentage,
			Timestamp:      time.Now(),
		})
	}

	return stats, nil
}

func findPreviousRateForGatewayMethod(stats []struct {
	Gateway     string
	Method      string
	SuccessRate float64
}, gateway, method string) float64 {
	for _, stat := range stats {
		if stat.Gateway == gateway && stat.Method == method {
			return stat.SuccessRate
		}
	}
	return 0.0
}

func findPreviousRateForGatewayMerchant(stats []struct {
	Gateway     string
	MerchantID  string
	SuccessRate float64
}, gateway, merchantID string) float64 {
	for _, stat := range stats {
		if stat.Gateway == gateway && stat.MerchantID == merchantID {
			return stat.SuccessRate
		}
	}
	return 0.0
}
