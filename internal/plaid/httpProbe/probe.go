package httpProbe

import (
	"context"
	"fmt"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
	"net/http"
	"time"
)

type probeResult struct {
	id      resources.Meta
	success bool
	problem error
}

type probe struct {
	id       resources.Meta
	period   *time.Ticker
	host     string
	port     uint16
	resource string
	notify   chan probeResult
}

func (p *probe) Serve(ctx context.Context) error {
	for {
		select {
		case _, ok := <-p.period.C:
			if !ok {
				return nil
			}
			if err := p.probe(ctx); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (p *probe) probe(ctx context.Context) error {
	url := fmt.Sprintf("http://%s:%d%s", p.host, p.port, p.resource)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		p.notify <- probeResult{
			id:      p.id,
			success: false,
			problem: err,
		}
		return nil
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		p.notify <- probeResult{
			id:      p.id,
			success: false,
			problem: err,
		}
		return nil
	}
	success := 200 <= resp.StatusCode && resp.StatusCode < 300
	p.notify <- probeResult{
		id:      p.id,
		success: success,
		problem: nil,
	}
	return nil
}
