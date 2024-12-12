package dps

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xackery/aatracker/player"
	"github.com/xackery/aatracker/tracker"
)

type DPS struct {
	parseStart   time.Time
	zone         string
	damageEvents map[string][]DamageEvent
}

type DamageEvent struct {
	Source string
	Target string
	Type   string
	Damage int
	Event  time.Time
	Origin string
}

var (
	instance          *DPS
	meleeDamageRegex  = regexp.MustCompile(`\] (.*) for (.*) points of damage.`)
	directDamageRegex = regexp.MustCompile(`\] (.*) for (.*) points of non-melee damage.`)
	dotDamageRegex    = regexp.MustCompile(`\] (.*) has taken (.*) damage from your (.*).`)
)

func New() (*DPS, error) {
	if instance != nil {
		return nil, fmt.Errorf("dps already exists")
	}
	a := &DPS{
		zone:         "Unknown",
		parseStart:   time.Now(),
		damageEvents: make(map[string][]DamageEvent),
	}

	err := tracker.Subscribe(a.onLine)
	if err != nil {
		return nil, fmt.Errorf("tracker subscribe: %w", err)
	}

	err = tracker.SubscribeToZoneEvent(a.onZone)
	if err != nil {
		return nil, fmt.Errorf("tracker subscribe to zone: %w", err)
	}

	instance = a
	return a, nil
}

func (a *DPS) onLine(event time.Time, line string) {

	a.onMeleeDPS(event, line)
	a.onDirectDamageDPS(event, line)
	a.onDotDamageDPS(event, line)

	a.dumpDPS(event)
}

func (a *DPS) onZone(event time.Time, zoneName string) {
	a.zone = zoneName

	a.dumpDPS(event)
}

func (a *DPS) dumpDPS(event time.Time) {
	//dpsPerSec := float64(a.totalDPSGained) / time.Since(a.parseStart).Seconds()
	//dpsPerHour := dpsPerSec * 3600

	if a.zone == "The Bazaar" {
		return
	}

	if len(a.damageEvents) == 0 {
		//fmt.Println("No damage events to report")
		return
	}

	//fmt.Println(len(a.damageEvents), "events to report")
	type dpsReport struct {
		total    int
		maxMelee int
		maxSpell int
	}

	damageTotals := make(map[string]dpsReport)

	tmpDamageEvents := make(map[string][]DamageEvent)

	for name, dmgEvents := range a.damageEvents {
		for _, dmgEvent := range dmgEvents {

			// skip any events older than 60s
			if event.Sub(dmgEvent.Event).Seconds() > 60 {
				continue
			}

			if len(tmpDamageEvents[name]) == 0 {
				tmpDamageEvents[name] = make([]DamageEvent, 0)
			}

			tmpDamageEvents[name] = append(tmpDamageEvents[name], dmgEvent)
			dps, ok := damageTotals[name]
			if !ok {
				dps = dpsReport{}
			}

			dps.total += dmgEvent.Damage
			if dmgEvent.Origin == "melee" {
				if dps.maxMelee < dmgEvent.Damage {
					dps.maxMelee = dmgEvent.Damage
				}
			} else {
				if dps.maxSpell < dmgEvent.Damage {
					dps.maxSpell = dmgEvent.Damage
				}
			}
			damageTotals[name] = dps
		}
	}

	a.damageEvents = tmpDamageEvents

	//fmt.Println(len(a.damageEvents), "events to report after filtering")

	for name, damageTotal := range damageTotals {
		player.SetPlayerStats(name, int(float64(damageTotal.total)/60), damageTotal.maxMelee, damageTotal.maxSpell)
	}

	err := player.Save()
	if err != nil {
		fmt.Println("Error saving player stats:", err)
	}

}

func (a *DPS) onMeleeDPS(event time.Time, line string) {
	match := meleeDamageRegex.FindStringSubmatch(line)
	if len(match) < 3 {
		return
	}

	amount, err := strconv.Atoi(match[2])
	if err != nil {
		return
	}

	chunk := match[1]

	if strings.Contains(chunk, " was hit ") {
		return
	}

	pos := 0
	pickedAdj := ""
	for _, adj := range adjectives {
		pos = strings.Index(chunk, adj)
		if pos <= 0 {
			continue
		}
		pickedAdj = adj
		break
	}
	if pos <= 0 {
		return
	}

	source := chunk[:pos]
	if strings.EqualFold(source, "you") {
		source = tracker.PlayerName()
	}
	target := chunk[pos+len(pickedAdj):]
	if strings.Contains(source, "`s doppleganger") {
		source = strings.ReplaceAll(source, "`s doppleganger", "")
	}
	damageEvent := DamageEvent{
		Source: source,
		Target: target,
		Type:   strings.TrimSpace(pickedAdj),
		Damage: amount,
		Event:  event,
		Origin: "melee",
	}

	_, ok := a.damageEvents[damageEvent.Source]
	if !ok {
		a.damageEvents[damageEvent.Source] = make([]DamageEvent, 0)
	}

	a.damageEvents[damageEvent.Source] = append(a.damageEvents[damageEvent.Source], damageEvent)
}

func (a *DPS) onDirectDamageDPS(event time.Time, line string) {
	match := directDamageRegex.FindStringSubmatch(line)
	if len(match) < 3 {
		return
	}

	amount, err := strconv.Atoi(match[2])
	if err != nil {
		return
	}

	chunk := match[1]

	pos := 0
	pickedAdj := ""
	for _, adj := range adjectives {
		pos = strings.Index(chunk, adj)
		if pos <= 0 {
			continue
		}
		pickedAdj = adj
		break
	}
	if pos <= 0 {
		return
	}

	source := chunk[:pos]
	target := chunk[pos+len(pickedAdj):]

	if strings.Contains(source, "`s doppleganger") {
		fmt.Println("FO?UND DOPPLE:", source)
		source = strings.ReplaceAll(source, "`s doppleganger", "")
	}

	damageEvent := DamageEvent{
		Source: source,
		Target: target,
		Type:   strings.TrimSpace(pickedAdj),
		Damage: amount,
		Event:  event,
		Origin: "direct",
	}

	_, ok := a.damageEvents[damageEvent.Source]
	if !ok {
		a.damageEvents[damageEvent.Source] = make([]DamageEvent, 0)
	}

	a.damageEvents[damageEvent.Source] = append(a.damageEvents[damageEvent.Source], damageEvent)
}

func (a *DPS) onDotDamageDPS(event time.Time, line string) {
	match := dotDamageRegex.FindStringSubmatch(line)

	if len(match) < 3 {
		return
	}

	amount, err := strconv.Atoi(match[2])
	if err != nil {
		return
	}

	source := tracker.PlayerName()
	target := match[1]
	if strings.Contains(source, "`s doppleganger") {
		source = strings.ReplaceAll(source, "`s doppleganger", "")
	}
	damageEvent := DamageEvent{
		Source: source,
		Target: target,
		Type:   match[3][0 : len(match[3])-2],
		Damage: amount,
		Event:  event,
		Origin: "dot",
	}

	_, ok := a.damageEvents[damageEvent.Source]
	if !ok {
		a.damageEvents[damageEvent.Source] = make([]DamageEvent, 0)
	}

	a.damageEvents[damageEvent.Source] = append(a.damageEvents[damageEvent.Source], damageEvent)
}

var adjectives = []string{
	" mauls ",
	" maul ",
	" bites ",
	" bite ",
	" claws ",
	" claw ",
	" gores ",
	" gore ",
	" stings ",
	" slices ",
	" slice ",
	" sting ",
	" smashes ",
	" smash ",
	" rend ",
	" rends ",
	" slash ",
	" slashes ",
	" punch ",
	" punches ",
	" hit ",
	" hits ",
	" You ",
	" yourself ",
	" YOU ",
	" himself ",
	" herself ",
	" itself ",
	" crush ",
	" crushes ",
	" pierce ",
	" pierces ",
	" kick ",
	" kicks ",
	" strike ",
	" strikes ",
	" backstab ",
	" backstabs ",
	" bash ",
	" bashes ",
}
