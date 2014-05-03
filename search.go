package playstore

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"math"
	"net/url"
	"strconv"
)

const MAX_RESULTS_PER_REQUEST = 20

var (
	ErrInvalidLimit = errors.New("invalid limit.")
)

// Searchs by term. The request is sliced into limited requests by
// MAX_RESULTS_PER_REQUEST. It returns a list with all apps found giving only
// a
func Search(httpGet httpGetFunc, term string, limit int, lang string) ([]*AppSlug, error) {
	results := []*AppSlug{}
	if limit <= 0 {
		return results, ErrInvalidLimit
	}
	requests := int(math.Ceil(float64(limit) / MAX_RESULTS_PER_REQUEST))
	fmt.Println("Peticiones: %d", requests)
	urls := make([]*url.URL, requests)
	for i := 0; i < requests; i++ {
		urls[i] = getSearchUrl(term, strconv.Itoa(i*MAX_RESULTS_PER_REQUEST), MAX_RESULTS_PER_REQUEST, lang)
	}
	for _, url := range urls {
		response, err := httpGet(url)
		if err != nil {
			return results, err
		}
		document, err := NewPlayStoreDocument(response)
		if err != nil {
			return results, err
		}
		res, err := parseAppList(document)
		if err != nil {
			return results, err
		}
		results = append(results, res...)
	}
	return results, nil
}

// Contructs the url to make search requests. Parameters on the url are:
// q = search term
// start = a multiple of MAX_RESULTS_PER_REQUEST that tells whether to start.
// c = filters the Google Play Store results to only show apps.
// lang = language
// amount = the limit of the results. For some reason this value is doubled.
func getSearchUrl(term string, start string, amount int, lang string) *url.URL {
	query := url.Values{}
	query.Set("q", term)
	query.Set("start", start)
	query.Set("num", strconv.Itoa(amount>>1))
	query.Set("c", "apps")
	query.Set("hl", lang)
	return &url.URL{
		Scheme:   "https",
		Host:     ENDPOINT,
		Path:     "/search",
		RawQuery: query.Encode(),
	}
}

// Iterates over a list of cards retrieving apps information.
func parseAppList(document *playStoreDocument) ([]*AppSlug, error) {
	apps := make([]*AppSlug, MAX_RESULTS_PER_REQUEST)
	var err error
	document.Find(`.card`).Each(func(i int, sel *goquery.Selection) {
		fmt.Println("La i vale %d", i)
		if apps[i], err = parseAppSlug(sel); err != nil {
			fmt.Println("%s", err.Error())
			return
		}
	})
	return apps, nil
}