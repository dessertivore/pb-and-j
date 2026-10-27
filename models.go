package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const championsURL = "https://www.leagueoflegends.com/en-gb/champions/"

type Champion struct {
	Name string // display name, e.g. "Wukong"
	Slug string // URL slug, e.g. "monkeyking"
	URL  string
}

// Matches /en-gb/champions/<slug>/ but not the index page itself.
var slugRe = regexp.MustCompile(`/champions/([a-z0-9]+)/?$`)

func getChamps(ctx context.Context) ([]Champion, error) {
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
	var champs []Champion

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
		champs = append(champs, Champion{Name: name, Slug: slug, URL: full})
	})

	return champs, nil
}
