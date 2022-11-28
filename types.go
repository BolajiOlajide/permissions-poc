package main

import (
	"net/http"
)

type batchChange struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Private         bool   `json:"private"`
	NamespaceUserID int    `json:"namespace_user_id"`
	NamespaceOrgID  int    `json:"namespace_org_id"`
	CreatorID       int    `json:"creator_id"`
}

func (b *batchChange) shareResourceAccess(recipientUserID int, action string) error {
	var pint int
	err := db.QueryRow("INSERT INTO batch_changes_namespace (resource_id, user_id, action) VALUES ($1, $2, $3) RETURNING id", b.ID, recipientUserID, action).Scan(&pint)

	return err
}

func (b *batchChange) Bind(r *http.Request) error {
	// https://stackoverflow.com/questions/44663496/chi-empty-http-request-body-in-render-bind#answer-44663794
	return nil
}

type notebook struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Content   string `json:"content"`
	Private   bool   `json:"private"`
	CreatorID int    `json:"creator_id"`
}

func (n *notebook) shareResourceAccess(recipientUserID int, action string) error {
	var pint int
	err := db.QueryRow("INSERT INTO notebooks_namespace (resource_id, user_id, action) VALUES ($1, $2, $3) RETURNING id", n.ID, recipientUserID, action).Scan(&pint)

	return err
}

type codeinsight struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	UserID int    `json:"user_id"`
}

func (c *codeinsight) shareResourceAccess(recipientUserID int, action string) error {
	var pint int
	err := db.QueryRow("INSERT INTO code_insights_namespace (resource_id, user_id, action) VALUES ($1, $2, $3) RETURNING id", c.ID, recipientUserID, action).Scan(&pint)

	return err
}
