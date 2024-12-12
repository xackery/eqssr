package player

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/xackery/aatracker/tracker"
)

var (
	mux            sync.Mutex
	players        = make(map[string]Player)
	trackedPlayers = []string{}
)

type Player struct {
	Name         string
	Class        string
	DPS60Sec     int
	MaxMeleeCrit int
	MaxSpellCrit int
	LastUpdate   time.Time
	IsDPSEnabled bool
}

func New() error {

	trackedPlayers = []string{"Shin", "Gromzek", "Ashrem", "Falendar"}
	players["Shin"] = Player{Name: "Shin", Class: "NEC/WIZ/WAR", IsDPSEnabled: true}
	players["Gromzek"] = Player{Name: "Gromzek", Class: "PAL/BER/ENC", IsDPSEnabled: false}
	players["Ashrem"] = Player{Name: "Ashrem", Class: "PAL/BER/MNK", IsDPSEnabled: false}
	players["Falendar"] = Player{Name: "Falendar", Class: "PAL/BRD/MNK", IsDPSEnabled: false}

	err := tracker.SubscribeToZoneEvent(onZone)
	if err != nil {
		return fmt.Errorf("tracker subscribe to zone: %w", err)
	}

	for _, tp := range trackedPlayers {

		w, err := os.Create(tp + ".txt")
		if err != nil {
			return fmt.Errorf("create: %w", err)
		}
		defer w.Close()

		p := players[tp]
		w.WriteString(fmt.Sprintf("%s\n(%s)\n", p.Name, p.Class))
		if p.IsDPSEnabled {
			w.WriteString(fmt.Sprintf("%d DPS/60s", p.DPS60Sec))
		}
	}

	return nil
}

func onZone(event time.Time, zoneName string) {
	mux.Lock()
	defer mux.Unlock()

	// flush players
	for _, p := range players {
		if p.LastUpdate.Before(event) {
			delete(players, p.Name)
		}
	}

}

func isTrackedPlayer(name string) bool {

	for _, p := range trackedPlayers {
		if p == name {
			return true
		}
	}

	return false
}

func SetPlayerStats(name string, dps60Sec int, maxMeleeCrit int, maxSpellCrit int) {
	mux.Lock()
	defer mux.Unlock()

	p, ok := players[name]
	if !ok {
		p = Player{}
	}
	p.Name = name
	p.DPS60Sec = dps60Sec
	if p.MaxMeleeCrit < maxMeleeCrit {
		p.MaxMeleeCrit = maxMeleeCrit
	}
	if p.MaxSpellCrit < maxSpellCrit {
		p.MaxSpellCrit = maxSpellCrit
	}
	p.LastUpdate = time.Now()
	players[name] = p
}

func Save() error {
	mux.Lock()
	defer mux.Unlock()

	for _, p := range players {
		if !isTrackedPlayer(p.Name) {
			continue
		}

		w, err := os.Create(p.Name + ".txt")
		if err != nil {
			return fmt.Errorf("create: %w", err)
		}
		defer w.Close()

		w.WriteString(fmt.Sprintf("%s\n(%s)\n", p.Name, p.Class))
		if p.IsDPSEnabled {
			w.WriteString(fmt.Sprintf("%d DPS/60s", p.DPS60Sec))
		}
	}

	return nil
}
