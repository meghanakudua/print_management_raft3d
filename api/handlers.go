// api/handlers.go
package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/meghanakudua/print_management_raft3d/raft"

	hashiraft "github.com/hashicorp/raft"
)

var raftApplyTimeout = 5 * time.Second

func mustMarshal(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}

// ========== PRINTER HANDLERS ==========

func AddPrinterHandler(rn *raft.RaftNode) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if rn.Raft.State() != hashiraft.Leader {
			http.Error(w, "Only leader can handle writes", http.StatusForbidden)
			return
		}

		var p raft.Printer
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "Invalid printer data", http.StatusBadRequest)
			return
		}

		cmd := raft.Command{
			Action: "add_printer",
			Data:   mustMarshal(p),
		}
		data, _ := json.Marshal(cmd)

		applyFuture := rn.Raft.Apply(data, raftApplyTimeout)
		if err := applyFuture.Error(); err != nil {
			http.Error(w, "Raft apply failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Printer %s added", p.ID)
	}
}

func GetPrintersHandler(rn *raft.RaftNode) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(rn.FSM.GetPrinters())
	}
}

// ========== FILAMENT HANDLERS ==========

func AddFilamentHandler(rn *raft.RaftNode) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if rn.Raft.State() != hashiraft.Leader {
			http.Error(w, "Only leader can handle writes", http.StatusForbidden)
			return
		}

		var f raft.Filament
		if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
			http.Error(w, "Invalid filament data", http.StatusBadRequest)
			return
		}

		cmd := raft.Command{
			Action: "add_filament",
			Data:   mustMarshal(f),
		}
		data, _ := json.Marshal(cmd)

		applyFuture := rn.Raft.Apply(data, raftApplyTimeout)
		if err := applyFuture.Error(); err != nil {
			http.Error(w, "Raft apply failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Filament %s added", f.ID)
	}
}

func GetFilamentsHandler(rn *raft.RaftNode) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(rn.FSM.GetFilaments())
	}
}

// ========== PRINT JOB HANDLERS ==========

func AddPrintJobHandler(rn *raft.RaftNode) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if rn.Raft.State() != hashiraft.Leader {
			http.Error(w, "Only leader can handle writes", http.StatusForbidden)
			return
		}

		var job raft.PrintJob
		if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
			http.Error(w, "Invalid print job data", http.StatusBadRequest)
			return
		}

		cmd := raft.Command{
			Action: "add_print_job",
			Data:   mustMarshal(job),
		}
		data, _ := json.Marshal(cmd)

		applyFuture := rn.Raft.Apply(data, raftApplyTimeout)
		if err := applyFuture.Error(); err != nil {
			http.Error(w, "Raft apply failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Print job %s added", job.ID)
	}
}

func GetPrintJobsHandler(rn *raft.RaftNode) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(rn.FSM.GetPrintJobs())
	}
}

func UpdatePrintJobStatusHandler(rn *raft.RaftNode) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if rn.Raft.State() != hashiraft.Leader {
			http.Error(w, "Only leader can update job status", http.StatusForbidden)
			return
		}

		jobID := chi.URLParam(r, "jobID")
		var update struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			http.Error(w, "Invalid status update", http.StatusBadRequest)
			return
		}

		// manually patch PrintJob status
		job := rn.FSM.GetPrintJobs()[jobID]
		job.Status = update.Status

		cmd := raft.Command{
			Action: "add_print_job", // re-apply entire updated job
			Data:   mustMarshal(job),
		}
		data, _ := json.Marshal(cmd)

		applyFuture := rn.Raft.Apply(data, raftApplyTimeout)
		if err := applyFuture.Error(); err != nil {
			http.Error(w, "Raft apply failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Print job %s updated to %s", jobID, update.Status)
	}
}

// JoinRequest is the payload from a joining node
type JoinRequest struct {
	NodeID string `json:"node_id"`
	Addr   string `json:"addr"`
}

func JoinHandler(rn *raft.RaftNode) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req JoinRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid join request", http.StatusBadRequest)
			return
		}

		f := rn.Raft.AddVoter(hashiraft.ServerID(req.NodeID), hashiraft.ServerAddress(req.Addr), 0, 0)
		if err := f.Error(); err != nil {
			http.Error(w, "Failed to add voter: "+err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("Node %s at %s joined successfully", req.NodeID, req.Addr)
		w.WriteHeader(http.StatusOK)
	}
}

func GetLeaderHandler(rn *raft.RaftNode) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		leader := rn.Raft.Leader()
		if leader == "" {
			http.Error(w, "No leader elected yet", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(leader))
	}
}

func GetRaftConfigHandler(rn *raft.RaftNode) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		future := rn.Raft.GetConfiguration()
		if err := future.Error(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		config := future.Configuration()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config.Servers)
	}
}
