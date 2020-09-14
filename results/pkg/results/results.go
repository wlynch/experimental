package results

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"
	pb "github.com/tektoncd/experimental/results/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server implements a Tekton Results server.
type Server struct {
	pb.UnimplementedResultsServer

	db *sql.DB
}

// CreateResult receives CreateTaskRunRequest from clients and save it to local Sqlite Server.
func (s *Server) CreateResult(ctx context.Context, req *pb.CreateResultRequest) (*pb.Result, error) {
	r := req.GetResult()

	r.Name = fmt.Sprintf("%s/results/%s", req.Parent, uuid.New().String())
	r.CreatedTime = timestamppb.Now()

	// serialize data and insert it into database.
	b, err := proto.Marshal(r)
	if err != nil {
		log.Println("taskrun marshaling error: ", err)
		return nil, fmt.Errorf("failed to marshal taskrun: %w", err)
	}

	statement, err := s.db.Prepare("INSERT INTO results (name, blob) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Printf("failed to insert a new taskrun: %v\n", err)
		return nil, fmt.Errorf("failed to insert a new taskrun: %w", err)
	}
	if _, err := statement.Exec(r.Name, b); err != nil {
		log.Printf("failed to execute insertion of a new result: %v\n", err)
		return nil, status.Errorf(status.Code(err), "error storing result: %v", err)
	}
	return r, nil
}

// GetResult received GetResultRequest from users and return a Result back to users
func (s *Server) GetResult(ctx context.Context, req *pb.GetResultRequest) (*pb.Result, error) {
	rows, err := s.db.Query("SELECT taskrunlog FROM taskrun WHERE name = ?", req.GetName())
	if err != nil {
		log.Fatalf("failed to query on database: %v", err)
		return nil, fmt.Errorf("failed to query on a taskrun: %w", err)
	}
	r := &pb.Result{}
	rowNum := 0
	for rows.Next() {
		var b []byte
		rowNum++
		if rowNum >= 2 {
			log.Println("Warning: multiple rows found")
			break
		}
		rows.Scan(&b)
		if err := proto.Unmarshal(b, r); err != nil {
			log.Fatal("unmarshaling error: ", err)
			return nil, fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}
	return r, nil
}

// UpdateResult receives Result and FieldMask from client and uses them to update records in local Sqlite Server.
func (s *Server) UpdateResult(ctx context.Context, req *pb.UpdateResultRequest) (*pb.Result, error) {

	b, err := proto.Marshal(req.GetResult())
	if err != nil {
		log.Println("taskrun marshaling error: ", err)
		return nil, fmt.Errorf("taskrun marshaling error: %w", err)
	}

	// Update the entire row in database based on uid of taskrun.
	statement, err := s.db.Prepare("UPDATE result SET name = ?, namespace = ?, taskrunlog = ? WHERE results_id = ?")
	if err != nil {
		log.Printf("failed to update a existing taskrun: %v\n", err)
		return nil, fmt.Errorf("failed to update a exsiting taskrun: %w", err)
	}
	if _, err := statement.Exec(taskrunMeta.GetName(), taskrunMeta.GetNamespace(), blobData, req.GetResultsId()); err != nil {
		log.Printf("failed to execute update of a new taskrun: %v\n", err)
		return nil, fmt.Errorf("failed to execute update of a new taskrun: %w", err)
	}
	return &pb.TaskRunResult{TaskRun: taskrunFromClient, ResultsId: req.GetResultsId()}, nil
}

// DeleteTaskRun receives DeleteTaskRun request from users and delete TaskRun in local Sqlite Server.
func (s *Server) DeleteTaskRunResult(ctx context.Context, req *pb.DeleteTaskRunRequest) (*empty.Empty, error) {
	statement, err := s.db.Prepare("DELETE FROM taskrun WHERE results_id = ?")
	if err != nil {
		log.Fatalf("failed to create delete statement: %v", err)
		return nil, fmt.Errorf("failed to create delete statement: %w", err)
	}
	results, err := statement.Exec(req.GetResultsId())
	if err != nil {
		log.Fatalf("failed to execute delete statement: %v", err)
		return nil, fmt.Errorf("failed to execute delete statement: %w", err)
	}
	affect, err := results.RowsAffected()
	if err != nil {
		log.Fatalf("failed to retrieve results: %v", err)
		return nil, fmt.Errorf("failed to retrieve results: %w", err)
	}
	if affect == 0 {
		return nil, status.Errorf(codes.NotFound, "TaskRun not found")
	}
	return nil, nil
}
