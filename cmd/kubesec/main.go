package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	whhttp "github.com/slok/kubewebhook/pkg/http"
	"github.com/slok/kubewebhook/pkg/log"
	"github.com/slok/kubewebhook/pkg/observability/metrics"

	"github.com/controlplaneio/kubesec-webhook/pkg/webhook"
)

// Defaults.
const (
	lAddressDef     = ":8080"
	lMetricsAddress = ":8081"
	debugDef        = false
	gracePeriod     = 3 * time.Second
)

// Flags are the flags of the program.
type Flags struct {
	ListenAddress        string
	MetricsListenAddress string
	Debug                bool
	CertFile             string
	KeyFile              string
	MinScore             int
}

// NewFlags returns the flags of the commandline.
func NewFlags() *Flags {
	flags := &Flags{}
	fl := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fl.StringVar(&flags.ListenAddress, "listen-address", lAddressDef, "webhook server listen address")
	fl.StringVar(&flags.MetricsListenAddress, "metrics-listen-address", lMetricsAddress, "metrics server listen address")
	fl.BoolVar(&flags.Debug, "debug", debugDef, "enable debug mode")
	fl.StringVar(&flags.CertFile, "tls-cert-file", "certs/cert.pem", "TLS certificate file")
	fl.StringVar(&flags.KeyFile, "tls-key-file", "certs/key.pem", "TLS key file")
	fl.IntVar(&flags.MinScore, "min-score", 0, "Kubesec.io minimum score to validate against")

	fl.Parse(os.Args[1:])

	return flags
}

type Main struct {
	flags  *Flags
	logger log.Logger
	stopC  chan struct{}
}

// Run will run the main program.
func (m *Main) Run() error {

	m.logger = &log.Std{
		Debug: m.flags.Debug,
	}

	// Register metrics
	promReg := prometheus.NewRegistry()
	metricsRec := metrics.NewPrometheus(promReg)

	// Create webhooks
	pw, err := webhook.NewPodWebhook(m.flags.MinScore, metricsRec, m.logger)
	if err != nil {
		return err
	}
	pwd, err := whhttp.HandlerFor(pw)
	if err != nil {
		return err
	}
	vdw, err := webhook.NewDeploymentWebhook(m.flags.MinScore, metricsRec, m.logger)
	if err != nil {
		return err
	}
	vdwh, err := whhttp.HandlerFor(vdw)
	if err != nil {
		return err
	}
	dw, err := webhook.NewDaemonSetWebhook(m.flags.MinScore, metricsRec, m.logger)
	if err != nil {
		return err
	}
	dwd, err := whhttp.HandlerFor(dw)
	if err != nil {
		return err
	}
	sw, err := webhook.NewStatefulSetWebhook(m.flags.MinScore, metricsRec, m.logger)
	if err != nil {
		return err
	}
	swd, err := whhttp.HandlerFor(sw)
	if err != nil {
		return err
	}
	errC := make(chan error)

	// Serve webhooks
	go func() {

		m.logger.Infof("webhooks listening on %s...", m.flags.ListenAddress)
		mux := http.NewServeMux()
		mux.Handle("/pod", pwd)
		mux.Handle("/deployment", vdwh)
		mux.Handle("/daemonset", dwd)
		mux.Handle("/statefulset", swd)
		errC <- http.ListenAndServeTLS(
			m.flags.ListenAddress,
			m.flags.CertFile,
			m.flags.KeyFile,
			mux,
		)
	}()

	// Serve metrics.
	metricsHandler := promhttp.HandlerFor(promReg, promhttp.HandlerOpts{})
	go func() {
		m.logger.Infof("metrics listening on %s...", m.flags.MetricsListenAddress)
		errC <- http.ListenAndServe(m.flags.MetricsListenAddress, metricsHandler)
	}()

	// Run everything
	defer m.stop()

	sigC := m.createSignalChan()
	select {
	case err := <-errC:
		if err != nil {
			m.logger.Errorf("error received: %s", err)
			return err
		}
		m.logger.Infof("app finished successfuly")
	case s := <-sigC:
		m.logger.Infof("signal %s received", s)
		return nil
	}

	return nil
}

func (m *Main) stop() {
	m.logger.Infof("stopping everything, waiting %s...", gracePeriod)

	close(m.stopC)

	// Stop everything and let them time to stop.
	time.Sleep(gracePeriod)
}

func (m *Main) createSignalChan() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	return c
}

func main() {
	m := Main{
		flags: NewFlags(),
		stopC: make(chan struct{}),
	}

	err := m.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}
	os.Exit(0)

}
