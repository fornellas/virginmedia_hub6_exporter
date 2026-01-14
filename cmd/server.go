package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/fornellas/slogxt/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"

	"github.com/fornellas/virginmedia_hub6_exporter/exporter"
)

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the Virgin Media Hub 6 Prometheus exporter HTTP server",
	Run: GetRunFn(func(cmd *cobra.Command, args []string) error {
		logger := log.MustLogger(cmd.Context())

		hubAddr, err := cmd.Flags().GetString("address")
		if err != nil {
			return err
		}
		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			return err
		}

		hubExporter := exporter.NewHubExporter(hubAddr, 5*time.Second)

		registry := prometheus.NewRegistry()
		if err := registry.Register(hubExporter); err != nil {
			return err
		}

		handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

		mux := http.NewServeMux()
		mux.Handle("/metrics", handler)

		listen := fmt.Sprintf(":%d", port)
		logger.Info("Starting server", "listen", listen, "hub_address", hubAddr)
		return http.ListenAndServe(listen, mux)
	}),
}

func init() {
	ServerCmd.Flags().String("address", "192.168.100.1", "Address of VirginMedia Hub 6")
	ServerCmd.Flags().Int("port", 9188, "HTTP listen port for the exporter")

	RootCmd.AddCommand(ServerCmd)
}
