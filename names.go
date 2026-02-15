package main

import (
	"fmt"
	"math/rand"
)

var adjectives = []string{
	"swift", "fuzzy", "fierce", "ancient", "shadow",
	"crimson", "golden", "silent", "wild", "arcane",
	"brave", "cunning", "dire", "elder", "frost",
	"grim", "iron", "jade", "keen", "lunar",
	"mystic", "noble", "pale", "rune", "storm",
}

var monsters = []string{
	"owlbear", "beholder", "mimic", "basilisk", "dragon",
	"griffon", "hydra", "kobold", "lich", "manticore",
	"naga", "ogre", "phoenix", "quasit", "roc",
	"sphinx", "troll", "umber-hulk", "vampire", "wyvern",
	"aboleth", "bugbear", "chimera", "djinni", "ettin",
}

// GenerateName returns a random adjective-monster name.
func GenerateName() string {
	adj := adjectives[rand.Intn(len(adjectives))]
	mon := monsters[rand.Intn(len(monsters))]
	return fmt.Sprintf("%s-%s", adj, mon)
}
