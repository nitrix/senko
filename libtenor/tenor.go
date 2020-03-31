package libtenor

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
)

type Tenor struct {
	Token string
	NSFW bool
}

func NewTenor(token string) Tenor {
	return Tenor {
		Token: token,
	}
}

func (t Tenor) RandomGif(search string) (Gif, error) {
	endpoint := url.URL{
		Scheme: "https",
		Host: "api.tenor.com",
		Path: "/v1/search",
	}

	query := url.Values{}
	query.Set("q", search)
	query.Set("key", t.Token)
	query.Set("media_filter", "basic")
	query.Set("limit", "50")

	if t.NSFW {
		query.Set("contentfilter", "off")
	}

	endpoint.RawQuery = query.Encode()

	randomGifResponse := randomGifResponse{}
	gif := Gif{}

	httpResponse, err := http.Get(endpoint.String())
	if err != nil {
		return gif, fmt.Errorf("error during tenor search: %w", err)
	}

	decoder := json.NewDecoder(httpResponse.Body)
	err = decoder.Decode(&randomGifResponse)
	if err != nil {
		return gif, fmt.Errorf("unable to json decode random gif response: %w", err)
	}

	rand.Shuffle(len(randomGifResponse.Results), func(i, j int) {
		randomGifResponse.Results[i], randomGifResponse.Results[j] = randomGifResponse.Results[j], randomGifResponse.Results[i]
	})

	for _, result := range randomGifResponse.Results {
		for _, media := range result.Media {
			if v, ok := media["gif"]; ok {
				gif.URL = v.URL
				gif.Width = v.Dimension[0]
				gif.Height = v.Dimension[1]
				gif.Duration = v.Duration
				gif.Preview = v.Preview

				return gif, nil
			}
		}
	}

	return gif, fmt.Errorf("no gif results")
}