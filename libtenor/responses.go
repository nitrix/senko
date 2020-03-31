package libtenor

type randomGifResponse struct {
	URL string `json:"weburl"`
	Results []struct {
		Media []map[string]struct {
			URL string `json:"url"`
			Dimension [2]int `json:"dims"`
			Duration float64 `json:"duration"`
			Size int `json:"size"`

			Preview string `json:"preview"`
		} `json:"media"`
	} `json:"results"`
}
