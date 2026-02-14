package exporter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/fornellas/slogxt/log"

	"github.com/prometheus/client_golang/prometheus"

	hub6 "github.com/fornellas/virginmedia_hub6_exporter/hub6"
)

// HubExporter collects metrics from a VirginMedia Hub 6 device.
type HubExporter struct {
	ctx     context.Context
	address string
	client  *http.Client

	// GET http://${address}/rest/v1/cablemodem/state_
	descStateInfo                   *prometheus.Desc
	descStateUptime                 *prometheus.Desc
	descStateAccessAllowed          *prometheus.Desc
	descStateMaxCPEs                *prometheus.Desc
	descStateBaselinePrivacyEnabled *prometheus.Desc
	descStateUp                     *prometheus.Desc

	// GET http://${address}/rest/v1/cablemodem/serviceflows
	descServiceFlowInfo            *prometheus.Desc
	descServiceFlowMaxTrafficRate  *prometheus.Desc
	descServiceFlowMaxTrafficBurst *prometheus.Desc
	descServiceFlowMinReservedRate *prometheus.Desc
	descServiceFlowMaxConcatBurst  *prometheus.Desc
	descServiceFlowsUp             *prometheus.Desc

	// GET http://${address}/rest/v1/cablemodem/upstream
	descUpstreamInfo       *prometheus.Desc
	descUpstreamLockStatus *prometheus.Desc
	descUpstreamPower      *prometheus.Desc
	descUpstreamSymbolRate *prometheus.Desc
	descUpstreamT1         *prometheus.Desc
	descUpstreamT2         *prometheus.Desc
	descUpstreamT3         *prometheus.Desc
	descUpstreamT4         *prometheus.Desc
	descUpstreamUp         *prometheus.Desc

	// GET http://${address}/rest/v1/cablemodem/downstream
	descDownstreamInfo                      *prometheus.Desc
	descDownstreamNumberOfActiveSubcarriers *prometheus.Desc
	descDownstreamPower                     *prometheus.Desc
	descDownstreamSnr                       *prometheus.Desc
	descDownstreamRxMer                     *prometheus.Desc
	descDownstreamCorrectedErrors           *prometheus.Desc
	descDownstreamUncorrectedErrors         *prometheus.Desc
	descDownstreamLockStatus                *prometheus.Desc
	descDownstreamUp                        *prometheus.Desc
}

// NewHubExporter creates a new exporter that will query the hub at address.
// timeout is applied to each HTTP request.
func NewHubExporter(ctx context.Context, address string, client *http.Client) *HubExporter {
	ctx, _ = log.MustWithAttrs(ctx, "virgin_media_hub_address", address)

	return &HubExporter{
		ctx:     ctx,
		address: address,
		client:  client,

		// GET http://${address}/rest/v1/cablemodem/state_
		descStateInfo: prometheus.NewDesc(
			"virginmedia_hub6_state_info",
			"Cable modem info labels (value is always 1)",
			[]string{"boot_filename", "docsis_version", "mac_address", "serial_number", "status"}, nil,
		),
		descStateUptime: prometheus.NewDesc(
			"virginmedia_hub6_state_uptime_seconds",
			"Cable modem uptime in seconds",
			[]string{}, nil,
		),
		descStateAccessAllowed: prometheus.NewDesc(
			"virginmedia_hub6_state_access_allowed",
			"Cable modem access allowed (1 = allowed, 0 = not allowed)",
			[]string{}, nil,
		),
		descStateMaxCPEs: prometheus.NewDesc(
			"virginmedia_hub6_state_max_cpes",
			"Cable modem maximum CPEs",
			[]string{}, nil,
		),
		descStateBaselinePrivacyEnabled: prometheus.NewDesc(
			"virginmedia_hub6_state_baseline_privacy_enabled",
			"Cable modem baseline privacy enabled (1 = enabled, 0 = disabled)",
			[]string{}, nil,
		),
		descStateUp: prometheus.NewDesc(
			"virginmedia_hub6_state_up",
			"Whether the state endpoint was scraped successfully (1 = up, 0 = down)",
			nil, nil,
		),

		// GET http://${address}/rest/v1/cablemodem/serviceflows
		descServiceFlowInfo: prometheus.NewDesc(
			"virginmedia_hub6_serviceflow_info",
			"ServiceFlow info labels (value is always 1)",
			[]string{"serviceflow_id", "direction", "schedule_type"}, nil,
		),
		descServiceFlowMaxTrafficRate: prometheus.NewDesc(
			"virginmedia_hub6_serviceflow_max_traffic_rate_bps",
			"ServiceFlow max traffic rate in bps",
			[]string{"serviceflow_id"}, nil,
		),
		descServiceFlowMaxTrafficBurst: prometheus.NewDesc(
			"virginmedia_hub6_serviceflow_max_traffic_burst_bytes",
			"ServiceFlow max traffic burst in bytes",
			[]string{"serviceflow_id"}, nil,
		),
		descServiceFlowMinReservedRate: prometheus.NewDesc(
			"virginmedia_hub6_serviceflow_min_reserved_rate_bps",
			"ServiceFlow min reserved rate in bps",
			[]string{"serviceflow_id"}, nil,
		),
		descServiceFlowMaxConcatBurst: prometheus.NewDesc(
			"virginmedia_hub6_serviceflow_max_concatenated_burst_bytes",
			"ServiceFlow max concatenated burst in bytes",
			[]string{"serviceflow_id"}, nil,
		),
		descServiceFlowsUp: prometheus.NewDesc(
			"virginmedia_hub6_serviceflows_up",
			"Whether the serviceflows endpoint was scraped successfully (1 = up, 0 = down)",
			nil, nil,
		),

		// GET http://${address}/rest/v1/cablemodem/upstream
		descUpstreamInfo: prometheus.NewDesc(
			"virginmedia_hub6_upstream_info",
			"Upstream info labels (value is always 1)",
			[]string{"channel_id", "modulation", "channel_type", "frequency_hz"}, nil,
		),
		descUpstreamLockStatus: prometheus.NewDesc(
			"virginmedia_hub6_upstream_lock_status",
			"Upstream channel lock status (1 = locked, 0 = unlocked)",
			[]string{"channel_id"}, nil,
		),
		descUpstreamPower: prometheus.NewDesc(
			"virginmedia_hub6_upstream_power_dbmv",
			"Upstream channel power in dBmV",
			[]string{"channel_id"}, nil,
		),
		descUpstreamSymbolRate: prometheus.NewDesc(
			"virginmedia_hub6_upstream_symbol_rate_ksps",
			"Upstream channel symbol rate in ksps",
			[]string{"channel_id"}, nil,
		),
		descUpstreamT1: prometheus.NewDesc(
			"virginmedia_hub6_upstream_t1_timeouts",
			"Upstream channel T1 timeouts",
			[]string{"channel_id"}, nil,
		),
		descUpstreamT2: prometheus.NewDesc(
			"virginmedia_hub6_upstream_t2_timeouts",
			"Upstream channel T2 timeouts",
			[]string{"channel_id"}, nil,
		),
		descUpstreamT3: prometheus.NewDesc(
			"virginmedia_hub6_upstream_t3_timeouts",
			"Upstream channel T3 timeouts",
			[]string{"channel_id"}, nil,
		),
		descUpstreamT4: prometheus.NewDesc(
			"virginmedia_hub6_upstream_t4_timeouts",
			"Upstream channel T4 timeouts",
			[]string{"channel_id"}, nil,
		),
		descUpstreamUp: prometheus.NewDesc(
			"virginmedia_hub6_upstream_up",
			"Whether the upstream endpoint was scraped successfully (1 = up, 0 = down)",
			nil, nil,
		),

		// GET http://${address}/rest/v1/cablemodem/downstream
		descDownstreamInfo: prometheus.NewDesc(
			"virginmedia_hub6_downstream_info",
			"Downstream info labels (value is always 1)",
			[]string{
				"channel_type",
				"channel_id",
				"fft_type",
				"modulation",
				"channel_width_hz",
				"frequency_hz",
				"first_active_subcarrier_hz",
			}, nil,
		),
		descDownstreamNumberOfActiveSubcarriers: prometheus.NewDesc(
			"virginmedia_hub6_downstream_number_of_active_subcarriers",
			"Downstream number of active subcarriers",
			[]string{"channel_id"}, nil,
		),
		descDownstreamPower: prometheus.NewDesc(
			"virginmedia_hub6_downstream_power_dbmv",
			"Downstream channel power in dBmV",
			[]string{"channel_id"}, nil,
		),
		descDownstreamSnr: prometheus.NewDesc(
			"virginmedia_hub6_downstream_snr_db",
			"Downstream channel SNR in dB",
			[]string{"channel_id"}, nil,
		),
		descDownstreamRxMer: prometheus.NewDesc(
			"virginmedia_hub6_downstream_rxmer_db",
			"Downstream channel RxMER in dB",
			[]string{"channel_id"}, nil,
		),
		descDownstreamCorrectedErrors: prometheus.NewDesc(
			"virginmedia_hub6_downstream_corrected_errors",
			"Downstream channel corrected RS errors",
			[]string{"channel_id"}, nil,
		),
		descDownstreamUncorrectedErrors: prometheus.NewDesc(
			"virginmedia_hub6_downstream_uncorrected_errors",
			"Downstream channel uncorrected RS errors",
			[]string{"channel_id"}, nil,
		),
		descDownstreamLockStatus: prometheus.NewDesc(
			"virginmedia_hub6_downstream_lock_status",
			"Downstream channel lock status (1 = locked, 0 = unlocked)",
			[]string{"channel_id"}, nil,
		),
		descDownstreamUp: prometheus.NewDesc(
			"virginmedia_hub6_downstream_up",
			"Whether the downstream endpoint was scraped successfully (1 = up, 0 = down)",
			nil, nil,
		),
	}
}

// Describe sends the descriptors of each metric over the provided channel.
func (h *HubExporter) Describe(ch chan<- *prometheus.Desc) {
	// GET http://${address}/rest/v1/cablemodem/state_
	ch <- h.descStateInfo
	ch <- h.descStateUptime
	ch <- h.descStateAccessAllowed
	ch <- h.descStateMaxCPEs
	ch <- h.descStateBaselinePrivacyEnabled

	// GET http://${address}/rest/v1/cablemodem/serviceflows
	ch <- h.descServiceFlowInfo
	ch <- h.descServiceFlowMaxTrafficRate
	ch <- h.descServiceFlowMaxTrafficBurst
	ch <- h.descServiceFlowMinReservedRate
	ch <- h.descServiceFlowMaxConcatBurst
	ch <- h.descServiceFlowsUp

	// GET http://${address}/rest/v1/cablemodem/upstream
	ch <- h.descUpstreamInfo
	ch <- h.descUpstreamLockStatus
	ch <- h.descUpstreamPower
	ch <- h.descUpstreamSymbolRate
	ch <- h.descUpstreamT1
	ch <- h.descUpstreamT2
	ch <- h.descUpstreamT3
	ch <- h.descUpstreamT4
	ch <- h.descUpstreamUp

	// GET http://${address}/rest/v1/cablemodem/downstream
	ch <- h.descDownstreamInfo
	ch <- h.descDownstreamNumberOfActiveSubcarriers
	ch <- h.descDownstreamPower
	ch <- h.descDownstreamSnr
	ch <- h.descDownstreamRxMer
	ch <- h.descDownstreamCorrectedErrors
	ch <- h.descDownstreamUncorrectedErrors
	ch <- h.descDownstreamLockStatus
	ch <- h.descDownstreamUp
}

func (h *HubExporter) get(path string, out any) (err error) {
	logger := log.MustLogger(h.ctx)

	url := fmt.Sprintf("http://%s%s", h.address, path)
	req, err := http.NewRequestWithContext(h.ctx, "GET", url, nil)
	if err != nil {
		return
	}

	logger.Info("Requesting", "request", req.URL)
	resp, err := h.client.Do(req)
	if err != nil {
		return
	}
	defer func() { errors.Join(err, resp.Body.Close()) }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected HTTP status %d from %s", resp.StatusCode, url)
	}

	dec := json.NewDecoder(resp.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(out)
}

// GET http://${address}/rest/v1/cablemodem/state_
func (h *HubExporter) collectState(ch chan<- prometheus.Metric) {
	logger := log.MustLogger(h.ctx)

	up := 0.0

	var state hub6.State
	path := "/rest/v1/cablemodem/state_"
	if err := h.get(path, &state); err != nil {
		logger.Error("failed to fetch state", "err", err)
	} else {
		up = 1.0

		ch <- prometheus.MustNewConstMetric(
			h.descStateInfo,
			prometheus.GaugeValue,
			1.0,
			state.CableModem.BootFilename,
			state.CableModem.DocsisVersion,
			state.CableModem.MacAddress,
			state.CableModem.SerialNumber,
			state.CableModem.Status,
		)

		ch <- prometheus.MustNewConstMetric(
			h.descStateUptime,
			prometheus.GaugeValue,
			float64(state.CableModem.UpTime),
		)

		access := 0.0
		if state.CableModem.AccessAllowed {
			access = 1.0
		}
		ch <- prometheus.MustNewConstMetric(
			h.descStateAccessAllowed,
			prometheus.GaugeValue, access,
		)

		ch <- prometheus.MustNewConstMetric(
			h.descStateMaxCPEs,
			prometheus.GaugeValue,
			float64(state.CableModem.MaxCpEs),
		)

		privacy := 0.0
		if state.CableModem.BaselinePrivacyEnabled {
			privacy = 1.0
		}
		ch <- prometheus.MustNewConstMetric(h.descStateBaselinePrivacyEnabled, prometheus.GaugeValue, privacy)
	}

	ch <- prometheus.MustNewConstMetric(h.descStateUp, prometheus.GaugeValue, up)
}

// GET http://${address}/rest/v1/cablemodem/serviceflows
func (h *HubExporter) collectServiceFlows(ch chan<- prometheus.Metric) {
	logger := log.MustLogger(h.ctx)

	up := 0.0

	var serviceFlows hub6.ServiceFlows
	if err := h.get("/rest/v1/cablemodem/serviceflows", &serviceFlows); err != nil {
		logger.Error("failed to fetch service flows", "err", err)
	} else {
		up = 1.0

		for _, serviceFlowItem := range serviceFlows.ServiceFlowItems {
			serviceFlowId := strconv.FormatUint(serviceFlowItem.ServiceFlow.ServiceFlowId, 10)
			ch <- prometheus.MustNewConstMetric(
				h.descServiceFlowInfo,
				prometheus.GaugeValue,
				1.0,
				serviceFlowId,
				serviceFlowItem.ServiceFlow.Direction,
				serviceFlowItem.ServiceFlow.ScheduleType,
			)
			ch <- prometheus.MustNewConstMetric(
				h.descServiceFlowMaxTrafficRate,
				prometheus.GaugeValue,
				float64(serviceFlowItem.ServiceFlow.MaxTrafficRate),
				serviceFlowId,
			)
			ch <- prometheus.MustNewConstMetric(
				h.descServiceFlowMaxTrafficBurst,
				prometheus.GaugeValue,
				float64(serviceFlowItem.ServiceFlow.MaxTrafficBurst),
				serviceFlowId,
			)
			ch <- prometheus.MustNewConstMetric(
				h.descServiceFlowMinReservedRate,
				prometheus.GaugeValue,
				float64(serviceFlowItem.ServiceFlow.MinReservedRate),
				serviceFlowId,
			)
			ch <- prometheus.MustNewConstMetric(
				h.descServiceFlowMaxConcatBurst,
				prometheus.GaugeValue,
				float64(serviceFlowItem.ServiceFlow.MaxConcatenatedBurst),
				serviceFlowId,
			)
		}
	}

	ch <- prometheus.MustNewConstMetric(h.descServiceFlowsUp, prometheus.GaugeValue, up)
}

// GET http://${address}/rest/v1/cablemodem/upstream
func (h *HubExporter) collectUpstream(ch chan<- prometheus.Metric) {
	logger := log.MustLogger(h.ctx)

	up := 0.0

	var us hub6.Upstream
	if err := h.get("/rest/v1/cablemodem/upstream", &us); err != nil {
		logger.Error("failed to fetch upstream", "err", err)
	} else {
		up = 1.0

		for _, upstreamChannel := range us.UpstreamItem.Channels {
			channelId := strconv.FormatUint(upstreamChannel.ChannelId, 10)
			ch <- prometheus.MustNewConstMetric(
				h.descUpstreamInfo,
				prometheus.GaugeValue,
				1.0,
				channelId,
				upstreamChannel.Modulation,
				upstreamChannel.ChannelType,
				strconv.FormatUint(upstreamChannel.Frequency, 10),
			)
			lock := 0.0
			if upstreamChannel.LockStatus {
				lock = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				h.descUpstreamLockStatus,
				prometheus.GaugeValue,
				lock,
				channelId,
			)
			ch <- prometheus.MustNewConstMetric(
				h.descUpstreamPower,
				prometheus.GaugeValue,
				upstreamChannel.Power,
				channelId,
			)
			ch <- prometheus.MustNewConstMetric(
				h.descUpstreamSymbolRate,
				prometheus.GaugeValue,
				float64(upstreamChannel.SymbolRate),
				channelId,
			)
			ch <- prometheus.MustNewConstMetric(
				h.descUpstreamT1,
				prometheus.GaugeValue,
				float64(upstreamChannel.T1Timeout),
				channelId,
			)
			ch <- prometheus.MustNewConstMetric(
				h.descUpstreamT2,
				prometheus.GaugeValue,
				float64(upstreamChannel.T2Timeout),
				channelId,
			)
			ch <- prometheus.MustNewConstMetric(
				h.descUpstreamT3,
				prometheus.GaugeValue,
				float64(upstreamChannel.T3Timeout),
				channelId,
			)
			ch <- prometheus.MustNewConstMetric(
				h.descUpstreamT4,
				prometheus.GaugeValue,
				float64(upstreamChannel.T4Timeout),
				channelId,
			)
		}
	}

	ch <- prometheus.MustNewConstMetric(h.descUpstreamUp, prometheus.GaugeValue, up)
}

// GET http://${address}/rest/v1/cablemodem/downstream
func (h *HubExporter) collectDownstream(ch chan<- prometheus.Metric) {
	logger := log.MustLogger(h.ctx)

	up := 0.0

	var downstream hub6.Downstream
	if err := h.get("/rest/v1/cablemodem/downstream", &downstream); err != nil {
		logger.Error("failed to fetch downstream", "err", err)
	} else {
		up = 1.0
		for _, downstreamChannel := range downstream.DownstreamItem.DownstreamChannels {
			channelId := strconv.FormatUint(downstreamChannel.ChannelId, 10)
			fftType := "N/A"
			if downstreamChannel.FFTType != nil {
				fftType = *downstreamChannel.FFTType
			}
			channelWidthHz := "N/A"
			if downstreamChannel.ChannelWidth != nil {
				channelWidthHz = strconv.FormatUint(*downstreamChannel.ChannelWidth, 10)
			}
			firstActiveSubcarrierHz := "N/A"
			if downstreamChannel.FirstActiveSubcarrier != nil {
				firstActiveSubcarrierHz = strconv.FormatUint(*downstreamChannel.FirstActiveSubcarrier, 10)
			}
			ch <- prometheus.MustNewConstMetric(
				h.descDownstreamInfo,
				prometheus.GaugeValue,
				1.0,
				downstreamChannel.ChannelType,
				channelId,
				fftType,
				downstreamChannel.Modulation,
				channelWidthHz,
				strconv.FormatUint(downstreamChannel.Frequency, 10),
				firstActiveSubcarrierHz,
			)
			if downstreamChannel.NumberOfActiveSubcarriers != nil {
				ch <- prometheus.MustNewConstMetric(
					h.descDownstreamNumberOfActiveSubcarriers,
					prometheus.GaugeValue,
					float64(*downstreamChannel.NumberOfActiveSubcarriers),
					channelId,
				)
			}
			ch <- prometheus.MustNewConstMetric(
				h.descDownstreamPower,
				prometheus.GaugeValue,
				downstreamChannel.Power,
				channelId,
			)
			ch <- prometheus.MustNewConstMetric(
				h.descDownstreamSnr,
				prometheus.GaugeValue,
				float64(downstreamChannel.Snr),
				channelId,
			)
			ch <- prometheus.MustNewConstMetric(
				h.descDownstreamRxMer,
				prometheus.GaugeValue,
				float64(downstreamChannel.RxMer),
				channelId,
			)
			ch <- prometheus.MustNewConstMetric(
				h.descDownstreamCorrectedErrors,
				prometheus.GaugeValue,
				float64(downstreamChannel.CorrectedErrors),
				channelId,
			)
			ch <- prometheus.MustNewConstMetric(
				h.descDownstreamUncorrectedErrors,
				prometheus.GaugeValue,
				float64(downstreamChannel.UncorrectedErrors),
				channelId,
			)
			lock := 0.0
			if downstreamChannel.LockStatus {
				lock = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				h.descDownstreamLockStatus,
				prometheus.GaugeValue,
				lock,
				channelId,
			)
		}
	}
	ch <- prometheus.MustNewConstMetric(h.descDownstreamUp, prometheus.GaugeValue, up)

}

// Collect fetches the current state from the Hub and exports metrics.
func (h *HubExporter) Collect(ch chan<- prometheus.Metric) {
	h.collectState(ch)
	h.collectServiceFlows(ch)
	h.collectUpstream(ch)
	h.collectDownstream(ch)
}
