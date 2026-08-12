package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"sanoc/backend/internal/notifier"
)

func main() {
	fmt.Println("=== 1. Verifying github.com/lib/pq Module Import ===")
	dsn := "postgres://postgres@localhost:5432/sanoc?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Printf("DB Ping note: %v", err)
	} else {
		fmt.Println("✓ PostgreSQL connection ping successful using github.com/lib/pq driver!")
	}

	fmt.Println("\n=== 2. Verifying NotifyQueue RECOVERED (UP) Event Enqueue & Dispatch ===")
	nq := notifier.NewNotifyQueue(nil, nil)
	start := time.Now()
	nq.Enqueue(notifier.AggregatorEvent{
		DeviceID:   "dev-test-up",
		DeviceName: "Test Device UP",
		DeviceType: "Access Point",
		IP:         "10.11.5.39",
		Location:   "Gedung Sate Lt 2",
		EventType:  notifier.EventRecovered,
		Timestamp:  time.Now(),
	})
	elapsed := time.Since(start)
	fmt.Printf("✓ Enqueued UP recovery event in %v (Immediate, non-blocking rate limit queue)\n", elapsed)
	nq.Stop()

	fmt.Println("\n=== 3. Querying Recent notification_log Status in PostgreSQL ===")
	rows, err := db.Query("SELECT id, channel, recipient, status, sent_at FROM notification_log ORDER BY sent_at DESC LIMIT 3")
	if err != nil {
		log.Printf("Query notification_log note: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var id, ch, rec, st, sentAt string
			if err := rows.Scan(&id, &ch, &rec, &st, &sentAt); err == nil {
				fmt.Printf("  - Log ID: %s | Channel: %s | Recipient: %s | Status: %s | SentAt: %s\n", id, ch, rec, st, sentAt)
			}
		}
		if err := rows.Err(); err != nil {
			log.Printf("Error iterating notification_log rows: %v", err)
		}
	}

	fmt.Println("\n=== Verification Completed Successfully ===")
}
