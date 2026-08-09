package main

import (
	"flag"
	"log"
	"time"

	"govmonitor-it/backend/internal/simulator"
)

func main() {
	deviceCount := flag.Int("devices", 20, "Number of simulated network devices to create")
	flakyCount := flag.Int("flaky", 3, "Number of simulated devices experiencing connection instability")
	flag.Parse()

	log.Println("==========================================================")
	log.Println("🌐 GovMonitor IT — Standalone Device Simulator")
	log.Printf("   Configured: %d devices | %d flaky nodes\n", *deviceCount, *flakyCount)
	log.Println("==========================================================")

	cluster := simulator.NewSimulatedCluster(*deviceCount, *flakyCount)

	log.Printf("[Simulator] Created %d mock infrastructure nodes:", len(cluster.Devices))
	for idx, d := range cluster.Devices {
		if idx < 5 || idx >= len(cluster.Devices)-2 {
			reachable, rtt, _ := cluster.Pinger.Ping(d.IP)
			statusStr := "UP"
			if !reachable {
				statusStr = "DOWN"
			}
			log.Printf("   • [%-4s] %-30s | IP: %-15s | MAC: %s | RTT: %v",
				statusStr, d.Name, d.IP, d.MAC, rtt)
		} else if idx == 5 {
			log.Println("   ... [middle nodes hidden] ...")
		}
	}

	log.Println("\n[Simulator] Starting active polling loop (Press Ctrl+C to stop)...")
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	cycle := 1
	for range ticker.C {
		log.Printf("\n--- Simulation Cycle #%d ---", cycle)
		upCount := 0
		downCount := 0

		for _, d := range cluster.Devices {
			reachable, rtt, err := cluster.Pinger.Ping(d.IP)
			if err != nil || !reachable {
				downCount++
			} else {
				upCount++
				_ = rtt
			}
		}

		log.Printf("[Summary] Total: %d | UP: %d | DOWN: %d | Peak Concurrency: %d workers",
			len(cluster.Devices), upCount, downCount, cluster.Pinger.GetMaxConcurrent())

		// Periodically flip state of a random flaky node to simulate flapping
		if cycle%3 == 0 && *flakyCount > 0 {
			targetIdx := len(cluster.Devices) - 1 - (cycle % *flakyCount)
			target := cluster.Devices[targetIdx]
			newState := cluster.Pinger.FlipState(target.IP)
			stateStr := "RECOVERED (UP)"
			if !newState {
				stateStr = "DROPPED (DOWN)"
			}
			log.Printf("⚠️  [Flap Simulation] Toggled state for %s (%s) → Now %s",
				target.Name, target.IP, stateStr)
		}

		cycle++
	}
}
