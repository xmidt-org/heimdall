package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os/signal"
	"strings"
	"sync"

	"github.com/Comcast/codex/db/postgresql"

	"github.com/Comcast/codex-heimdall/shuffle"
	"github.com/Comcast/comcast-bascule/bascule/acquire"
	"github.com/Comcast/webpa-common/concurrent"
	"github.com/Comcast/webpa-common/logging"
	"github.com/Comcast/webpa-common/server"
	"github.com/goph/emperror"
	"github.com/prometheus/common/route"

	"os"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	_ "net/http/pprof"
)

const (
	applicationName, apiBase = "heimdall", "/api/v1"
	DEFAULT_KEY_ID           = "current"
	applicationVersion       = "0.1.3"
)

type deviceGetter interface {
	GetDeviceList(string, int) ([]string, error)
}

type StatusConfig struct {
	Db           postgresql.Config
	CodexAddress string
	CodexSAT     acquire.JWTAcquirerOptions
	XmidtAddress string
	XmidtSAT     acquire.JWTAcquirerOptions
	ChannelSize  uint64
	MaxPoolSize  int
	Sender       SenderConfig

	// Rate is the number of devices to check
	Rate int

	// Tick is the time unit for the Rate field.  If Rate is set but this field is not set,
	// a tick of 1 second is used as the default.
	Tick time.Duration
}

type SenderConfig struct {
	ClientTimeout         time.Duration
	ResponseHeaderTimeout time.Duration
	IdleConnTimeout       time.Duration
	DeliveryRetries       int
	DeliveryInterval      time.Duration
}

func start(arguments []string) int {
	start := time.Now()

	var (
		f, v                                = pflag.NewFlagSet(applicationName, pflag.ContinueOnError), viper.New()
		logger, metricsRegistry, codex, err = server.Initialize(applicationName, arguments, f, v, postgresql.Metrics, Metrics)
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to initialize viper: %s\n", err.Error())
		return 1
	}

	printVer := f.BoolP("version", "v", false, "displays the version number")
	if err := f.Parse(arguments); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse arguments: %s\n", err.Error())
		return 1
	}

	if *printVer {
		fmt.Println(applicationVersion)
		return 0
	}

	logging.Info(logger).Log(logging.MessageKey(), "Successfully loaded config file", "configurationFile", v.ConfigFileUsed())

	config := new(StatusConfig)
	v.Unmarshal(config)

	dbConn, err := postgresql.CreateDbConnection(config.Db, metricsRegistry, nil)
	if err != nil {
		logging.Error(logger, emperror.Context(err)...).Log(logging.MessageKey(), "Failed to initialize database connection",
			logging.ErrorKey(), err.Error())
		fmt.Fprintf(os.Stderr, "Database Initialize Failed: %#v\n", err)
		return 2
	}
	stopConfidence := make(chan struct{}, 1)
	stopPopulate := make(chan struct{}, 1)
	populateWG := sync.WaitGroup{}

	tr := &http.Transport{
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		ResponseHeaderTimeout: config.Sender.ResponseHeaderTimeout,
		IdleConnTimeout:       config.Sender.IdleConnTimeout,
	}

	fmt.Println(config.MaxPoolSize)

	confidence := Confidence{
		codexAddress: config.CodexAddress,
		codexAuth:    acquire.NewJWTAcquirer(config.CodexSAT),
		xmidtAddress: config.XmidtAddress,
		xmidtAuth:    acquire.NewJWTAcquirer(config.XmidtSAT),
		logger:       logger,
		measures:     NewMeasures(metricsRegistry),
		client: (&http.Client{
			Transport: tr,
			Timeout:   config.Sender.ClientTimeout,
		}).Do,
	}
	confidence.wg.Add(1)
	populateWG.Add(1)
	incoming, getDevice := shuffle.NewStreamShuffler(config.MaxPoolSize, int(config.ChannelSize))
	go populate(dbConn, incoming, stopPopulate, populateWG, confidence.measures, config.MaxPoolSize)

	// fix interval
	if config.Tick <= 0 {
		config.Tick = time.Second
	}
	if config.Rate <= 0 {
		config.Rate = 5
	}
	interval := config.Tick / time.Duration(config.Rate)

	go confidence.handleConfidence(stopConfidence, interval, getDevice)

	_, runnable, done := codex.Prepare(logger, nil, metricsRegistry, route.New())

	waitGroup, shutdown, err := concurrent.Execute(runnable)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to start device manager: %s\n", err)
		return 1
	}

	logging.Info(logger).Log(logging.MessageKey(), fmt.Sprintf("%s is up and running!", applicationName), "elapsedTime", time.Since(start))
	signals := make(chan os.Signal, 10)
	signal.Notify(signals)
	for exit := false; !exit; {
		select {
		case s := <-signals:
			if s != os.Kill && s != os.Interrupt {
				logging.Info(logger).Log(logging.MessageKey(), "ignoring signal", "signal", s)
			} else {
				logging.Error(logger).Log(logging.MessageKey(), "exiting due to signal", "signal", s)
				exit = true
			}
		case <-done:
			exit = true
		}
	}

	stopPopulate <- struct{}{}
	stopConfidence <- struct{}{}
	close(shutdown)
	confidence.wg.Wait()
	populateWG.Wait()
	waitGroup.Wait()
	err = dbConn.Close()
	if err != nil {
		logging.Error(logger, emperror.Context(err)...).Log(logging.MessageKey(), "closing database threads failed",
			logging.ErrorKey(), err.Error())
	}

	return 0
}

func main() {
	os.Exit(start(os.Args))
}

func populate(conn deviceGetter, input chan interface{}, stop chan struct{}, wg sync.WaitGroup, measures *Measures, maxPoolSize int) {
	defer wg.Done()
	lastID := "mac:"
	for {
		select {
		case <-stop:
			close(input)
			return
		default:
			list, err := conn.GetDeviceList(lastID, 100)
			if len(list) != 0 {
				lastID = list[len(list)-1]
			} else if len(list) == 0 && err == nil {
				lastID = "mac:"
			} else if err != nil {
				fmt.Fprintf(os.Stderr, "%s", err.Error())
			}
			for _, elem := range list {
				if strings.HasPrefix(elem, "mac") {
					input <- elem
					measures.DeviceSize.Add(1)
				}
			}
		}
	}
}
