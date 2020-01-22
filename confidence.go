package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/xmidt-org/webpa-common/logging"
	"github.com/xmidt-org/webpa-common/xhttp"

	"github.com/xmidt-org/bascule/acquire"
)

type Status struct {
	DeviceID          string    `json:"deviceid"`
	State             string    `json:"state"`
	Since             time.Time `json:"since"`
	Now               time.Time `json:"now"`
	LastOfflineReason string    `json:"last_offline_reason"`
}
type Event []struct {
	MsgType         int               `json:"msg_type"`
	Source          string            `json:"source"`
	Dest            string            `json:"dest"`
	TransactionUUID string            `json:"transaction_uuid"`
	ContentType     string            `json:"content_type"`
	Metadata        map[string]string `json:"metadata"`
	Payload         string            `json:"payload"`
	BirthDate       int               `json:"birth_date"`
}

type Confidence struct {
	logger log.Logger

	codexAddress string
	codexAuth    *acquire.RemoteBearerTokenAcquirer
	xmidtAddress string
	xmidtAuth    *acquire.RemoteBearerTokenAcquirer
	wg           sync.WaitGroup
	measures     *Measures
	client       func(req *http.Request) (*http.Response, error)
}

func (confidence *Confidence) handleConfidence(quit chan struct{}, interval time.Duration, getDevice func() interface{}) {
	defer confidence.wg.Done()
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-quit:
			return
		case <-t.C:
			go func() {
				device := getDevice()
				if device == nil {
					return
				}
				logging.Debug(confidence.logger).Log(logging.MessageKey(), "testing new device", "device", device)
				confidence.handleDevice(device.(string))
			}()
		}
	}
}

func (confidence *Confidence) handleDevice(device string) {
	wg := sync.WaitGroup{}
	wg.Add(2)
	state := OfflineState
	codex, xmidt := false, false
	go func() {
		codex = confidence.codexOnline(device)
		wg.Done()
	}()

	go func() {
		xmidt = confidence.xmidtOnline(device)
		if xmidt {
			state = OnlineState
		}
		wg.Done()
	}()
	wg.Wait()

	confidence.measures.Completed.Add(1)
	if codex == xmidt {
		confidence.measures.Success.With(StateLabel, state).Add(1)
		logging.Debug(confidence.logger).Log(logging.MessageKey(), "YAY, devices matched")
	} else {
		confidence.measures.Failure.With(StateLabel, state).Add(1)
		logging.Info(confidence.logger).Log(logging.MessageKey(), "XMiDT and Codex don't match", "xmidt-online", xmidt, "codex-online", codex, "device", device)
	}
}

func (confidence *Confidence) codexOnline(device string) bool {
	request, err := http.NewRequest("GET", confidence.codexAddress+"/api/v1/device/"+device+"/status", nil)
	if err != nil {
		logging.Error(confidence.logger).Log(logging.ErrorKey(), err, logging.MessageKey(), "failed to create request")
		return false
	}

	auth, err := confidence.codexAuth.Acquire()
	if err != nil {
		logging.Error(confidence.logger).Log(logging.ErrorKey(), err, logging.MessageKey(), "failed to get codex auth")
		return false
	}
	request.Header.Add("Authorization", auth)

	status, data, err := confidence.doRequest(request)
	if err != nil {
		logging.Error(confidence.logger).Log(logging.ErrorKey(), err, logging.MessageKey(), "failed to complete request")
		return false
	}

	if status != 200 {
		if status != 404 {
			logging.Error(confidence.logger).Log("status", status, logging.MessageKey(), "non 200", "url", request.URL, "auth", request.Header.Get("Authorization"), "device", device)
		}
		return false
	}

	value := Status{}

	err = json.Unmarshal(data, &value)
	if err != nil {
		logging.Error(confidence.logger).Log(logging.ErrorKey(), err, logging.MessageKey(), "failed to read body")
		return false
	}
	return strings.ToLower(value.State) == "online"
}

func (confidence *Confidence) xmidtOnline(device string) bool {
	request, err := http.NewRequest("GET", confidence.xmidtAddress+"/api/v2/device/"+device+"/stat", nil)
	if err != nil {
		logging.Error(confidence.logger).Log(logging.ErrorKey(), err, logging.MessageKey(), "failed to create request")
		return false
	}
	auth, err := confidence.xmidtAuth.Acquire()
	if err != nil {
		logging.Error(confidence.logger).Log(logging.ErrorKey(), err, logging.MessageKey(), "failed to get codex auth")
		return false
	}
	request.Header.Add("Authorization", auth)

	status, _, err := confidence.doRequest(request)
	if err != nil {
		logging.Error(confidence.logger).Log(logging.ErrorKey(), err, logging.MessageKey(), "failed to complete request")
		return false
	}

	if status != 200 {
		if status != 404 {
			logging.Error(confidence.logger).Log("status", status, logging.MessageKey(), "non 200", "url", request.URL, "auth", request.Header.Get("Authorization"), "device", device)
		}
		return false
	}
	return true
}

func (confidence *Confidence) doRequest(request *http.Request) (int, []byte, error) {
	retryOptions := xhttp.RetryOptions{
		Logger:  confidence.logger,
		Retries: 3,

		// Always retry on failures up to the max count.
		ShouldRetry:       func(error) bool { return true },
		ShouldRetryStatus: func(code int) bool { return false },
	}
	response, err := xhttp.RetryTransactor(retryOptions, confidence.client)(request)
	defer response.Body.Close()
	if err != nil {
		logging.Error(confidence.logger).Log(logging.ErrorKey(), err, logging.MessageKey(), "RetryTransactor failed")
		return 0, []byte{}, err
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logging.Error(confidence.logger).Log(logging.ErrorKey(), err, logging.MessageKey(), "failed to read body")
		return 0, []byte{}, err
	}

	return response.StatusCode, data, nil
}
