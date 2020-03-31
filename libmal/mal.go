package libmal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Mal struct {}

func NewMal() Mal {
	return Mal {}
}

func (m Mal) SearchAnime(name string) (SearchResponse, error) {
	searchResponse := SearchResponse{}

	endpoint := fmt.Sprintf("https://api.jikan.moe/v3/search/anime?q=%s&limit=1", url.QueryEscape(name))
	response, err := http.Get(endpoint)
	if err != nil {
		return searchResponse, fmt.Errorf("unable to contact MAL's API: %w", err)
	}

	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&searchResponse)
	if err != nil {
		return searchResponse, fmt.Errorf("invalid JSON response from MAL's API: %w", err)
	}

	return searchResponse, nil
}