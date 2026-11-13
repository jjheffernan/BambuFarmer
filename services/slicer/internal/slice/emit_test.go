package slice

import (
	"context"
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/events"
)

type memBus struct {
	envs []events.Envelope
}

func (m *memBus) Publish(_ context.Context, stream string, env events.Envelope) (string, error) {
	if stream != events.StreamSlice {
		return "", nil
	}
	m.envs = append(m.envs, env)
	return "1-0", nil
}

func TestEmitNilBus(t *testing.T) {
	Emit(context.Background(), nil, "default", Request{Name: "a"}, Result{Name: "a"})
}

func TestEmitQueuedThenOutcome(t *testing.T) {
	bus := &memBus{}
	Emit(context.Background(), bus, "shop", Request{Name: "plate", InputPath: "/in.3mf"}, Result{
		Sliced: true, Name: "plate", OutputPath: "/out.gcode.3mf",
	})
	if len(bus.envs) != 2 {
		t.Fatalf("%d", len(bus.envs))
	}
	if bus.envs[0].Type != events.TypeSliceQueued || bus.envs[1].Type != events.TypeSliceSucceeded {
		t.Fatalf("%s %s", bus.envs[0].Type, bus.envs[1].Type)
	}
	if bus.envs[0].SiteID != "shop" || bus.envs[0].Source != "slicer" {
		t.Fatalf("%+v", bus.envs[0])
	}
	bus = &memBus{}
	Emit(context.Background(), bus, "", Request{Name: "plate"}, Result{Sliced: false, Name: "plate", Reason: "ORCA_BIN unset"})
	if bus.envs[1].Type != events.TypeSliceFailed || bus.envs[1].Payload["reason"] != "ORCA_BIN unset" {
		t.Fatalf("%+v", bus.envs[1])
	}
	if bus.envs[1].SiteID != "default" {
		t.Fatal(bus.envs[1].SiteID)
	}
}
