package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	// here we read from the request context and fetch out `"user"` key set in
	// the MyMiddleware example above.
	rawUser := r.Context().Value(userKey)

	if rawUser != nil {
		user := rawUser.(*User)
		fmt.Fprintf(w, "Hello %s!\n", user.Name)
		return
	}

	fmt.Fprintf(w, "Hello World!\n")
}

func getBatchChangesHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userKey).(*User)

	isAuthorized, err := user.checkNamespaceAccess("BATCHCHANGES", "READ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	if !isAuthorized {
		http.Error(w, "You are not authorized to view batch changes", http.StatusForbidden)
		return
	}

	rows, err := db.Query(`SELECT
	id, name, private, namespace_org_id, namespace_user_id, creator_id
FROM
	batch_changes bc
WHERE
	(bc.namespace_user_id = $1) OR
	(bc.namespace_user_id <> $1 AND bc.private = false) OR
	(bc.private = true AND EXISTS(SELECT 1 FROM batch_changes_namespace bcn WHERE bcn.action = 'WRITE' AND bcn.subject_id = $1 AND bcn.resource_id = bc.id)) OR
	(bc.namespace_org_id IS NOT NULL AND EXISTS (SELECT 1 FROM org_members om WHERE om.org_id = bc.namespace_org_id AND om.user_id = $1))
`, user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer rows.Close()

	bcs := []*batchChange{}
	for rows.Next() {
		var bc batchChange
		err := rows.Scan(&bc.ID, &bc.Name, &bc.Private, &NullInt{N: &bc.NamespaceOrgID}, &NullInt{N: &bc.NamespaceUserID}, &bc.CreatorID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		bcs = append(bcs, &bc)
	}

	render.JSON(w, r, bcs)
	return
}

func shareBatchChange(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userKey).(*User)

	isAuthorized, err := user.checkNamespaceAccess("BATCHCHANGES", "WRITE")
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	if !isAuthorized {
		http.Error(w, "You are not authorized to share batch changes", http.StatusForbidden)
		return
	}

	bcID := chi.URLParam(r, "batchChangeID")
	rUID := chi.URLParam(r, "recipientUserID")
	action := chi.URLParam(r, "action")

	recipientUserID, err := strconv.Atoi(rUID)
	if err != nil {
		http.Error(w, "recipient user id must be an integer", http.StatusBadRequest)
		return
	}

	if user.ID == recipientUserID {
		http.Error(w, "You cannot share a batch change with yourself", http.StatusBadRequest)
		return
	}

	bc := &batchChange{}
	err = db.QueryRow(`SELECT
	id, private
FROM
	batch_changes bc
WHERE id = $2 AND (
	(bc.namespace_user_id = $1) OR
	(bc.namespace_org_id IS NOT NULL AND EXISTS (SELECT 1 FROM org_members om WHERE om.org_id = bc.namespace_org_id AND om.user_id = $1)) OR
	(EXISTS (SELECT 1 FROM batch_changes_namespace bcn WHERE bcn.resource_id = bc.id AND bcn.subject_id = $1 AND bcn.action = 'WRITE'))
)
`, user.ID, bcID).Scan(&bc.ID, &bc.Private)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			http.Error(w, fmt.Sprintf("Batch Change with ID %s does not exist.", bcID), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Sharing batch change %s was unsuccessful.", bcID), http.StatusBadRequest)
		return
	}

	if !bc.Private {
		http.Error(w, fmt.Sprintf("Batch Change with ID %s is public.", bcID), http.StatusBadRequest)
		return
	}

	err = bc.shareResourceAccess(recipientUserID, strings.ToUpper(action))
	if err != nil {
		http.Error(w, "unable ", http.StatusBadRequest)
		return
	}

	http.Error(w, fmt.Sprintf("Sharing batch change %s successful.", bcID), http.StatusOK)
}

func createBatchChange(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userKey).(*User)

	isAuthorized, err := user.checkNamespaceAccess("BATCHCHANGES", "WRITE")
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	if !isAuthorized {
		http.Error(w, "You are not authorized to create batch changes", http.StatusForbidden)
		return
	}

	bc := &batchChange{}
	err = render.Bind(r, bc)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot read request body: %s", err.Error()), http.StatusBadRequest)
		return
	}

	if bc.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	err = db.QueryRow("INSERT INTO batch_changes (name, namespace_user_id, creator_id, private) VALUES ($1, $2, $3, $4) RETURNING namespace_user_id, creator_id, id", bc.Name, user.ID, user.ID, bc.Private).Scan(&bc.NamespaceUserID, &bc.CreatorID, &bc.ID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	render.JSON(w, r, bc)
}

func getBatchChange(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userKey).(*User)

	isAuthorized, err := user.checkNamespaceAccess("BATCHCHANGES", "READ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	if !isAuthorized {
		http.Error(w, "You are not authorized to view batch changes", http.StatusForbidden)
		return
	}

	bcID := chi.URLParam(r, "batchChangeID")

	bc := &batchChange{}
	err = db.QueryRow(`
SELECT
	id, name, private, namespace_org_id, namespace_user_id, creator_id
FROM
	batch_changes bc
WHERE
	bc.id = $1 AND (
		(bc.namespace_user_id = $2) OR
		(bc.namespace_user_id <> $2 AND bc.private = false) OR
		EXISTS(SELECT 1 FROM batch_changes_namespace bcn WHERE bcn.resource_id = bc.id AND bcn.subject_id = $2 AND bcn.action = 'READ') OR
		(bc.namespace_org_id IS NOT NULL AND EXISTS(SELECT 1  FROM org_members WHERE org_id = bc.namespace_org_id AND user_id = $2))
	)
`, bcID, user.ID).Scan(&bc.ID, &bc.Name, &bc.Private, &NullInt{N: &bc.NamespaceOrgID}, &NullInt{N: &bc.NamespaceUserID}, &bc.CreatorID)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			http.Error(w, fmt.Sprintf("Batch Change with ID %s does not exist.", bcID), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Error occurred while fetching batch change: %s", err.Error()), http.StatusBadRequest)
		return
	}

	render.JSON(w, r, bc)
}
