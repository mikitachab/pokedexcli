package main

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"strings"
)

const PROMT = "Pokedex> "
const LOCATION_AREAS_URL = "https://pokeapi.co/api/v2/location-area"

type Command struct {
	Name string
	Desc string
}

func main() {
	commands := map[string]Command{
		"help": {
			Name: "help",
			Desc: "Prints help message",
		},
		"exit": {
			Name: "exit",
			Desc: "Exiting pokedex repl",
		},
		"map": {
			Name: "map",
			Desc: "Print next 20 location areas",
		},
		"mapb": {
			Name: "mapb",
			Desc: "Print previous 20 location areas",
		},
		"explore": {
			Name: "explore",
			Desc: "explore <area_name> prints pokemons in <area_name>",
		},
		"catch": {
			Name: "catch",
			Desc: "catch <pokemon name> try to catch pokemon",
		},
		"inspect": {
			Name: "inspect",
			Desc: "inspect <pokemon name> to print pokemon stats",
		},
	}
	catchedPokemons := make(map[string]PokemonStats)

	scanner := bufio.NewScanner(os.Stdin)
	api := NewPokeApi()
	areas := &LocationAreas{
		api:        api,
		CurrentURL: "",
		NextURL:    LOCATION_AREAS_URL,
		History:    []string{},
		Page:       0,
	}

Loop:
	for {
		fmt.Print(PROMT)
		scanner.Scan()

		text := scanner.Text()
		args := cleanInput(text)

		if len(args) == 0 {
			continue
		}

		command := args[0]

		switch command {
		case "help":
			printHelp(commands)
		case "map":
			printMap(areas)
		case "mapb":
			printBMap(areas)
		case "explore":
			printExplore(args, api)
		case "catch":
			catchPokemon(args, api, catchedPokemons)
		case "inspect":
			printInspect(args, catchedPokemons)
		case "pokedex":
			printPokemons(catchedPokemons)
		case "exit":
			break Loop
		default:
			fmt.Println("Unknown command:", command)
			continue
		}
	}
}

func printPokemons(pokemons map[string]PokemonStats) {
	fmt.Println("Your Pokedex:")
	for name := range pokemons {
		fmt.Println(" -", name)
	}
}

func catchPokemon(args []string, api *PokeAPI, pokemons map[string]PokemonStats) {
	if len(args) != 2 {
		fmt.Println("error: not enough arguments given")
	}
	pokemonName := args[1]

	fmt.Println("Throwing a Pokeball at", pokemonName+"...")

	if rand.IntN(100) > 50 {
		fmt.Println(pokemonName, "escaped!")
		return
	}

	stats := api.Inspect(pokemonName)
	fmt.Println(pokemonName, "was caught!")
	pokemons[pokemonName] = stats
}

func printInspect(args []string, pokemons map[string]PokemonStats) {
	if len(args) != 2 {
		fmt.Println("error: not enough arguments given")
	}

	name := args[1]
	stats, ok := pokemons[name]

	if !ok {
		fmt.Println("you have not caught that pokemon")
		return
	}

	fmt.Println("Name:", name)
	fmt.Println("Weight:", stats.Weight)
	fmt.Println("Height:", stats.Height)

	fmt.Println("Stats:")
	for k, v := range stats.Stats {
		fmt.Println(" -", k, ":", v)
	}

	fmt.Println("Types:")
	for _, t := range stats.Types {
		fmt.Println(" -", t)
	}
}

func printExplore(args []string, api *PokeAPI) {
	if len(args) != 2 {
		fmt.Println("error: not enough arguments given")
	}

	areaName := args[1]

	pokemons := api.Explore(areaName)

	fmt.Println("Exploring", areaName, "...")
	fmt.Println("Found Pokemon:")

	for _, pokemon := range pokemons {
		fmt.Println(" -", pokemon)
	}
}

func printMap(areas *LocationAreas) {
	names, _ := areas.NextAreas()
	for _, name := range names {
		fmt.Println(name)
	}
}

func printBMap(areas *LocationAreas) {
	names, err := areas.BackAreas()
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, name := range names {
		fmt.Println(name)
	}
}

func printHelp(commands map[string]Command) {
	fmt.Println("\n\nWelcome to the Pokedex!")
	fmt.Println("Usage:")

	for name, command := range commands {
		fmt.Println(name+":", command.Desc)
	}
	fmt.Println()
}

func cleanInput(text string) []string {
	return strings.Fields(strings.ToLower(text))
}

type LocationAreas struct {
	api        *PokeAPI
	CurrentURL string
	NextURL    string
	History    []string
	Page       int
}

func (la *LocationAreas) NextAreas() ([]string, error) {
	la.Page++
	names, next := la.api.GetLocationsAreas(la.NextURL)
	if la.CurrentURL != "" {
		la.History = append(la.History, la.CurrentURL)
	}
	la.CurrentURL = la.NextURL
	la.NextURL = next
	return names, nil
}

func (la *LocationAreas) BackAreas() ([]string, error) {
	if len(la.History) == 0 {
		return []string{}, errors.New("no previous areas")
	}
	la.Page--

	url := la.History[len(la.History)-1]
	la.History = la.History[:len(la.History)-1]

	names, next := la.api.GetLocationsAreas(url)

	la.CurrentURL = url
	la.NextURL = next
	return names, nil
}
