package exporter

type DownstreamChannel struct {
	// sc_qam=3.0, ofdm=3.1
	ChannelType string `json:"channelType"`
	// Channel ID
	ChannelId uint64 `json:"channelId"`
	// Frequency (Hz)
	Frequency uint64 `json:"frequency"`
	// Power (dBmV)
	Power float64 `json:"power"`
	// Modulation
	Modulation string `json:"modulation"`
	// SNR (dB)
	Snr uint64 `json:"snr"`
	// RxMER (dB)
	RxMer uint64 `json:"rxMer"`
	// Pre RS Errors
	CorrectedErrors uint64 `json:"correctedErrors"`
	// Post RS Errors
	UncorrectedErrors uint64 `json:"uncorrectedErrors"`
	// Locked Status
	LockStatus bool `json:"lockStatus"`
}

type DownstreamItem struct {
	DownstreamChannels []DownstreamChannel `json:"channels"`
}

// http://${address}/rest/v1/cablemodem/downstream
type Downstream struct {
	DownstreamItem DownstreamItem `json:"downstream"`
}
