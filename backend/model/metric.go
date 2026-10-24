package model

// SystemMetric 性能监控指标
type SystemMetric struct {
	ID           uint    `json:"id" gorm:"primaryKey"`
	Timestamp    int64   `json:"timestamp" gorm:"index:idx_timestamp"`
	CPUPercent   float64 `json:"cpu_percent"`
	CPUMaxCore   float64 `json:"cpu_max_core_percent"`
	MemUsed      float64 `json:"mem_used"`       // GB, 保留2位小数
	SwapUsed     float64 `json:"swap_used"`      // GB, 保留2位小数
	NetUpSpeed   float64 `json:"net_up_speed"`   // KB/s, 保留2位小数
	NetDownSpeed float64 `json:"net_down_speed"` // KB/s, 保留2位小数
	DiskUsed     float64 `json:"disk_used"`      // GB, 保留2位小数
}
