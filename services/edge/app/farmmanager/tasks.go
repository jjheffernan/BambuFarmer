package farmmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

// CreateTaskRequest is POST /task (Client "Start Direct Print" body).
// Field types are version-sensitive; f3mf_id must be a string.
type CreateTaskRequest struct {
	TaskPrintModel int              `json:"task_print_model"`
	QueueModelCnt  int              `json:"queue_model_cnt"`
	DevicePool     []string         `json:"device_pool"`
	DevicePool2    []DevicePoolEntry `json:"device_pool2"`
	TaskName       string           `json:"task_name"`
	PrintOption    PrintOption      `json:"print_option"`
	F3MFID         string           `json:"f3mf_id"`
	AMSMapping2    []AMSMapping     `json:"ams_mapping2"`
}

// DevicePoolEntry is one printer target with optional AMS mapping.
type DevicePoolEntry struct {
	DevID       string       `json:"dev_id"`
	AMSMapping2 []AMSMapping `json:"ams_mapping2"`
}

// AMSMapping selects an AMS unit/slot; empty slice means auto-match.
type AMSMapping struct {
	AMSID  int `json:"ams_id"`
	SlotID int `json:"slot_id"`
}

// PrintOption mirrors Client print options (defaults match observed UI).
type PrintOption struct {
	AutoBedLeveling         bool `json:"auto_bed_leveling"`
	FlowDynamicCalibration  bool `json:"flow_dynamic_calibration"`
	Timelapse               bool `json:"timelapse"`
	BedLevelingMode         int  `json:"bed_leveling_mode"`
	FlowDynamicCaliMode     int  `json:"flow_dynamic_cali_mode"`
	NozzleOffsetCaliMode    int  `json:"nozzle_offset_cali_mode"`
}

// CreateTaskResponse is POST /task success wrapper.
type CreateTaskResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    CreateTaskData  `json:"data"`
}

// CreateTaskData holds task ids from POST /task.
type CreateTaskData struct {
	TaskID    json.RawMessage `json:"task_id"`
	SubtaskID json.RawMessage `json:"subtask_id"`
	F3MFID    string          `json:"f3mf_id"`
}

// TaskRef is a normalized create-task result.
type TaskRef struct {
	TaskID    string
	SubtaskID string
	F3MFID    string
}

// DefaultPrintOption returns Client-shaped defaults for direct print.
func DefaultPrintOption() PrintOption {
	return PrintOption{
		AutoBedLeveling:        true,
		FlowDynamicCalibration: true,
		Timelapse:              false,
		BedLevelingMode:        2,
		FlowDynamicCaliMode:    2,
		NozzleOffsetCaliMode:   2,
	}
}

// CreateDirectPrint builds a direct-print POST /task body for one printer.
func CreateDirectPrint(f3mfID, sn, taskName string) CreateTaskRequest {
	if taskName == "" {
		taskName = "bambufarmer"
	}
	return CreateTaskRequest{
		TaskPrintModel: 1,
		QueueModelCnt:  0,
		DevicePool:     []string{sn},
		DevicePool2: []DevicePoolEntry{{
			DevID:       sn,
			AMSMapping2: []AMSMapping{},
		}},
		TaskName:    taskName,
		PrintOption: DefaultPrintOption(),
		F3MFID:      f3mfID,
		AMSMapping2: []AMSMapping{},
	}
}

// CreateTask calls POST /task and returns the farm task id.
func (c *Client) CreateTask(ctx context.Context, req CreateTaskRequest) (TaskRef, error) {
	if req.F3MFID == "" {
		return TaskRef{}, fmt.Errorf("create task: empty f3mf_id")
	}
	if len(req.DevicePool) == 0 {
		return TaskRef{}, fmt.Errorf("create task: empty device_pool")
	}
	var out CreateTaskResponse
	if err := c.doJSON(ctx, "POST", "/task", req, &out); err != nil {
		return TaskRef{}, err
	}
	taskID, err := decodeID(out.Data.TaskID)
	if err != nil || taskID == "" {
		return TaskRef{}, fmt.Errorf("create task: missing task_id in response")
	}
	sub, _ := decodeID(out.Data.SubtaskID)
	return TaskRef{
		TaskID:    taskID,
		SubtaskID: sub,
		F3MFID:    firstNonEmpty(out.Data.F3MFID, req.F3MFID),
	}, nil
}

func decodeID(raw json.RawMessage) (string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return "", nil
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s, nil
	}
	var n json.Number
	if err := json.Unmarshal(raw, &n); err == nil {
		return n.String(), nil
	}
	var i int64
	if err := json.Unmarshal(raw, &i); err == nil {
		return strconv.FormatInt(i, 10), nil
	}
	return "", fmt.Errorf("unsupported id json %s", string(raw))
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
