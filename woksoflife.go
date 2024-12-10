package webscrapers

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Recipe struct {
	Name         string
	Image        string
	PrepTime     time.Duration
	Ingredients  []string
	Instructions []string
	Notes        string
}

func (r Recipe) String() string {
	recipe, _ := json.Marshal(r)
	return string(recipe)
}

func GetRecipe(browserlessToken, url string) (Recipe, error) {
	htmlContent, err := GrabContent(browserlessToken, url)
	if err != nil {
		return Recipe{}, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(htmlContent))
	if err != nil {
		return Recipe{}, err
	}

	name := doc.Find(".wprm-recipe-name").Text()

	image, _ := doc.Find(".attachment-post-thumbnail").Attr("data-lazy-src")

	var prepTime time.Duration

	hours, err := strconv.Atoi(doc.Find(".wprm-recipe-total_time-hours").Text())
	if err == nil {
		prepTime += time.Duration(hours) * time.Hour
	}

	minutes, err := strconv.Atoi(doc.Find(".wprm-recipe-total_time-minutes").Text())
	if err == nil {
		prepTime += time.Duration(minutes) * time.Minute
	}

	var ingredients []string
	doc.Find(".wprm-recipe-ingredient").Each(
		func(_ int, s *goquery.Selection) {
			ingredients = append(ingredients, strings.TrimLeft(s.Text(), "▢ "))
		},
	)

	var instructions []string
	doc.Find(".wprm-recipe-instruction").Each(
		func(_ int, s *goquery.Selection) {
			instructions = append(instructions, s.Text())
		},
	)

	notes := doc.Find(".wprm-recipe-notes").Text()

	return Recipe{
		Name:         name,
		Image:        image,
		Ingredients:  ingredients,
		Instructions: instructions,
		Notes:        notes,
		PrepTime:     prepTime,
	}, nil
}
