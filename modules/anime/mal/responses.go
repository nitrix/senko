package mal

import "time"

type AnimeId = int

type SearchResponse struct {
	Results []struct {
		Id           AnimeId   `json:"mal_id"`
		Type         string    `json:"type"`
		Title        string    `json:"title"`
		ImageURL     string    `json:"image_url"`
		PageURL      string    `json:"url"`
		Description  string    `json:"synopsis"`
		Score        float64   `json:"score"`
		EpisodeCount int       `json:"episodes"`
		Airing       bool      `json:"airing"`
		StartDate    time.Time `json:"start_date,omitempty"`
		EndDate      time.Time `json:"end_date,omitempty"`
	} `json:"results"`
}
