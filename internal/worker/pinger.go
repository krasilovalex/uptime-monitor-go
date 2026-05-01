package worker

import (
	"context"
	"net/http"
	"time"
)

type Pinger struct {
	client *http.Client
}

func NewPinger() *Pinger {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.DisableKeepAlives = true
	t.MaxIdleConnsPerHost = 100

	return &Pinger{
		client: &http.Client{
			Transport: t,
		},
	}
}

func (p *Pinger) Ping(ctx context.Context, task PingTask) PingResult {
	start := time.Now()

	timeout := time.Duration(task.TimeoutSeconds) * time.Second
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, task.URL, nil)

	res := PingResult{
		TargetID:  task.TargetID,
		CheckedAt: start,
	}

	if err != nil {
		res.IsUp = false
		res.ErrorMessage = ("Failed to create request: " + err.Error())
		res.LatencyMs = int(time.Since(start).Milliseconds())
		return res
	}

	resp, err := p.client.Do(req)
	latency := int(time.Since(start).Microseconds())
	res.LatencyMs = latency

	if err != nil {
		res.IsUp = false
		res.ErrorMessage = err.Error()
		return res
	}

	res.StatusCode = resp.StatusCode
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		res.IsUp = true
	} else {
		res.IsUp = false
		res.ErrorMessage = "Bad status code"
	}

	return res
}
