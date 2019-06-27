package main

import (
	"encoding/json"
	"github.com/Comcast/webpa-common/logging"
	"github.com/Comcast/webpa-common/xhttp"
	"github.com/go-kit/kit/log"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Comcast/comcast-bascule/bascule/acquire"
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

	gungnirAddress string
	gungnirAuth    acquire.JWTAcquirer
	xmidtAddress   string
	xmidtAuth      acquire.JWTAcquirer
	wg             sync.WaitGroup
	measures       *Measures
	client         func(req *http.Request) (*http.Response, error)
}

func (confidence *Confidence) handleConfidence(quit chan struct{}, interval time.Duration, getDevice func() (interface{}, error)) {
	defer confidence.wg.Done()
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-quit:
			return
		case <-t.C:
			go func() {
				item, err := getDevice()
				if err != nil {
					return
				}
				device := item.(string)
				logging.Debug(confidence.logger).Log(logging.MessageKey(), "testing new device", "device", device)
				confidence.measures.DeviceSize.Add(-1)
				confidence.handleDevice(device)
			}()
		}
	}
}

func (confidence *Confidence) handleDevice(device string) {
	wg := sync.WaitGroup{}
	wg.Add(2)
	gungnir, xmidt := false, false
	go func() {
		gungnir = confidence.gungnirOnline(device)
		wg.Done()
	}()

	go func() {
		xmidt = confidence.xmidtOnline(device)
		wg.Done()
	}()
	wg.Wait()

	confidence.measures.Completed.Add(1)
	if gungnir == xmidt {
		confidence.measures.Success.Add(1)
		logging.Debug(confidence.logger).Log(logging.MessageKey(), "YAY")
	} else {
		logging.Info(confidence.logger).Log(logging.MessageKey(), "XMiDT and Gungnir don't match", "xmidt-online", xmidt, "gungnir-online", gungnir, "device", device)
	}
}

func (confidence *Confidence) gungnirOnline(device string) bool {
	request, err := http.NewRequest("GET", confidence.gungnirAddress+"/api/v1/device/"+device+"/status", nil)
	if err != nil {
		logging.Error(confidence.logger).Log(logging.ErrorKey(), err, logging.MessageKey(), "failed to create request")
		return false
	}

	auth, err := confidence.gungnirAuth.Acquire()
	if err != nil {
		logging.Error(confidence.logger).Log(logging.ErrorKey(), err, logging.MessageKey(), "failed to get gungnir auth")
		return false
	}
	request.Header.Add("Authorization", auth)

	status, data, err := confidence.doRequest(request)
	if err != nil {
		logging.Error(confidence.logger).Log(logging.ErrorKey(), err, logging.MessageKey(), "failed to complete request")
		return false
	}

	if status != 200 {
		logging.Error(confidence.logger).Log("status", status, logging.MessageKey(), "non 200", "url", request.URL, "auth", request.Header.Get("Authorization"))
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
		logging.Error(confidence.logger).Log(logging.ErrorKey(), err, logging.MessageKey(), "failed to get gungnir auth")
		return false
	}
	request.Header.Add("Authorization", auth)

	status, _, err := confidence.doRequest(request)
	if err != nil {
		logging.Error(confidence.logger).Log(logging.ErrorKey(), err, logging.MessageKey(), "failed to complete request")
		return false
	}

	if status != 200 {
		logging.Error(confidence.logger).Log("status", status, logging.MessageKey(), "non 200")
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
