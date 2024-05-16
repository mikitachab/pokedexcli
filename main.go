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
	Name    string
	Desc    string
	Handler func([]string) error
}

type Commands struct {
	CS map[string]Command
}

func (cs *Commands) Register(c Command) {
	cs.CS[c.Name] = c
}

func main() {
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

	cs := Commands{CS: make(map[string]Command)}

	cs.Register(Command{
		Name: "help",
		Desc: "Prints help message",
		Handler: func([]string) error {
			return cs.printHelp()
		},
	})
	cs.Register(Command{
		Name: "exit",
		Desc: "Exiting pokedex repl",
		Handler: func([]string) error {
			os.Exit(0)
			return nil
		},
	})
	cs.Register(Command{
		Name: "map",
		Desc: "Print next 20 location areas",
		Handler: func([]string) error {
			return printMap(areas)
		},
	})
	cs.Register(Command{
		Name: "mapb",
		Desc: "Print previous 20 location areas",
		Handler: func([]string) error {
			return printBMap(areas)
		},
	})
	cs.Register(Command{
		Name: "explore",
		Desc: "explore <area_name> prints pokemons in <area_name>",
		Handler: func(args []string) error {
			return printExplore(args, api)
		},
	})
	cs.Register(Command{
		Name: "catch",
		Desc: "catch <pokemon name> try to catch pokemon",
		Handler: func(args []string) error {
			return catchPokemon(args, api, catchedPokemons)
		},
	})
	cs.Register(Command{
		Name: "inspect",
		Desc: "inspect <pokemon name> to print pokemon stats",
		Handler: func(args []string) error {
			return printInspect(args, catchedPokemons)
		},
	})
	cs.Register(Command{
		Name: "pokedex",
		Desc: "pokedex",
		Handler: func(args []string) error {
			return printPokemons(catchedPokemons)
		},
	})

	for {
		fmt.Print(PROMT)
		scanner.Scan()

		text := scanner.Text()
		args := cleanInput(text)

		if len(args) == 0 {
			continue
		}

		command := args[0]

		handler, ok := cs.CS[command]
		if !ok {
			fmt.Println("Unknown command:", command)
			continue
		}

		err := handler.Handler(args)
		if err != nil {
			fmt.Printf("something went wrong: %v\n", err)
		}
	}
}

func printPokemons(pokemons map[string]PokemonStats) error {
	fmt.Println("Your Pokedex:")
	for name := range pokemons {
		fmt.Println(" -", name)
	}

	return nil
}

func catchPokemon(args []string, api *PokeAPI, pokemons map[string]PokemonStats) error {
	if len(args) != 2 {
		return errors.New("error: not enough arguments given")
	}
	pokemonName := args[1]

	fmt.Println("Throwing a Pokeball at", pokemonName+"...")

	if rand.IntN(100) > 50 {
		fmt.Println(pokemonName, "escaped!")
		return nil
	}

	stats := api.Inspect(pokemonName)
	fmt.Println(pokemonName, "was caught!")
	pokemons[pokemonName] = stats

	return nil
}

func printInspect(args []string, pokemons map[string]PokemonStats) error {
	if len(args) != 2 {
		return errors.New("error: not enough arguments given")
	}

	name := args[1]
	stats, ok := pokemons[name]

	if !ok {
		fmt.Println("you have not caught that pokemon")
		return nil
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

func printExplore(args []string, api *PokeAPI) error {
	if len(args) != 2 {
		return errors.New("error: not enough arguments given")
	}

	areaName := args[1]

	pokemons := api.Explore(areaName)

	fmt.Println("Exploring", areaName, "...")
	fmt.Println("Found Pokemon:")

	for _, pokemon := range pokemons {
		fmt.Println(" -", pokemon)
	}

	return nil
}

func printMap(areas *LocationAreas) error {
	names, _ := areas.NextAreas()
	for _, name := range names {
		fmt.Println(name)
	}

	return nil
}

func printBMap(areas *LocationAreas) error {
	names, err := areas.BackAreas()
	if err != nil {
		return err
	}

	for _, name := range names {
		fmt.Println(name)
	}

	return nil
}

func (cs *Commands) printHelp() error {
	fmt.Println("\n\nWelcome to the Pokedex!")
	fmt.Println("Usage:")

	for name, command := range cs.CS {
		fmt.Println(name+":", command.Desc)
	}
	fmt.Println()

	return nil
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
