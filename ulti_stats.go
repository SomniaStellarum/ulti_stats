// Ulti stats: Small program to summarize data from http://www.ultimate-numbers.com/ 
// Individual stats, by game. Stats to provide.
// Goals, Assists, Catches, Drops, Throws, Touches, Throwing Percent, Catching Percent, D's, Points Played,
// O Points Played, D Points Played
// Team stats, by game, and by O Line, D Line
// Scores, Possesions, Turnovers, Scoring Efficiency, D's, Opponents Turns. 
package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
)

type IndStats struct {
	Goals int
	Assists int
	Catches int
	Drops int
	Throws int
	Throwaways int
	ThrowIntoDrop int
	Ds int
	PointsPlayed int
}

func IndHeader(w *csv.Writer) {
	err := w.Write([]string{
		"Name",
		"Goals",
		"Assists",
		"Catches",
		"Drops",
		"Catching Efficiency",
		"Throws",
		"Throwaways",
		"Throwing Efficiency",
		"Throws into Drops",
		"Efficiency incld drops",
		"D's",
		"Points Played",
	})
	if err != nil {
		log.Fatal(err)
	}
	w.Flush()	
}

func (i IndStats) Summary(w *csv.Writer, n string) {
	err := w.Write([]string{
		n,
		fmt.Sprintf("%v", i.Goals),
		fmt.Sprintf("%v", i.Assists),
		fmt.Sprintf("%v", i.Catches),
		fmt.Sprintf("%v", i.Drops),
		fmt.Sprintf("%.2f", 100 * float64(i.Catches)/float64(i.Catches + i.Drops)),
		fmt.Sprintf("%v", i.Throws),
		fmt.Sprintf("%v", i.Throwaways),
		fmt.Sprintf("%.2f", 100 * float64(i.Throws - i.Throwaways)/float64(i.Throws)),
		fmt.Sprintf("%v", i.ThrowIntoDrop),
		fmt.Sprintf("%.2f", 
			100 * float64(i.Throws - i.ThrowIntoDrop - i.Throwaways)/float64(i.Throws)),
		fmt.Sprintf("%v", i.Ds),
		fmt.Sprintf("%v", i.PointsPlayed),
	})
	if err != nil {
		log.Fatal(err)
	}
	w.Flush()
}

func (i IndStats) String() string {
	str := []string{
		fmt.Sprintf("\tGoals: %v, Assists %v",
			i.Goals, i.Assists),
		fmt.Sprintf("\t\tCatches: %v, Drops: %v, Percent: %.2f",
			i.Catches, i.Drops,
			100 * float64(i.Catches)/(float64(i.Catches + i.Drops))),
		fmt.Sprintf("\t\tThrows: %v, Throwaways: %v, Percent: %.2f",
			i.Throws, i.Throwaways,
			100 * float64(i.Throws - i.Throwaways)/float64(i.Throws)),
		fmt.Sprintf("\t\tThrows into Drop: %v, Percent: %.2f",
			i.ThrowIntoDrop,
			100 * float64(i.Throws - i.ThrowIntoDrop - i.Throwaways)/float64(i.Throws)),
		fmt.Sprintf("\t\tD's: %v, Points Played: %v",
			i.Ds, i.PointsPlayed),
		}
	return strings.Join(str, "\n")
}

func FetchPlayer(p map[string]*IndStats, s string) *IndStats {
	i, ok := p[s]
	if !ok {
		p[s] = &IndStats{ }
		i, _ = p[s]
	}
	return i
}

func FetchGame(g map[string]*Game, opp string, tmCode string) *Game {
	i, ok := g[tmCode]
	if !ok {
		g[tmCode] = NewGame(opp)
		i, _ = g[tmCode]
	}
	return i
}

func Summarize(r *csv.Reader, w *csv.Writer) {
	games := make(map[string]*Game)
	for {
		s, err := r.Read()
		if err != nil {
			for k, g := range games {
				w.Write([]string{g.Opponent, k})
				w.Write([]string{"O Line"})
				TeamHeader(w)
				g.OlineT.Summary(w)
				IndHeader(w)
				for n, ply := range g.OlineP {
					ply.Summary(w, n)
				}
				fmt.Println(" ")
				w.Write([]string{"D Line"})
				TeamHeader(w)
				g.DlineT.Summary(w)
				IndHeader(w)
				for n, ply := range g.DlineP {
					ply.Summary(w, n)
				}
				fmt.Println(" ")
			}
			log.Fatal(err)
			return
		}
		g := FetchGame(games, s[2], s[0])
		p := g.DlineP
		t := &g.DlineT
		if IsOLine(s) {
			p = g.OlineP
			t = &g.OlineT
		}
		switch s[8] {
		case "Goal":
			for i := 0; i < 7; i++ {
				player := FetchPlayer(p, s[12 + i])
				player.PointsPlayed++
			}
			if s[7] == "Offense" {
				thrower := FetchPlayer(p, s[9])
				rec := FetchPlayer(p, s[10])
				thrower.Throws++
				thrower.Assists++
				rec.Catches++
				rec.Goals++
				t.Scored++
				t.Possesions++
			} else {
				t.OppScore++
			}
		case "Throwaway":
			if s[7] == "Offense" {
				thrower := FetchPlayer(p, s[9])
				thrower.Throws++
				thrower.Throwaways++
				t.Possesions++
			} else {
				t.OpponentTurns++
			}
		case "Drop":
			thrower := FetchPlayer(p, s[9])
			rec := FetchPlayer(p, s[10])
			thrower.Throws++
			thrower.ThrowIntoDrop++
			rec.Drops++
			t.Possesions++
		case "Catch":
			thrower := FetchPlayer(p, s[9])
			rec := FetchPlayer(p, s[10])
			thrower.Throws++
			rec.Catches++
		case "D":
			def := FetchPlayer(p, s[11])
			def.Ds++
			t.OpponentTurns++
			t.Ds++
		default:
			//fmt.Println(s[8], "Don't know what event that is")
		}
	}
}

func IsOLine(s []string) (b bool) {
	b = false
	for i := 0; i < 7; i++ {
		if s[12 + i] == "*******" { b = true }
	}
	return b
}

type TeamStats struct {
	Scored int
	OppScore int
	Possesions int
	Ds int
	OpponentTurns int
}

func TeamHeader(w *csv.Writer) {
	err := w.Write([]string{
		"Scored",
		"Opponent Scored",
		"Possesions",
		"Efficiency",
		"D's",
		"Opponent Turns",
		"Percent D's",
	})
	if err != nil {
		log.Fatal(err)
	}
	w.Flush()	
}

func (t TeamStats) Summary(w *csv.Writer) {
	err := w.Write([]string{
		fmt.Sprintf("%v", t.Scored),
		fmt.Sprintf("%v", t.OppScore),
		fmt.Sprintf("%v", t.Possesions),
		fmt.Sprintf("%.2f", 100 * float64(t.Scored)/float64(t.Possesions)),
		fmt.Sprintf("%v", t.Ds),
		fmt.Sprintf("%v", t.OpponentTurns),
		fmt.Sprintf("%.2f", 100 * float64(t.Ds)/float64(t.OpponentTurns)),
	})
	if err != nil {
		log.Fatal(err)
	}
	w.Flush()
}

func (t TeamStats) String() string {
	return fmt.Sprintf(`		Score %v-%v, Possesions %v, Efficiency %.2f
		D's %v,	Opponent Turns %v, Percent D's %.2f
`, 		t.Scored, 
		t.OppScore,
		t.Possesions,
		100 * float64(t.Scored)/float64(t.Possesions),
		t.Ds,
		t.OpponentTurns,
		100 * float64(t.Ds)/float64(t.OpponentTurns),
	)
}

type Game struct {
	Opponent string
	OlineT TeamStats
	DlineT TeamStats
	OlineP map[string]*IndStats
	DlineP map[string]*IndStats
}

func NewGame(Opp string) *Game {
	g := new(Game)
	g.Opponent = Opp
	g.OlineP = make(map[string]*IndStats)
	g.DlineP = make(map[string]*IndStats)
	return g
}

func main() {
	f, err := os.Open("stats.csv")
	if err != nil {
		log.Fatal(err)
	}
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	_, err = r.Read()
	if err != nil {
		log.Fatal(err)
	}
	w := csv.NewWriter(os.Stdout)
	Summarize(r, w)
	return
}
