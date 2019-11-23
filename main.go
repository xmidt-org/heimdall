package main

import (
	"crypto/tls"
	"fmt"
	"github.com/xmidt-org/codex-db/cassandra"
	"io"
	"net/http"
	"os/signal"
	"runtime"
	"strings"
	"sync"

	"github.com/goph/emperror"
	"github.com/prometheus/common/route"
	"github.com/xmidt-org/bascule/acquire"
	"github.com/xmidt-org/heimdall/shuffle"
	"github.com/xmidt-org/webpa-common/concurrent"
	"github.com/xmidt-org/webpa-common/logging"
	"github.com/xmidt-org/webpa-common/server"

	"os"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	_ "net/http/pprof"
)

const (
	applicationName, apiBase = "heimdall", "/api/v1"
	DEFAULT_KEY_ID           = "current"
)

var (
	GitCommit = "undefined"
	Version   = "undefined"
	BuildTime = "undefined"
)

type deviceGetter interface {
	GetDeviceList(time.Time, time.Time, int, int) ([]string, error)
}

type StatusConfig struct {
	Db           cassandra.Config
	CodexAddress string
	CodexSAT     acquire.RemoteBearerTokenAcquirerOptions
	XmidtAddress string
	XmidtSAT     acquire.RemoteBearerTokenAcquirerOptions
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
		logger, metricsRegistry, codex, err = server.Initialize(applicationName, arguments, f, v, cassandra.Metrics, Metrics)
	)

	printVer := f.BoolP("version", "v", false, "displays the version number")
	if versionParseErr := f.Parse(arguments); versionParseErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse arguments: %s\n", versionParseErr.Error())
		return 1
	}

	if *printVer {
		printVersionInfo(os.Stdout)
		return 0
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to initialize viper: %s\n", err.Error())
		return 1
	}

	logging.Info(logger).Log(logging.MessageKey(), "Successfully loaded config file", "configurationFile", v.ConfigFileUsed())

	config := new(StatusConfig)
	v.Unmarshal(config)

	dbConn, err := cassandra.CreateDbConnection(config.Db, metricsRegistry, nil)
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

	codexAuth, err := acquire.NewRemoteBearerTokenAcquirer(config.CodexSAT)
	if err != nil {
		logging.Error(logger, emperror.Context(err)...).Log(logging.MessageKey(), "Failed to setup codex Remote Bearer Token Acquirer",
			logging.ErrorKey(), err.Error())
		fmt.Fprintf(os.Stderr, "codex Remote Bearer Token Acquirer Initialize Failed: %#v\n", err)
		return 2
	}

	xmidtAuth, err := acquire.NewRemoteBearerTokenAcquirer(config.CodexSAT)
	if err != nil {
		logging.Error(logger, emperror.Context(err)...).Log(logging.MessageKey(), "Failed to setup xmidt Remote Bearer Token Acquirer",
			logging.ErrorKey(), err.Error())
		fmt.Fprintf(os.Stderr, "xmdit Remote Bearer Token Acquirer Initialize Failed: %#v\n", err)
		return 2
	}

	confidence := Confidence{
		codexAddress: config.CodexAddress,
		codexAuth:    codexAuth,
		xmidtAddress: config.XmidtAddress,
		xmidtAuth:    xmidtAuth,
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
	go populate(dbConn, incoming, stopPopulate, populateWG, confidence.measures)

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

func printVersionInfo(writer io.Writer) {
	fmt.Fprintf(writer, "%s:\n", applicationName)
	fmt.Fprintf(writer, "  version: \t%s\n", Version)
	fmt.Fprintf(writer, "  go version: \t%s\n", runtime.Version())
	fmt.Fprintf(writer, "  built time: \t%s\n", BuildTime)
	fmt.Fprintf(writer, "  git commit: \t%s\n", GitCommit)
	fmt.Fprintf(writer, "  os/arch: \t%s/%s\n", runtime.GOOS, runtime.GOARCH)
}

func main() {
	os.Exit(start(os.Args))
}

func populate(conn deviceGetter, input chan interface{}, stop chan struct{}, wg sync.WaitGroup, measures *Measures) {
	defer wg.Done()
	for {
		select {
		case <-stop:
			close(input)
			return
		default:
			beginTime := time.Now().Add(-time.Hour * 12)
			endTime := beginTime.Add(time.Minute)
			limit := 100
			offset := 0
			for {

				list, err := conn.GetDeviceList(beginTime, endTime, offset, limit)
				if len(list) == 0 {
					fmt.Fprintln(os.Stderr, "list is empty")
					break
				} else if err != nil {
					fmt.Fprintf(os.Stderr, "%s", err.Error())
					break
				}
				for _, elem := range list {
					if strings.HasPrefix(elem, "mac") {
						input <- elem
						measures.DeviceSize.Add(1)
					}
				}
				offset += limit
			}
		}
	}
}
