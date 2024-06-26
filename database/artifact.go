package database

import (
	"context"
	"fmt"

	"github.com/arangodb/go-driver"

	"github.com/sarcb/catalyst-sp24/bus"
	"github.com/sarcb/catalyst-sp24/database/busdb"
	"github.com/sarcb/catalyst-sp24/generated/model"
	"github.com/sarcb/catalyst-sp24/generated/time"
)

func (db *Database) ArtifactGet(ctx context.Context, id int64, name string) (*model.Artifact, error) {
	ticketFilterQuery, ticketFilterVars, err := db.Hooks.TicketWriteFilter(ctx)
	if err != nil {
		return nil, err
	}

	query := `LET d = DOCUMENT(@@collection, @ID) 
	` + ticketFilterQuery + `
	FOR a in NOT_NULL(d.artifacts, [])
	FILTER a.name == @name
	RETURN a`
	cursor, _, err := db.Query(ctx, query, mergeMaps(ticketFilterVars, map[string]any{
		"@collection": TicketCollectionName,
		"ID":          fmt.Sprint(id),
		"name":        name,
	}), busdb.ReadOperation)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()

	var doc model.Artifact
	_, err = cursor.ReadDocument(ctx, &doc)
	if err != nil {
		return nil, err
	}

	return &doc, nil
}

func (db *Database) ArtifactUpdate(ctx context.Context, id int64, name string, artifact *model.Artifact) (*model.TicketWithTickets, error) {
	ticketFilterQuery, ticketFilterVars, err := db.Hooks.TicketWriteFilter(ctx)
	if err != nil {
		return nil, err
	}

	query := `LET d = DOCUMENT(@@collection, @ID)
	` + ticketFilterQuery + `
	FOR a IN NOT_NULL(d.artifacts, [])
	FILTER a.name == @name
	LET newartifacts = APPEND(REMOVE_VALUE(d.artifacts, a), @artifact)
	UPDATE d WITH { "artifacts": newartifacts } IN @@collection
	RETURN NEW`

	return db.ticketGetQuery(ctx, id, query, mergeMaps(map[string]any{
		"@collection": TicketCollectionName,
		"ID":          id,
		"name":        name,
		"artifact":    artifact,
	}, ticketFilterVars), &busdb.Operation{
		Type: bus.DatabaseEntryUpdated,
		Ids: []driver.DocumentID{
			driver.DocumentID(fmt.Sprintf("%s/%d", TicketCollectionName, id)),
		},
	})
}

func (db *Database) EnrichArtifact(ctx context.Context, id int64, name string, enrichmentForm *model.EnrichmentForm) (*model.TicketWithTickets, error) {
	enrichment := model.Enrichment{Created: time.Now().UTC(), Data: enrichmentForm.Data, Name: enrichmentForm.Name}

	ticketFilterQuery, ticketFilterVars, err := db.Hooks.TicketWriteFilter(ctx)
	if err != nil {
		return nil, err
	}

	query := `LET d = DOCUMENT(@@collection, @ID)
	` + ticketFilterQuery + `
	FOR a IN NOT_NULL(d.artifacts, [])
	FILTER a.name == @name
	LET enrichments = NOT_NULL(a.enrichments, {})
	LET newenrichments = MERGE(enrichments, ZIP( [@enrichmentname], [@enrichment]) )
	LET newartifacts = APPEND(REMOVE_VALUE(d.artifacts, a), MERGE(a, { "enrichments": newenrichments }))
	UPDATE d WITH { "artifacts": newartifacts } IN @@collection
	RETURN NEW`

	return db.ticketGetQuery(ctx, id, query, mergeMaps(map[string]any{
		"@collection":    TicketCollectionName,
		"ID":             id,
		"name":           name,
		"enrichmentname": enrichment.Name,
		"enrichment":     enrichment,
	}, ticketFilterVars), &busdb.Operation{
		Type: bus.DatabaseEntryUpdated,
		Ids: []driver.DocumentID{
			driver.DocumentID(fmt.Sprintf("%s/%d", TicketCollectionName, id)),
		},
	})
}
