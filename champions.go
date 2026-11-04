package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
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

var Champions = make(map[string]Champion)

// Matches /en-gb/champions/<slug>/ but not the index page itself.
var slugRe = regexp.MustCompile(`/champions/([a-z0-9]+)/?$`)

func getChamps(ctx context.Context) ([]string, error) {
	if len(Champions) != 0 {
		var output = make([]string, len(Champions))
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
		Champions[strings.ToLower(name)] = Champion{Name: name, Slug: slug, URL: full}
	})

	return champs, nil
}

//func getChampsString(ctx context.Context) ([]string, error) {
//	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
//	defer cancel()

//	req, err := http.NewRequestWithContext(ctx, http.MethodGet, championsURL, nil)
//	if err != nil {
//		return nil, err
//	}
//	resp, err := http.DefaultClient.Do(req)
//	if err != nil {
//		return nil, fmt.Errorf("fetching %s: %w", championsURL, err)
//	}
//	defer resp.Body.Close()
//
//	if resp.StatusCode != http.StatusOK {
//		return nil, fmt.Errorf("fetching %s: status %d", championsURL, resp.StatusCode)
//	}
//
//	doc, err := goquery.NewDocumentFromReader(resp.Body)
//	if err != nil {
//		return nil, fmt.Errorf("parsing html: %w", err)
//	}
//
//	seen := make(map[string]bool)
//	var champs []string
//
//	doc.Find(`a[href*="/champions/"]`).Each(func(_ int, s *goquery.Selection) {
//		name := s.Text()
//		if name == "" {
//			return
//		}
//		// Collapse whitespace/newlines inside the anchor's text.
//		if seen[name] {
//			return
//		}
//
//		seen[name] = true
//		champs = append(champs, name)
//	})
//
//	return champs, nil
//}

func scrapeWinRates(winRateURL string) float64 {

	client := &http.Client{Timeout: 15 * time.Second}

	req, _ := http.NewRequest("GET", winRateURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	html := string(body)

	// Find the Win rate element and its parent container.
	labelRegex := regexp.MustCompile(
		`(?s)<em[^>]*>\s*Win rate\s*</em>`,
	)

	loc := labelRegex.FindStringIndex(html)
	if loc == nil {
		fmt.Println("Win rate element not found")
		return 0
	}

	// Search the surrounding HTML for the percentage.
	start := loc[0] - 500
	if start < 0 {
		start = 0
	}

	end := loc[1] + 500
	if end > len(html) {
		end = len(html)
	}

	section := html[start:end]

	percentageRegex := regexp.MustCompile(`(\d{1,3}(?:\.\d+)?)%`)

	matches := percentageRegex.FindStringSubmatch(section)
	if len(matches) <= 1 {
		return 0
	}
	winRate := matches[1]
	fmt.Println("Win rate string:", winRate)
	output, err := strconv.ParseFloat(winRate, 64)
	if err != nil {

		fmt.Println("Win rate percentage could not be converted to float: ", err)
		return 0
	}
	return output
}
