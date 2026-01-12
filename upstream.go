package exporter

type UpstreamChannel struct {
	// Channel ID
	ChannelId uint64 `json:"channelId"`
	// Frequency (Hz)
	Frequency  uint64 `json:"frequency"`
	LockStatus bool   `json:"lockStatus"`
	// Power (dBmV)
	Power float64 `json:"power"`
	// Symbol Rate (ksps)
	SymbolRate uint64 `json:"symbolRate"`
	// Modulation
	Modulation string `json:"modulation"`
	// T1 Timeouts
	T1Timeout uint64 `json:"t1Timeout"`
	// T2 Timeouts
	T2Timeout uint64 `json:"t2Timeout"`
	// T3 Timeouts
	T3Timeout uint64 `json:"t3Timeout"`
	// T4 Timeouts
	T4Timeout uint64 `json:"t4Timeout"`
	// Channel Type
	ChannelType string `json:"channelType"`
}

type UpstreamItem struct {
	// 3.0 Upstream channels
	Channels []UpstreamChannel `json:"channels"`
}

type Upstream struct {
	UpstreamItem UpstreamItem `json:"upstream"`
}
