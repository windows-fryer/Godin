package resource

import (
	"encoding/json"
	"net/http"

	"github.com/wordgen/wordgen"
	"golang.org/x/text/language"
)

func GenerateJSONResponse(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		return err
	}

	return nil
}

func GenerateResourceName(format string, count int) string {
	words := []string{
		"Nexus",
		"Haven",
		"Citadel",
		"Collective",
		"Realm",
		"Legion",
		"Order",
		"Syndicate",
		"Den",
		"Sanctuary",
		"Outpost",
		"Conclave",
		"Keep",
		"Dominion",
		"Vanguard",
		"Alliance",
		"Echo",
		"Archive",
		"Lair",
		"Vault",
		"Forge",
		"Crucible",
		"Bastion",
		"Asylum",
		"Assembly",
		"Brotherhood",
		"Enclave",
		"Summit",
		"Temple",
		"Empire",
		"Chamber",
		"Faction",
		"Circle",
		"Guildhall",
		"Coalition",
		"Monastery",
		"Dynasty",
		"Watch",
		"Arcadia",
		"Syndra",
		"Frontier",
		"Refuge",
		"Genesis",
		"Obsidian",
		"Echelon",
		"Accord",
		"Horizon",
		"Sanctum",
		"Paragon",
		"Citadel",
	}

	generator := wordgen.NewGenerator()

	generator.Words = words

	if format == "title" {
		generator.Separator = " "
	} else {
		generator.Separator = "-"
	}

	generator.Casing = format
	generator.Count = count
	generator.Language = language.English

	name, _ := generator.Generate()

	return name
}
