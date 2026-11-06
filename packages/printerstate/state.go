// Package printerstate mirrors Bambu Farm Manager / printer gcode_state values
// and Farm Manager Client attention buckets.
package printerstate

import "fmt"

// GcodeState is the printer job state reported by Farm Manager / MQTT push_status.
// Values match Bambu firmware strings exactly (case-sensitive).
type GcodeState string

const (
	Idle    GcodeState = "IDLE"
	Prepare GcodeState = "PREPARE"
	Running GcodeState = "RUNNING"
	Pause   GcodeState = "PAUSE"
	Finish  GcodeState = "FINISH"
	Failed  GcodeState = "FAILED"
)

// AllGcodeStates is the closed set we accept from FM.
var AllGcodeStates = []GcodeState{Idle, Prepare, Running, Pause, Finish, Failed}

// ParseGcodeState normalizes known aliases and validates.
func ParseGcodeState(s string) (GcodeState, error) {
	switch GcodeState(s) {
	case Idle, Prepare, Running, Pause, Finish, Failed:
		return GcodeState(s), nil
	case "PAUSED": // occasional alias in community dumps
		return Pause, nil
	case "FINISHED":
		return Finish, nil
	case "":
		return Idle, nil
	default:
		return "", fmt.Errorf("unknown gcode_state %q", s)
	}
}

// AttentionBucket mirrors Farm Manager Client status navigation.
type AttentionBucket string

const (
	BucketReadyToStart    AttentionBucket = "ready_to_start"    // IDLE with pending work (caller supplies pending)
	BucketAttentionNeeded AttentionBucket = "attention_needed" // PAUSE, FAILED, or offline
	BucketCompletedTask   AttentionBucket = "completed_task"   // FINISH awaiting bed_clean / confirm removal
	BucketNotInTask       AttentionBucket = "not_in_task"      // IDLE, no pending work
	BucketPrinting        AttentionBucket = "printing"         // PREPARE / RUNNING
)

// BucketFor maps printer online flag + gcode_state (+ optional pending queue) to FM-like nav.
// pendingTask means BambuFarmer/FM has queued work waiting for this idle printer.
func BucketFor(online bool, state GcodeState, pendingTask bool) AttentionBucket {
	if !online {
		return BucketAttentionNeeded
	}
	switch state {
	case Pause, Failed:
		return BucketAttentionNeeded
	case Finish:
		return BucketCompletedTask
	case Prepare, Running:
		return BucketPrinting
	case Idle:
		if pendingTask {
			return BucketReadyToStart
		}
		return BucketNotInTask
	default:
		return BucketAttentionNeeded
	}
}

// Machine tracks one printer's gcode_state with validated transitions.
type Machine struct {
	SN    string
	State GcodeState
}

// NewMachine starts in IDLE (Bambu default when unknown).
func NewMachine(sn string) *Machine {
	return &Machine{SN: sn, State: Idle}
}

// allowed transitions mirror observed farm lifecycles.
// IDLE → PREPARE → RUNNING → FINISH → IDLE
// RUNNING ⇄ PAUSE; RUNNING/PAUSE → FAILED; FAILED → IDLE (after clear/bed_clean)
// FINISH → IDLE (bed_clean / confirm removal)
var allowed = map[GcodeState]map[GcodeState]bool{
	Idle: {
		Prepare: true,
		Running: true, // some firmwares skip PREPARE in reports
		Failed:  true,
	},
	Prepare: {
		Running: true,
		Pause:   true,
		Failed:  true,
		Idle:    true, // cancel during prepare
	},
	Running: {
		Pause:  true,
		Finish: true,
		Failed: true,
		Idle:   true, // abrupt stop
	},
	Pause: {
		Running: true,
		Failed:  true,
		Idle:    true,
		Finish:  true, // rare
	},
	Finish: {
		Idle:   true, // bed_clean / collected
		Failed: true,
	},
	Failed: {
		Idle:    true,
		Prepare: true,
		Running: true,
	},
}

// Transition attempts to move to next. same-state is a no-op success (idempotent reports).
func (m *Machine) Transition(next GcodeState) error {
	if m.State == next {
		return nil
	}
	if allowed[m.State][next] {
		m.State = next
		return nil
	}
	return fmt.Errorf("invalid transition %s → %s for %s", m.State, next, m.SN)
}

// Observe applies a telemetry update: parse, transition if legal, or force-set with reason if FM jumped.
// Farm reports can skip states; we allow Force when soft transition fails so edge stays in sync.
func (m *Machine) Observe(raw string) (from, to GcodeState, forced bool, err error) {
	to, err = ParseGcodeState(raw)
	if err != nil {
		return m.State, m.State, false, err
	}
	from = m.State
	if from == to {
		return from, to, false, nil
	}
	if err := m.Transition(to); err != nil {
		m.State = to
		return from, to, true, nil
	}
	return from, to, false, nil
}

// CanAcceptPrint is true when FM would allow starting a new job on an online printer.
func CanAcceptPrint(online bool, state GcodeState) bool {
	return online && state == Idle
}

// NeedsBedClean is true when operator must confirm removal (FINISH).
func NeedsBedClean(state GcodeState) bool {
	return state == Finish
}

// IsActivePrint is true while filament is (or soon will be) moving.
func IsActivePrint(state GcodeState) bool {
	return state == Prepare || state == Running || state == Pause
}
