package server_test

import (
	"bufio"
	server "graceful-shutdown-server"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	msgCh := make(chan string)
	workerPool := server.NewWorkerPool(msgCh, 10)
	broker := server.NewBroker(msgCh)

	signalCh := make(chan os.Signal)
	s := server.NewServer(workerPool, broker, server.WithServerPort(3001), server.WithSignalChan(signalCh))
	go s.Start()

	c := http.DefaultClient
	req, err := http.NewRequest(http.MethodPost, "http://localhost:3001/schedule-job", nil)
	assert.NoError(t, err)

	jobsStatusCh := make(chan struct{})
	go func() {
		req, err := http.NewRequest(http.MethodGet, "http://localhost:3001/jobs-status", nil)
		assert.NoError(t, err)

		resp, err := c.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		var interrupted bool
		var eventsCount int
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				eventsCount++
			}
			t.Log(line)
			t.Log("eventsCount", eventsCount)
			if eventsCount == 3 && !interrupted {
				signalCh <- os.Interrupt
				interrupted = true
			}

			if eventsCount == 4 {
				assert.Contains(t, line, "event: shutdown")
			}
		}

		jobsStatusCh <- struct{}{}
	}()

	c.Do(req)
	c.Do(req)
	c.Do(req)

	<-jobsStatusCh
}
