package arcade

// Game is a code-defined minigame: a stable key, a human name, and the plausible
// upper bound its score may not exceed (server-side anti-cheat). The roster lives
// in code (not the DB) so it is fixed per release; adding a game is a code change.
type Game struct {
	Key      string
	Name     string
	MaxScore int64
}

// roster is the code-defined game set. The delivered design ships a 4-game
// roster with two fully playable (catch-the-scoop, scoop-match); their keys +
// bounds are here. Add the remaining games when their mechanics are designed.
var roster = map[string]Game{
	"catch-the-scoop": {Key: "catch-the-scoop", Name: "Catch the Scoop", MaxScore: 100000},
	"scoop-match":     {Key: "scoop-match", Name: "Scoop Match", MaxScore: 100000},
}

// GameFor returns the game for key, and whether it exists.
func GameFor(key string) (Game, bool) {
	g, ok := roster[key]
	return g, ok
}

// Games returns the full roster (registration order not guaranteed).
func Games() []Game {
	out := make([]Game, 0, len(roster))
	for _, g := range roster {
		out = append(out, g)
	}
	return out
}
