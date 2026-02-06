package main

import (
	"context"
	"log/slog"
	"net"
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
		ctx := cmd.Context()

		logger := log.MustLogger(ctx)

		address, err := cmd.Flags().GetString("listen-address")
		if err != nil {
			return err
		}

		timeout, err := cmd.Flags().GetDuration("http-client-timeout")
		if err != nil {
			return err
		}

		serveMux := http.NewServeMux()
		// /probe implements the multi-target exporter pattern. It expects a GET
		// parameter "target" containing the address of the Hub to probe.
		serveMux.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
			target := r.URL.Query().Get("target")
			if target == "" {
				http.Error(w, "missing 'target' parameter", http.StatusBadRequest)
				return
			}

			ctx := r.Context()

			hubExporter := exporter.NewHubExporter(ctx, target, &http.Client{Timeout: timeout})

			registry := prometheus.NewRegistry()
			registry.MustRegister(hubExporter)

			handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
			handler.ServeHTTP(w, r)
		})

		server := &http.Server{
			Addr: address,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx, logger := log.MustWithAttrs(
					r.Context(),
					"method", r.Method,
					"url", r.URL,
					"proto", r.Proto,
					"host", r.Host,
					"remote_addr", r.RemoteAddr,
				)
				logger.Info("Serving request")
				serveMux.ServeHTTP(w, r.Clone(ctx))
			}),
			ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
			BaseContext: func(listener net.Listener) context.Context {
				listenerCtx, _ := log.MustWithAttrs(ctx, "address", listener.Addr())
				return listenerCtx
			},
			ConnContext: func(ctx context.Context, c net.Conn) context.Context {
				connCtx, _ := log.MustWithAttrs(
					ctx,
					"local_address", c.LocalAddr(),
					"remote_address", c.RemoteAddr(),
				)
				return connCtx
			},
		}
		logger.Info("Starting server", "address", address)
		return server.ListenAndServe()
	}),
}

func init() {
	ServerCmd.Flags().String("listen-address", ":9188", "HTTP listen port for the exporter")
	ServerCmd.Flags().Duration("http-client-timeout", 5*time.Second, "HTTP client timeount")

	RootCmd.AddCommand(ServerCmd)
}
