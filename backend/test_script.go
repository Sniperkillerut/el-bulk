package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type ClientRequest struct {
	ID                 string  db:"id"
	CustomerID         *string db:"customer_id"
	CustomerName       string  db:"customer_name"
	CustomerContact    string  db:"customer_contact"
	CardName           string  db:"card_name"
	SetName            *string db:"set_name"
	Details            *string db:"details"
	Quantity           int     db:"quantity"
	TCG                string  db:"tcg"
	Status             string  db:"status"
	CancellationReason *string db:"cancellation_reason,omitempty"
	BountyID           *string db:"bounty_id"
	MatchType          string  db:"match_type"
	ScryfallID         *string db:"scryfall_id"
	OracleID           *string db:"oracle_id"
	ImageURL           *string db:"image_url"
	FoilTreatment      *string db:"foil_treatment"
	CardTreatment      *string db:"card_treatment"
	SetCode            *string db:"set_code"
	CollectorNumber    *string db:"collector_number"
	CreatedAt          string  db:"created_at"
}

func main() {
    dsn := "postgres://elbulk:SXEz9OCjrks6ieq0Dh7YqPY89fwny4iN@localhost:5432/elbulk?sslmode=disable"
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatal("Could not connect to DB")
	}

	// Test ListRequests
	const requestColumns = "id, customer_id, customer_name, customer_contact, card_name, set_name, details, quantity, tcg, status, cancellation_reason, bounty_id, match_type, scryfall_id, created_at"
	requests := []ClientRequest{}
	query := "SELECT " + requestColumns + " FROM client_request ORDER BY created_at DESC"
	err = db.SelectContext(context.Background(), &requests, query)
	if err != nil {
		fmt.Printf("ListRequests error: %v\n", err)
	} else {
		fmt.Printf("ListRequests success: %d requests\n", len(requests))
	}
}
