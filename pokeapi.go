package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const LOCATION_AREA_DETAILS_URL = "https://pokeapi.co/api/v2/location-area/%s/"
const INSPECT_URL = "https://pokeapi.co/api/v2/pokemon/%s/"

type PokeAPI struct {
	cache *Cache
}

func NewPokeApi() *PokeAPI {
	return &PokeAPI{
		cache: NewCache(10 * time.Minute),
	}
}

type LocationsAreasResponse struct {
	Results []LocationsAreaResults `json:"results"`
	Next    string                 `json:"next"`
}

type LocationsAreaResults struct {
	Name string `json:"name"`
}

func (a *PokeAPI) CachedGet(url string) []byte {
	data, ok := a.cache.Get(url)
	if ok {
		return data
	}

	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)

	a.cache.Add(url, body)

	return body

}

func (a *PokeAPI) GetLocationsAreas(url string) ([]string, string) {
	body := a.CachedGet(url)
	var areas LocationsAreasResponse
	json.Unmarshal(body, &areas)

	names := make([]string, len(areas.Results))
	for i, result := range areas.Results {
		names[i] = result.Name
	}
	return names, areas.Next
}

type PokemonDeails struct {
	Name string `json:"name"`
}

type AreaPokemon struct {
	Pokemon PokemonDeails `json:"pokemon"`
}

type LocationAreasDetails struct {
	Pokemons []AreaPokemon `json:"pokemon_encounters"`
}

func (a *PokeAPI) Explore(area string) []string {
	url := fmt.Sprintf(LOCATION_AREA_DETAILS_URL, area)
	body := a.CachedGet(url)
	var areaDeails LocationAreasDetails
	json.Unmarshal(body, &areaDeails)

	pokemons := make([]string, len(areaDeails.Pokemons))
	for i, pokemon := range areaDeails.Pokemons {
		pokemons[i] = pokemon.Pokemon.Name
	}

	return pokemons
}

type ApiStatDesc struct {
	Name string `json:"name"`
}

type ApiStat struct {
	Value    int         `json:"base_stat"`
	StatDesc ApiStatDesc `json:"stat"`
}

type ApiInspectDetails struct {
	Height int              `json:"height"`
	Weight int              `json:"weight"`
	Stats  []ApiStat        `json:"stats"`
	Types  []ApiPokemonType `json:"types"`
}

type ApiTypeDesc struct {
	Name string `json:"name"`
}

type ApiPokemonType struct {
	TypeDecs ApiTypeDesc `json:"type"`
}

type PokemonStats struct {
	Height int
	Weight int
	Stats  map[string]int
	Types  []string
}

func (a *PokeAPI) Inspect(area string) PokemonStats {
	url := fmt.Sprintf(INSPECT_URL, area)
	body := a.CachedGet(url)
	var details ApiInspectDetails
	json.Unmarshal(body, &details)

	stats := PokemonStats{
		Weight: details.Weight,
		Height: details.Height,
		Stats:  make(map[string]int),
		Types:  make([]string, len(details.Types)),
	}

	for _, stat := range details.Stats {
		stats.Stats[stat.StatDesc.Name] = stat.Value
	}
	for i, t := range details.Types {
		stats.Types[i] = t.TypeDecs.Name
	}

	return stats
}
