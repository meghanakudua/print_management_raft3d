// raft/fsm.go
package raft

import (
	"encoding/json"
	"io"

	"github.com/hashicorp/raft"
)

type RaftFSM struct {
	state *RaftState // this will hold our printers, filaments, print jobs
}

type RaftState struct {
	Printers  map[string]Printer
	Filaments map[string]Filament
	PrintJobs map[string]PrintJob
}

type Printer struct {
	ID      string `json:"id"`
	Company string `json:"company"`
	Model   string `json:"model"`
}

type Filament struct {
	ID                     string `json:"id"`
	Type                   string `json:"type"`
	Color                  string `json:"color"`
	TotalWeightInGrams     int    `json:"total_weight_in_grams"`
	RemainingWeightInGrams int    `json:"remaining_weight_in_grams"`
}

type PrintJob struct {
	ID                 string `json:"id"`
	PrinterID          string `json:"printer_id"`
	FilamentID         string `json:"filament_id"`
	FilePath           string `json:"filepath"`
	PrintWeightInGrams int    `json:"print_weight_in_grams"`
	Status             string `json:"status"` // Queued, Running, Done, Canceled
}

// Commands passed through Raft
type Command struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

func NewFSM() *RaftFSM {
	return &RaftFSM{
		state: &RaftState{
			Printers:  make(map[string]Printer),
			Filaments: make(map[string]Filament),
			PrintJobs: make(map[string]PrintJob),
		},
	}
}

func (f *RaftFSM) Apply(log *raft.Log) interface{} {
	var cmd Command
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		return err
	}

	switch cmd.Action {
	case "add_printer":
		var p Printer
		if err := json.Unmarshal(cmd.Data, &p); err != nil {
			return err
		}
		f.state.Printers[p.ID] = p
	case "add_filament":
		var f1 Filament
		if err := json.Unmarshal(cmd.Data, &f1); err != nil {
			return err
		}
		f.state.Filaments[f1.ID] = f1
	case "add_print_job":
		var pj PrintJob
		if err := json.Unmarshal(cmd.Data, &pj); err != nil {
			return err
		}
		f.state.PrintJobs[pj.ID] = pj
	}

	return nil
}

// Snapshot and Restore for fault tolerance
func (f *RaftFSM) Snapshot() (raft.FSMSnapshot, error) {
	buf, err := json.Marshal(f.state)
	if err != nil {
		return nil, err
	}
	return &fsmSnapshot{state: buf}, nil
}

func (f *RaftFSM) Restore(rc io.ReadCloser) error {
	defer rc.Close()
	var state RaftState
	if err := json.NewDecoder(rc).Decode(&state); err != nil {
		return err
	}
	f.state = &state
	return nil
}

type fsmSnapshot struct {
	state []byte
}

func (s *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	if _, err := sink.Write(s.state); err != nil {
		sink.Cancel()
		return err
	}
	return sink.Close()
}

func (s *fsmSnapshot) Release() {}

func (f *RaftFSM) GetPrinters() map[string]Printer {
	return f.state.Printers
}

func (f *RaftFSM) GetFilaments() map[string]Filament {
	return f.state.Filaments
}

func (f *RaftFSM) GetPrintJobs() map[string]PrintJob {
	return f.state.PrintJobs
}
