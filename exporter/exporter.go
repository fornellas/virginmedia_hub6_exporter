package exporter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

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
	descStateStatus                 *prometheus.Desc
	descStateMaxCPEs                *prometheus.Desc
	descStateBaselinePrivacyEnabled *prometheus.Desc
	descStateUp                     *prometheus.Desc

	// Descriptors
	// descDownstreamPower       *prometheus.Desc
	// descDownstreamSnr         *prometheus.Desc
	// descDownstreamRxMer       *prometheus.Desc
	// descDownstreamCorrected   *prometheus.Desc
	// descDownstreamUncorrected *prometheus.Desc
	// descDownstreamLockStatus  *prometheus.Desc
	// descDownstreamFrequencyHz *prometheus.Desc

	// descUpstreamPower       *prometheus.Desc
	// descUpstreamSymbolRate  *prometheus.Desc
	// descUpstreamLockStatus  *prometheus.Desc
	// descUpstreamFrequencyHz *prometheus.Desc
	// descUpstreamT1          *prometheus.Desc
	// descUpstreamT2          *prometheus.Desc
	// descUpstreamT3          *prometheus.Desc
	// descUpstreamT4          *prometheus.Desc

	// descServiceMaxTrafficRate  *prometheus.Desc
	// descServiceMaxTrafficBurst *prometheus.Desc
	// descServiceMinReservedRate *prometheus.Desc
	// descServiceMaxConcatBurst  *prometheus.Desc

	// Per-endpoint up metrics (1 = endpoint scraped successfully, 0 = failure)
	// descDownstreamUp   *prometheus.Desc
	// descUpstreamUp     *prometheus.Desc
	// descServiceFlowsUp *prometheus.Desc

}

// NewHubExporter creates a new exporter that will query the hub at address.
// timeout is applied to each HTTP request.
func NewHubExporter(ctx context.Context, address string, client *http.Client) *HubExporter {
	ctx, _ = log.MustWithAttrs(ctx, "virgin_media_hub_address", address)

	// labelsDS := []string{"channel_id", "channel_type", "modulation"}
	// labelsUS := []string{"channel_id", "channel_type", "modulation"}
	// labelsSF := []string{"serviceflow_id", "direction", "schedule_type"}

	return &HubExporter{
		ctx:     ctx,
		address: address,
		client:  client,

		// GET http://${address}/rest/v1/cablemodem/state_
		descStateInfo: prometheus.NewDesc(
			"virginmedia_hub6_state_info",
			"Cable modem info labels (value is always 1)",
			[]string{"boot_filename", "docsis_version", "mac_address", "serial_number"}, nil,
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
		descStateStatus: prometheus.NewDesc(
			"virginmedia_hub6_state_status",
			"Cable modem status (value 1 with status label)",
			[]string{"status"}, nil,
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

		// descDownstreamPower: prometheus.NewDesc(
		// 	"virginmedia_hub6_downstream_power_dbmv",
		// 	"Downstream channel power in dBmV",
		// 	labelsDS, nil,
		// ),
		// descDownstreamSnr: prometheus.NewDesc(
		// 	"virginmedia_hub6_downstream_snr_db",
		// 	"Downstream channel SNR in dB",
		// 	labelsDS, nil,
		// ),
		// descDownstreamRxMer: prometheus.NewDesc(
		// 	"virginmedia_hub6_downstream_rxmer_db",
		// 	"Downstream channel RxMER in dB",
		// 	labelsDS, nil,
		// ),
		// descDownstreamCorrected: prometheus.NewDesc(
		// 	"virginmedia_hub6_downstream_corrected_errors",
		// 	"Downstream channel corrected RS errors",
		// 	labelsDS, nil,
		// ),
		// descDownstreamUncorrected: prometheus.NewDesc(
		// 	"virginmedia_hub6_downstream_uncorrected_errors",
		// 	"Downstream channel uncorrected RS errors",
		// 	labelsDS, nil,
		// ),
		// descDownstreamLockStatus: prometheus.NewDesc(
		// 	"virginmedia_hub6_downstream_lock_status",
		// 	"Downstream channel lock status (1 = locked, 0 = unlocked)",
		// 	labelsDS, nil,
		// ),
		// descDownstreamFrequencyHz: prometheus.NewDesc(
		// 	"virginmedia_hub6_downstream_frequency_hertz",
		// 	"Downstream channel frequency in Hz",
		// 	labelsDS, nil,
		// ),

		// descUpstreamPower: prometheus.NewDesc(
		// 	"virginmedia_hub6_upstream_power_dbmv",
		// 	"Upstream channel power in dBmV",
		// 	labelsUS, nil,
		// ),
		// descUpstreamSymbolRate: prometheus.NewDesc(
		// 	"virginmedia_hub6_upstream_symbol_rate_ksps",
		// 	"Upstream channel symbol rate in ksps",
		// 	labelsUS, nil,
		// ),
		// descUpstreamLockStatus: prometheus.NewDesc(
		// 	"virginmedia_hub6_upstream_lock_status",
		// 	"Upstream channel lock status (1 = locked, 0 = unlocked)",
		// 	labelsUS, nil,
		// ),
		// descUpstreamFrequencyHz: prometheus.NewDesc(
		// 	"virginmedia_hub6_upstream_frequency_hertz",
		// 	"Upstream channel frequency in Hz",
		// 	labelsUS, nil,
		// ),
		// descUpstreamT1: prometheus.NewDesc(
		// 	"virginmedia_hub6_upstream_t1_timeouts",
		// 	"Upstream channel T1 timeouts",
		// 	labelsUS, nil,
		// ),
		// descUpstreamT2: prometheus.NewDesc(
		// 	"virginmedia_hub6_upstream_t2_timeouts",
		// 	"Upstream channel T2 timeouts",
		// 	labelsUS, nil,
		// ),
		// descUpstreamT3: prometheus.NewDesc(
		// 	"virginmedia_hub6_upstream_t3_timeouts",
		// 	"Upstream channel T3 timeouts",
		// 	labelsUS, nil,
		// ),
		// descUpstreamT4: prometheus.NewDesc(
		// 	"virginmedia_hub6_upstream_t4_timeouts",
		// 	"Upstream channel T4 timeouts",
		// 	labelsUS, nil,
		// ),

		// descServiceMaxTrafficRate: prometheus.NewDesc(
		// 	"virginmedia_hub6_serviceflow_max_traffic_rate_bps",
		// 	"ServiceFlow max traffic rate in bps",
		// 	labelsSF, nil,
		// ),
		// descServiceMaxTrafficBurst: prometheus.NewDesc(
		// 	"virginmedia_hub6_serviceflow_max_traffic_burst_bytes",
		// 	"ServiceFlow max traffic burst in bytes",
		// 	labelsSF, nil,
		// ),
		// descServiceMinReservedRate: prometheus.NewDesc(
		// 	"virginmedia_hub6_serviceflow_min_reserved_rate_bps",
		// 	"ServiceFlow min reserved rate in bps",
		// 	labelsSF, nil,
		// ),
		// descServiceMaxConcatBurst: prometheus.NewDesc(
		// 	"virginmedia_hub6_serviceflow_max_concatenated_burst_bytes",
		// 	"ServiceFlow max concatenated burst in bytes",
		// 	labelsSF, nil,
		// ),
		// // per-endpoint up metrics
		// descDownstreamUp: prometheus.NewDesc(
		// 	"virginmedia_hub6_downstream_up",
		// 	"Whether the downstream endpoint was scraped successfully (1 = up, 0 = down)",
		// 	nil, nil,
		// ),
		// descUpstreamUp: prometheus.NewDesc(
		// 	"virginmedia_hub6_upstream_up",
		// 	"Whether the upstream endpoint was scraped successfully (1 = up, 0 = down)",
		// 	nil, nil,
		// ),
		// descServiceFlowsUp: prometheus.NewDesc(
		// 	"virginmedia_hub6_serviceflows_up",
		// 	"Whether the serviceflows endpoint was scraped successfully (1 = up, 0 = down)",
		// 	nil, nil,
		// ),
	}
}

// Describe sends the descriptors of each metric over the provided channel.
func (h *HubExporter) Describe(ch chan<- *prometheus.Desc) {
	// GET http://${address}/rest/v1/cablemodem/state_
	ch <- h.descStateInfo
	ch <- h.descStateUptime
	ch <- h.descStateAccessAllowed
	ch <- h.descStateStatus
	ch <- h.descStateMaxCPEs
	ch <- h.descStateBaselinePrivacyEnabled

	// ch <- e.descDownstreamPower
	// ch <- e.descDownstreamSnr
	// ch <- e.descDownstreamRxMer
	// ch <- e.descDownstreamCorrected
	// ch <- e.descDownstreamUncorrected
	// ch <- e.descDownstreamLockStatus
	// ch <- e.descDownstreamFrequencyHz

	// ch <- e.descUpstreamPower
	// ch <- e.descUpstreamSymbolRate
	// ch <- e.descUpstreamLockStatus
	// ch <- e.descUpstreamFrequencyHz
	// ch <- e.descUpstreamT1
	// ch <- e.descUpstreamT2
	// ch <- e.descUpstreamT3
	// ch <- e.descUpstreamT4

	// ch <- e.descServiceMaxTrafficRate
	// ch <- e.descServiceMaxTrafficBurst
	// ch <- e.descServiceMinReservedRate
	// ch <- e.descServiceMaxConcatBurst

	// // describe per-endpoint up metrics
	// ch <- e.descDownstreamUp
	// ch <- e.descUpstreamUp
	// ch <- e.descServiceFlowsUp
	// ch <- e.descStateUp
}

func (h *HubExporter) get(ctx context.Context, path string, out any) (err error) {
	url := fmt.Sprintf("http://%s%s", h.address, path)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return
	}

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

	stateUp := 0.0

	var st hub6.State
	path := "/rest/v1/cablemodem/state_"
	if err := h.get(h.ctx, path, &st); err != nil {
		logger.Error("failed to fetch state", "err", err)
	} else {
		stateUp = 1.0

		ch <- prometheus.MustNewConstMetric(
			h.descStateInfo,
			prometheus.GaugeValue,
			1.0,
			st.CableModem.BootFilename,
			st.CableModem.DocsisVersion,
			st.CableModem.MacAddress,
			st.CableModem.SerialNumber,
		)

		ch <- prometheus.MustNewConstMetric(
			h.descStateUptime,
			prometheus.GaugeValue,
			float64(st.CableModem.UpTime),
		)

		access := 0.0
		if st.CableModem.AccessAllowed {
			access = 1.0
		}
		ch <- prometheus.MustNewConstMetric(
			h.descStateAccessAllowed,
			prometheus.GaugeValue, access,
		)

		ch <- prometheus.MustNewConstMetric(
			h.descStateStatus,
			prometheus.GaugeValue,
			1.0,
			st.CableModem.Status,
		)

		ch <- prometheus.MustNewConstMetric(
			h.descStateMaxCPEs,
			prometheus.GaugeValue,
			float64(st.CableModem.MaxCpEs),
		)

		privacy := 0.0
		if st.CableModem.BaselinePrivacyEnabled {
			privacy = 1.0
		}
		ch <- prometheus.MustNewConstMetric(h.descStateBaselinePrivacyEnabled, prometheus.GaugeValue, privacy)
	}
	ch <- prometheus.MustNewConstMetric(h.descStateUp, prometheus.GaugeValue, stateUp)
}

// Collect fetches the current state from the Hub and exports metrics.
func (h *HubExporter) Collect(ch chan<- prometheus.Metric) {
	h.collectState(ch)

	// // Downstream
	// dsUp := 0.0
	// if ds, err := e.fetchDownstream(ctx); err == nil {
	// 	dsUp = 1.0
	// 	for _, c := range ds.DownstreamItem.DownstreamChannels {
	// 		labels := []string{strconv.FormatUint(c.ChannelId, 10), c.ChannelType, c.Modulation}
	// 		ch <- prometheus.MustNewConstMetric(e.descDownstreamPower, prometheus.GaugeValue, c.Power, labels...)
	// 		ch <- prometheus.MustNewConstMetric(e.descDownstreamSnr, prometheus.GaugeValue, float64(c.Snr), labels...)
	// 		ch <- prometheus.MustNewConstMetric(e.descDownstreamRxMer, prometheus.GaugeValue, float64(c.RxMer), labels...)
	// 		ch <- prometheus.MustNewConstMetric(e.descDownstreamCorrected, prometheus.GaugeValue, float64(c.CorrectedErrors), labels...)
	// 		ch <- prometheus.MustNewConstMetric(e.descDownstreamUncorrected, prometheus.GaugeValue, float64(c.UncorrectedErrors), labels...)
	// 		lock := 0.0
	// 		if c.LockStatus {
	// 			lock = 1.0
	// 		}
	// 		ch <- prometheus.MustNewConstMetric(e.descDownstreamLockStatus, prometheus.GaugeValue, lock, labels...)
	// 		ch <- prometheus.MustNewConstMetric(e.descDownstreamFrequencyHz, prometheus.GaugeValue, float64(c.Frequency), labels...)
	// 	}
	// }
	// // emit downstream up metric
	// ch <- prometheus.MustNewConstMetric(e.descDownstreamUp, prometheus.GaugeValue, dsUp)

	// // Upstream
	// usUp := 0.0
	// if us, err := e.fetchUpstream(ctx); err == nil {
	// 	usUp = 1.0
	// 	for _, c := range us.UpstreamItem.Channels {
	// 		labels := []string{strconv.FormatUint(c.ChannelId, 10), c.ChannelType, c.Modulation}
	// 		ch <- prometheus.MustNewConstMetric(e.descUpstreamPower, prometheus.GaugeValue, c.Power, labels...)
	// 		ch <- prometheus.MustNewConstMetric(e.descUpstreamSymbolRate, prometheus.GaugeValue, float64(c.SymbolRate), labels...)
	// 		lock := 0.0
	// 		if c.LockStatus {
	// 			lock = 1.0
	// 		}
	// 		ch <- prometheus.MustNewConstMetric(e.descUpstreamLockStatus, prometheus.GaugeValue, lock, labels...)
	// 		ch <- prometheus.MustNewConstMetric(e.descUpstreamFrequencyHz, prometheus.GaugeValue, float64(c.Frequency), labels...)
	// 		ch <- prometheus.MustNewConstMetric(e.descUpstreamT1, prometheus.GaugeValue, float64(c.T1Timeout), labels...)
	// 		ch <- prometheus.MustNewConstMetric(e.descUpstreamT2, prometheus.GaugeValue, float64(c.T2Timeout), labels...)
	// 		ch <- prometheus.MustNewConstMetric(e.descUpstreamT3, prometheus.GaugeValue, float64(c.T3Timeout), labels...)
	// 		ch <- prometheus.MustNewConstMetric(e.descUpstreamT4, prometheus.GaugeValue, float64(c.T4Timeout), labels...)
	// 	}
	// }
	// // emit upstream up metric
	// ch <- prometheus.MustNewConstMetric(e.descUpstreamUp, prometheus.GaugeValue, usUp)

	// // Service Flows
	// sfUp := 0.0
	// if sf, err := e.fetchServiceFlows(ctx); err == nil {
	// 	sfUp = 1.0
	// 	for _, s := range sf.ServiceFlowItem.ServiceFlows {
	// 		labels := []string{strconv.FormatUint(s.ServiceFlowId, 10), s.Direction, s.ScheduleType}
	// 		ch <- prometheus.MustNewConstMetric(e.descServiceMaxTrafficRate, prometheus.GaugeValue, float64(s.MaxTrafficRate), labels...)
	// 		ch <- prometheus.MustNewConstMetric(e.descServiceMaxTrafficBurst, prometheus.GaugeValue, float64(s.MaxTrafficBurst), labels...)
	// 		ch <- prometheus.MustNewConstMetric(e.descServiceMinReservedRate, prometheus.GaugeValue, float64(s.MinReservedRate), labels...)
	// 		ch <- prometheus.MustNewConstMetric(e.descServiceMaxConcatBurst, prometheus.GaugeValue, float64(s.MaxConcatenatedBurst), labels...)
	// 	}
	// }
	// // emit serviceflows up metric
	// ch <- prometheus.MustNewConstMetric(e.descServiceFlowsUp, prometheus.GaugeValue, sfUp)
}

// func (e *HubExporter) fetchDownstream(ctx context.Context) (*hub6.Downstream, error) {
// 	var ds hub6.Downstream
// 	if err := e.fetch(ctx, "/rest/v1/cablemodem/downstream", &ds); err != nil {
// 		return nil, err
// 	}
// 	return &ds, nil
// }

// func (e *HubExporter) fetchUpstream(ctx context.Context) (*hub6.Upstream, error) {
// 	var us hub6.Upstream
// 	if err := e.fetch(ctx, "/rest/v1/cablemodem/upstream", &us); err != nil {
// 		return nil, err
// 	}
// 	return &us, nil
// }

// func (e *HubExporter) fetchServiceFlows(ctx context.Context) (*hub6.ServiceFlows, error) {
// 	var sf hub6.ServiceFlows
// 	if err := e.fetch(ctx, "/rest/v1/cablemodem/serviceflows", &sf); err != nil {
// 		return nil, err
// 	}
// 	return &sf, nil
// }
