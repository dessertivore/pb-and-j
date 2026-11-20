package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/PuerkitoBio/goquery"
)

const championsURL = "https://www.leagueoflegends.com/en-gb/champions/"

type Champion struct {
	gorm.Model
	Name string // display name, e.g. "Wukong"
	Slug string // URL slug, e.g. "monkeyking"
	URL  string
}

var Champions = make(map[string]Champion)

// Matches /en-gb/champions/<slug>/ but not the index page itself.
var slugRe = regexp.MustCompile(`/champions/([a-z0-9]+)/?$`)

func getChamps(ctx context.Context, db *gorm.DB) ([]string, error) {
	if len(Champions) != 0 {
		var output = make([]string, 0, len(Champions))
		for _, champion := range Champions {
			output = append(output, champion.Name)
		}
		return output, nil
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, championsURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", championsURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching %s: status %d", championsURL, resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parsing html: %w", err)
	}

	base, _ := url.Parse(championsURL)
	seen := make(map[string]bool)
	var champs []string

	doc.Find(`a[href*="/champions/"]`).Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr("href")

		m := slugRe.FindStringSubmatch(href)
		if m == nil {
			return
		}
		slug := m[1]
		if seen[slug] {
			return
		}

		// Collapse whitespace/newlines inside the anchor's text.
		name := strings.Join(strings.Fields(s.Text()), " ")
		if name == "" {
			return
		}

		// Resolve relative hrefs against the page URL.
		full := href
		if ref, err := url.Parse(href); err == nil {
			full = base.ResolveReference(ref).String()
		}
		seen[slug] = true
		champs = append(champs, name)
		championOrm := Champion{Name: name, Slug: slug, URL: full}
		err := gorm.G[Champion](db).Create(ctx, &championOrm)
		if err != nil {
			fmt.Println("Failed to add", name, "to database")
		}
		Champions[strings.ToLower(name)] = championOrm
	})
	return champs, nil
}

func scrapeWinRates(winRateURL string) (float64, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	req, err := http.NewRequest("GET", winRateURL, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	// Parse the HTML response.
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return 0, err
	}

	// Find the Win rate element and extract its percentage.
	var winRate float64
	var found bool

	doc.Find("em").EachWithBreak(func(i int, s *goquery.Selection) bool {

		if strings.TrimSpace(s.Text()) == "Win rate" {

			value := strings.TrimSuffix(
				strings.TrimSpace(s.Parent().Find("b").First().Text()),
				"%",
			)

			winRate, err = strconv.ParseFloat(value, 64)
			found = err == nil

			return false
		}

		return true
	})

	if !found {
		if err != nil {
			return 0, err
		}
		return 0, fmt.Errorf("win rate not found")
	}

	return winRate, nil
}
