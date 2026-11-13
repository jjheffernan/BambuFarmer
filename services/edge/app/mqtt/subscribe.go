package mqtt

import (
	"context"
	"fmt"
	"sync"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/jjheffernan/BambuFarmer/packages/farm"
	"github.com/jjheffernan/BambuFarmer/services/edge/app/farmmanager"
)

// Message is one broker publish.
type Message struct {
	Topic   string
	Payload []byte
}

// Source subscribes to device/+/report when a broker URL and cert material exist.
type Source struct {
	URL      string
	CertDir  string
	Insecure bool
}

// Open returns ErrNotConfigured until both FARM_MQTT_URL and cert files are set.
func Open(url, certDir string, insecure bool) (*Source, error) {
	if url == "" || !farmmanager.CertMaterial(certDir) {
		return nil, farm.ErrNotConfigured
	}
	return &Source{URL: url, CertDir: certDir, Insecure: insecure}, nil
}

// Subscribe implements farm.EventSource. The poll loop still runs.
func (s *Source) Subscribe(ctx context.Context) (<-chan farm.DeviceReport, error) {
	tlsCfg, err := farmmanager.ClientTLS(s.CertDir, s.Insecure)
	if err != nil {
		return nil, err
	}
	opts := pahomqtt.NewClientOptions()
	opts.AddBroker(s.URL)
	opts.SetTLSConfig(tlsCfg)
	opts.SetClientID(fmt.Sprintf("bambufarmer-edge-%d", time.Now().UnixNano()))
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(false)
	opts.SetOrderMatters(false)
	client := pahomqtt.NewClient(opts)
	tok := client.Connect()
	if !tok.WaitTimeout(10 * time.Second) {
		return nil, fmt.Errorf("mqtt: connect timeout")
	}
	if err := tok.Error(); err != nil {
		return nil, err
	}
	in := make(chan Message, 32)
	tok = client.Subscribe("device/+/report", 0, func(_ pahomqtt.Client, m pahomqtt.Message) {
		body := append([]byte(nil), m.Payload()...)
		select {
		case <-ctx.Done():
		case in <- Message{Topic: m.Topic(), Payload: body}:
		default:
		}
	})
	if !tok.WaitTimeout(10 * time.Second) {
		client.Disconnect(200)
		return nil, fmt.Errorf("mqtt: subscribe timeout")
	}
	if err := tok.Error(); err != nil {
		client.Disconnect(200)
		return nil, err
	}
	go func() {
		<-ctx.Done()
		client.Disconnect(200)
	}()
	return Fold(ctx, in), nil
}

// Fold applies report messages in order. A skipped command is ignored.
func Fold(ctx context.Context, in <-chan Message) <-chan farm.DeviceReport {
	out := make(chan farm.DeviceReport, 16)
	go func() {
		defer close(out)
		cache := map[string]farm.DeviceReport{}
		var mu sync.Mutex
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-in:
				if !ok {
					return
				}
				mu.Lock()
				sn, _ := SerialFromTopic(msg.Topic)
				next, err := Apply(cache[sn], msg.Topic, msg.Payload)
				if err == nil {
					cache[next.ExternalSN] = next
				}
				mu.Unlock()
				if err != nil {
					continue
				}
				select {
				case <-ctx.Done():
					return
				case out <- next:
				}
			}
		}
	}()
	return out
}
