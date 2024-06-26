package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/arangodb/go-driver"
	"github.com/docker/docker/client"

	"github.com/sarcb/catalyst-sp24/bus"
	"github.com/sarcb/catalyst-sp24/caql"
	"github.com/sarcb/catalyst-sp24/database/busdb"
	"github.com/sarcb/catalyst-sp24/generated/model"
)

func toJob(doc *model.JobForm) *model.Job {
	return &model.Job{
		Automation: doc.Automation,
		Payload:    doc.Payload,
		Origin:     doc.Origin,
		Running:    true,
		Status:     "created",
	}
}

func (db *Database) toJobResponse(ctx context.Context, key string, doc *model.Job, update bool) (*model.JobResponse, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	status := doc.Status

	if doc.Running {
		inspect, err := cli.ContainerInspect(ctx, key)
		if err != nil || inspect.State == nil {
			if update {
				_, _ = db.JobUpdate(ctx, key, &model.JobUpdate{
					Status:  doc.Status,
					Running: false,
				})
			}
		} else if doc.Status != inspect.State.Status {
			status = inspect.State.Status
			if update {
				_, _ = db.JobUpdate(ctx, key, &model.JobUpdate{
					Status:  status,
					Running: doc.Running,
				})
			}
		}
	}

	return &model.JobResponse{
		Automation: doc.Automation,
		ID:         key,
		Log:        doc.Log,
		Payload:    doc.Payload,
		Origin:     doc.Origin,
		Output:     doc.Output,
		Status:     status,
		Container:  doc.Container,
	}, nil
}

func (db *Database) JobCreate(ctx context.Context, id string, job *model.JobForm) (*model.JobResponse, error) {
	if job == nil {
		return nil, errors.New("requires job")
	}

	var doc model.Job
	newctx := driver.WithReturnNew(ctx, &doc)

	meta, err := db.jobCollection.CreateDocument(ctx, newctx, id, toJob(job))
	if err != nil {
		return nil, err
	}

	return db.toJobResponse(ctx, meta.Key, &doc, true)
}

func (db *Database) JobGet(ctx context.Context, id string) (*model.JobResponse, error) {
	var doc model.Job
	meta, err := db.jobCollection.ReadDocument(ctx, id, &doc)
	if err != nil {
		return nil, err
	}

	return db.toJobResponse(ctx, meta.Key, &doc, true)
}

func (db *Database) JobUpdate(ctx context.Context, id string, job *model.JobUpdate) (*model.JobResponse, error) {
	var doc model.Job
	ctx = driver.WithReturnNew(ctx, &doc)

	meta, err := db.jobCollection.UpdateDocument(ctx, id, job)
	if err != nil {
		return nil, err
	}

	return db.toJobResponse(ctx, meta.Key, &doc, true)
}

func (db *Database) JobLogAppend(ctx context.Context, id string, logLine string) error {
	query := `LET d = DOCUMENT(@@collection, @ID)
	UPDATE d WITH { "log": CONCAT(NOT_NULL(d.log, ""), @logline) } IN @@collection`
	cur, _, err := db.Query(ctx, query, map[string]any{
		"@collection": JobCollectionName,
		"ID":          id,
		"logline":     logLine,
	}, &busdb.Operation{
		Type: bus.DatabaseEntryUpdated,
		Ids: []driver.DocumentID{
			driver.DocumentID(fmt.Sprintf("%s/%s", JobCollectionName, id)),
		},
	})
	if err != nil {
		return err
	}
	defer cur.Close()

	return nil
}

func (db *Database) JobComplete(ctx context.Context, id string, out any) error {
	query := `LET d = DOCUMENT(@@collection, @ID)
	UPDATE d WITH { "output": @out, "status": "completed", "running": false } IN @@collection`
	cur, _, err := db.Query(ctx, query, map[string]any{
		"@collection": JobCollectionName,
		"ID":          id,
		"out":         out,
	}, &busdb.Operation{
		Type: bus.DatabaseEntryUpdated,
		Ids: []driver.DocumentID{
			driver.DocumentID(fmt.Sprintf("%s/%s", JobCollectionName, id)),
		},
	})
	if err != nil {
		return err
	}
	defer cur.Close()

	return nil
}

func (db *Database) JobDelete(ctx context.Context, id string) error {
	_, err := db.jobCollection.RemoveDocument(ctx, id)

	return err
}

func (db *Database) JobList(ctx context.Context) ([]*model.JobResponse, error) {
	query := "FOR d IN @@collection RETURN d"
	cursor, _, err := db.Query(ctx, query, map[string]any{"@collection": JobCollectionName}, busdb.ReadOperation)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()
	var docs []*model.JobResponse
	for {
		var doc model.Job
		meta, err := cursor.ReadDocument(ctx, &doc)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, err
		}

		job, err := db.toJobResponse(ctx, meta.Key, &doc, false)
		if err != nil {
			return nil, err
		}

		docs = append(docs, job)
	}

	return docs, err
}

func publishJobMapping(id, automation string, contextStructs *model.Context, origin *model.Origin, payloadMapping map[string]string, db *Database) error {
	msg, err := generatePayload(payloadMapping, contextStructs)
	if err != nil {
		return fmt.Errorf("message generation failed: %w", err)
	}

	db.bus.JobChannel.Publish(&bus.JobMsg{
		ID:         id,
		Automation: automation,
		Origin:     origin,
		Message: &model.Message{
			Context: contextStructs,
			Payload: msg,
		},
	})

	return nil
}

func generatePayload(msgMapping map[string]string, contextStructs *model.Context) (map[string]any, error) {
	contextJSON, err := json.Marshal(contextStructs)
	if err != nil {
		return nil, err
	}

	automationContext := map[string]any{}
	err = json.Unmarshal(contextJSON, &automationContext)
	if err != nil {
		return nil, err
	}

	parser := caql.Parser{}
	msg := map[string]any{}
	for arg, expr := range msgMapping {
		tree, err := parser.Parse(expr)
		if err != nil {
			return nil, err
		}

		v, err := tree.Eval(automationContext)
		if err != nil {
			return nil, err
		}
		msg[arg] = v
	}

	return msg, nil
}
