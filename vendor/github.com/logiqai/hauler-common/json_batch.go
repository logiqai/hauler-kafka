package hauler_common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type LogiqJsonBatchBackend struct {
	LogiqHost string
	LogiqPort string
}

func (db *LogiqJsonBatchBackend) SendToLogiqSingleBatch(bts []byte) error {
	logiqClient := &http.Client{}
	logiqURL := fmt.Sprintf("https://%s:%s/v1/json_batch", db.LogiqHost, db.LogiqPort)
	logiqClient.Timeout = time.Second * 60
	token, exists := os.LookupEnv("x_api_key")

	req, reqErr := http.NewRequest("POST", logiqURL, bytes.NewBuffer(bts))
	if reqErr != nil {
		logEntry.ContextLogger().Errorf("Error creating request: %v", reqErr)
		return reqErr
	}
	if exists {
		req.Header.Add("Content-type", "application/json")
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	} else {
		req.Header.Add("Content-type", "application/json")
	}
	resp, respErr := logiqClient.Do(req)
	if respErr != nil {
		logEntry.ContextLogger().Errorf("Error sending to logiq: %v", respErr)
		return respErr
	} else {
		body, bodyErr := ioutil.ReadAll(resp.Body)
		if bodyErr != nil {
			logEntry.ContextLogger().Errorf("Error reading response from logiq: %v", bodyErr)
			return bodyErr
		} else {
			logEntry.ContextLogger().Debugf("Response from logiq: %v", string(body))
		}
	}
	return nil
}

func (db *LogiqJsonBatchBackend) PublishBatchSizeLimitOrFlushTimeout(messageBatch []map[string]interface{}) error {
	pbytes, pbytesErr := json.Marshal(messageBatch)
	if pbytesErr != nil {
		logEntry.ContextLogger().Errorf("Error marshalling payload: %v", pbytesErr)
		return pbytesErr
	} else {
		if stdout, stdoutErr := os.LookupEnv("STDOUT"); stdoutErr && stdout == "true" {
			fmt.Printf("%s", string(pbytes))
			return nil
		} else {
			if errSend := db.SendToLogiqSingleBatch(pbytes); errSend != nil {
				logEntry.ContextLogger().Errorf("Error sending to logiq: %v", errSend)
				return errSend
			} else {
				return nil
			}
		}
	}
}
