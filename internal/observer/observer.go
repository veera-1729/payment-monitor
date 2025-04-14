package observer

import (
	"context"
	"fmt"
	"time"

	"github.com/yourusername/payment-monitor/pkg/models"
	"gorm.io/gorm"
)

type Observer struct {
	db           *gorm.DB
	config       *Config
	alertChannel chan<- *models.Alert
}

type Config struct {
	Interval        time.Duration
	Threshold       float64
	MinTransactions int
	Dimensions      []string
}

func NewObserver(db *gorm.DB, config *Config, alertChannel chan<- *models.Alert) *Observer {
	return &Observer{
		db:           db,
		config:       config,
		alertChannel: alertChannel,
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
			if stat.Total < o.config.MinTransactions {
				continue
			}

			if true {
				fmt.Println("alerting for dimension", dimension, stat.Value)
				alert := &models.Alert{
					ID:             fmt.Sprintf("%s-%s-%d", dimension, stat.Value, time.Now().Unix()),
					Dimension:      dimension,
					Value:          stat.Value,
					CurrentRate:    stat.SuccessRate,
					PreviousRate:   stat.PreviousRate,
					DropPercentage: stat.DropPercentage,
					Timestamp:      time.Now(),
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
	case "gateway_payment_method":
		return o.getGatewayMethodStats(oneHourAgo, twoHoursAgo)
	case "gateway_merchant":
		return o.getGatewayMerchantStats(oneHourAgo, twoHoursAgo)
	default:
		return nil, fmt.Errorf("unknown dimension: %s", dimension)
	}
}

func (o *Observer) getGatewayStats(oneHourAgo, twoHoursAgo time.Time) ([]*models.PaymentStats, error) {
	// Try the stats query using status to determine success
	var rawResults []struct {
		Gateway     string
		Total       int64
		Successful  int64
		SuccessRate float64
	}
	rawQuery := `
		SELECT 
			gateway,
			COUNT(*) as total,
			SUM(CASE WHEN status = 'STATUS_CAPTURED' THEN 1 ELSE 0 END) as successful,
			ROUND(AVG(CASE WHEN status = 'STATUS_CAPTURED' THEN 1.0 ELSE 0.0 END) * 100, 2) as success_rate
		FROM payments
		GROUP BY gateway
	`
	if err := o.db.Raw(rawQuery).Scan(&rawResults).Error; err != nil {
		fmt.Printf("Error executing stats query: %v\n", err)
		return nil, err
	}
	fmt.Printf("Stats query results: %+v\n", rawResults)

	// Combine stats
	stats := make([]*models.PaymentStats, 0, len(rawResults))
	for _, result := range rawResults {
		stats = append(stats, &models.PaymentStats{
			Dimension:      "gateway",
			Value:          result.Gateway,
			Total:          int(result.Total),
			Successful:     int(result.Successful),
			SuccessRate:    result.SuccessRate,
			PreviousRate:   0.0,
			DropPercentage: 0.0,
			Timestamp:      time.Now(),
		})
	}

	return stats, nil
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
		Where("created_at >= ?", oneHourAgo.Unix()).
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
		Where("created_at >= ? AND created_at < ?", twoHoursAgo.Unix(), oneHourAgo.Unix()).
		Group("gateway, method").
		Scan(&previousStats).Error; err != nil {
		return nil, err
	}

	// Combine stats
	stats := make([]*models.PaymentStats, 0, len(currentStats))
	for _, current := range currentStats {
		previousRate := findPreviousRateForGatewayMethod(previousStats, current.Gateway, current.Method)
		dropPercentage := current.SuccessRate - previousRate

		stats = append(stats, &models.PaymentStats{
			Dimension:      "gateway_payment_method",
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
		Where("created_at >= ?", oneHourAgo.Unix()).
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
		Where("created_at >= ? AND created_at < ?", twoHoursAgo.Unix(), oneHourAgo.Unix()).
		Group("gateway, merchant_id").
		Scan(&previousStats).Error; err != nil {
		return nil, err
	}

	// Combine stats
	stats := make([]*models.PaymentStats, 0, len(currentStats))
	for _, current := range currentStats {
		previousRate := findPreviousRateForGatewayMerchant(previousStats, current.Gateway, current.MerchantID)
		dropPercentage := current.SuccessRate - previousRate

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
