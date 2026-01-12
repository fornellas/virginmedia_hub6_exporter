package exporter

// Primary Service Flow
type ServiceFlow struct {
	// SFID
	ServiceFlowId uint64 `json:"serviceFlowId"`
	Direction     string `json:"direction"`
	// Max Traffic Rate (bps)
	MaxTrafficRate uint64 `json:"maxTrafficRate"`
	// Max Traffic Burst (bytes)
	MaxTrafficBurst uint64 `json:"maxTrafficBurst"`
	// Min Traffic Rate (bps)
	MinReservedRate uint64 `json:"minReservedRate"`
	// Max Concatenated Burst (bytes)
	MaxConcatenatedBurst uint64 `json:"maxConcatenatedBurst"`
	// Scheduling Type
	ScheduleType string `json:"scheduleType"`
}

type ServiceFlowItem struct {
	ServiceFlows []ServiceFlow `json:"serviceFlow"`
}

// GET http://${address}/rest/v1/cablemodem/serviceflows
type ServiceFlows struct {
	ServiceFlowItem `json:"serviceFlows"`
}
