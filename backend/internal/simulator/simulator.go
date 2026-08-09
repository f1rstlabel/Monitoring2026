// Package simulator provides a Pinger interface and a FakePinger test-helper
// allowing the poller and notification pipeline to be tested without needing
// physical 414 production network devices.
package simulator

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Pinger defines the contract for pinging network targets.
type Pinger interface {
	Ping(ip string) (reachable bool, rtt time.Duration, err error)
}

// RealICMPPinger is the production implementation (stubbed/mockable).
type RealICMPPinger struct{}

func (r *RealICMPPinger) Ping(ip string) (bool, time.Duration, error) {
	// Simple stub for real ping
	return true, 12 * time.Millisecond, nil
}

// FakePinger allows unit tests and CLI simulators to control device states.
type FakePinger struct {
	mu           sync.RWMutex
	deviceStates map[string]bool          // IP -> reachable
	latencies    map[string]time.Duration // IP -> RTT
	failCounts   map[string]int           // IP -> fail count
	concurrent   int                      // current active ping calls
	maxSeenConc  int                      // max observed concurrent calls
}

func NewFakePinger() *FakePinger {
	return &FakePinger{
		deviceStates: make(map[string]bool),
		latencies:    make(map[string]time.Duration),
		failCounts:   make(map[string]int),
	}
}

// SetState sets whether a specific IP is reachable.
func (f *FakePinger) SetState(ip string, reachable bool, rtt time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deviceStates[ip] = reachable
	f.latencies[ip] = rtt
}

// FlipState toggles a device between UP and DOWN.
func (f *FakePinger) FlipState(ip string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	curr := f.deviceStates[ip]
	f.deviceStates[ip] = !curr
	return !curr
}

func (f *FakePinger) Ping(ip string) (bool, time.Duration, error) {
	f.mu.Lock()
	f.concurrent++
	if f.concurrent > f.maxSeenConc {
		f.maxSeenConc = f.concurrent
	}
	f.mu.Unlock()

	defer func() {
		f.mu.Lock()
		f.concurrent--
		f.mu.Unlock()
	}()

	// Simulate network delay
	time.Sleep(2 * time.Millisecond)

	f.mu.RLock()
	defer f.mu.RUnlock()

	reachable, exists := f.deviceStates[ip]
	if !exists {
		// Default reachable with 8ms latency
		return true, 8 * time.Millisecond, nil
	}

	rtt := f.latencies[ip]
	if rtt == 0 {
		rtt = 10 * time.Millisecond
	}

	if !reachable {
		return false, 0, fmt.Errorf("request timeout for %s", ip)
	}

	return true, rtt, nil
}

// GetMaxConcurrent returns the maximum number of concurrent pings observed.
func (f *FakePinger) GetMaxConcurrent() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.maxSeenConc
}

// ResetMaxConcurrent resets the peak concurrency counter.
func (f *FakePinger) ResetMaxConcurrent() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.maxSeenConc = 0
}

// SimulatedCluster generates N fake network devices for CLI simulation.
type SimulatedCluster struct {
	Pinger  *FakePinger
	Devices []SimulatedNode
}

type SimulatedNode struct {
	ID   string
	Name string
	Type string
	IP   string
	MAC  string
}

func NewSimulatedCluster(count int, flakyCount int) *SimulatedCluster {
	pinger := NewFakePinger()
	var nodes []SimulatedNode

	types := []string{"Access Point", "Switch", "Router", "CCTV", "SmartPower"}
	for i := 1; i <= count; i++ {
		t := types[i%len(types)]
		ip := fmt.Sprintf("10.20.%d.%d", i/254+1, i%254+1)
		mac := fmt.Sprintf("00:1a:2b:%02x:%02x:%02x", i/65536, (i/256)%256, i%256)
		name := fmt.Sprintf("Simulated-%s-%03d", t, i)

		nodes = append(nodes, SimulatedNode{
			ID:   fmt.Sprintf("sim-%d", i),
			Name: name,
			Type: t,
			IP:   ip,
			MAC:  mac,
		})

		// Make last flakyCount nodes DOWN initially
		isUp := i <= (count - flakyCount)
		pinger.SetState(ip, isUp, time.Duration(10+rand.Intn(20))*time.Millisecond)
	}

	return &SimulatedCluster{
		Pinger:  pinger,
		Devices: nodes,
	}
}
