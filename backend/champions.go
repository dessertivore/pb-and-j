package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/PuerkitoBio/goquery"
)

const championsURL = "https://www.leagueoflegends.com/en-gb/champions/"

type Champion struct {
	Name string `gorm:"primaryKey"` // display name, e.g. "Wukong"
	Slug string // URL slug, e.g. "monkeyking"
	URL  string
}

var Champions = make(map[string]Champion)

// Matches /en-gb/champions/<slug>/ but not the index page itself.
var slugRe = regexp.MustCompile(`/champions/([a-z0-9]+)/?$`)

func getChamps(ctx context.Context, db *gorm.DB, forceRefreshCache bool) ([]string, error) {
	if len(Champions) != 0 && forceRefreshCache == false {
		var output = make([]string, 0, len(Champions))
		for _, champion := range Champions {
			output = append(output, champion.Name)
		}
		return output, nil
	}
	if forceRefreshCache == true {
		fmt.Println("Forcing cache refresh for champions")
	}
	var champions []Champion
	dbResult := db.Find(&champions)
	if dbResult.Error == nil && dbResult.RowsAffected != 0 {
		var output = make([]string, 0, len(Champions))
		for _, champion := range champions {
			output = append(output, champion.Name)
			Champions[strings.ToLower(champion.Name)] = champion
		}
		return output, nil
	}
	fmt.Println("Failed to load champs from db, scraping them again.", dbResult.Error, "Scraping again.")

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
		result := db.Create(&championOrm)
		if result.Error != nil {
			fmt.Println("Failed to add", name, "to database", result.Error)
		}
		Champions[strings.ToLower(name)] = championOrm
	})
	return champs, nil
}

func ValidateChampsAgainstCache(champs *[]string) []Champion {
	validatedChamps := make([]Champion, 0, 5)

	for _, champ := range *champs {
		champ = strings.ToLower(champ)
		if validChamp, ok := Champions[champ]; ok {
			validatedChamps = append(validatedChamps, validChamp)
		} else {
			fmt.Println("Champion", champ, "does not exist")
		}
	}
	return validatedChamps
}
