package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

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
