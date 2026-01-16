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

		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			return err
		}

		// /probe implements the multi-target exporter pattern. It expects a GET
		// parameter "target" containing the address of the Hub to probe.
		mux := http.NewServeMux()
		mux.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
			target := r.URL.Query().Get("target")
			if target == "" {
				http.Error(w, "missing 'target' parameter", http.StatusBadRequest)
				return
			}

			registry := prometheus.NewRegistry()
			hubExporter := exporter.NewHubExporter(target, 5*time.Second)
			registry.MustRegister(hubExporter)

			handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
			handler.ServeHTTP(w, r)
		})

		listen := fmt.Sprintf(":%d", port)
		logger.Info("Starting server", "listen", listen)
		return http.ListenAndServe(listen, mux)
	}),
}

func init() {
	ServerCmd.Flags().Int("port", 9188, "HTTP listen port for the exporter")

	RootCmd.AddCommand(ServerCmd)
}
