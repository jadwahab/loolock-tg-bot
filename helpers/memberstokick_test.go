package helpers_test

import (
	"database/sql"
	"fmt"
	"log"
	"testing"

	_ "github.com/lib/pq" // Add this line to import the pq package

	"github.com/jadwahab/loolock-tg-bot/config"
	"github.com/jadwahab/loolock-tg-bot/db"
	"github.com/jadwahab/loolock-tg-bot/helpers"
)

func TestDetermineMembersToKick(t *testing.T) {
	conn, err := sql.Open("postgres", "postgres://vqoujrosezrtqs:93f96afdbb0e84a56b365db8a737926e58e871e6f06fbc4063c4c5241e707110@ec2-18-202-145-226.eu-west-1.compute.amazonaws.com:5432/d2tu7tu0liv5vo")
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}
	defer conn.Close()
	// Test the connection
	err = conn.Ping()
	if err != nil {
		log.Fatalf("Error pinging database: %q", err)
	}

	dbp := &db.DBParams{
		DB: conn,
	}

	cfg, err := config.LoadConfig("../config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %s", err)
	}
	members, err := dbp.GetCurrentMembers(cfg.Groups[config.TopLockers].ChatID)
	if err != nil {
		log.Printf("Failed to get current members: %s", err)
	}

	membersToKick, err := helpers.DetermineMembersToKick(dbp, members, 50)
	if err != nil {
		t.Errorf("DetermineMembersToKick() failed: %v", err)
	}
	fmt.Println(membersToKick)
}
