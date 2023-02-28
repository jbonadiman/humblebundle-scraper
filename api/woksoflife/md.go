package woksoflife

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"webscrapers/internal"
)

const ImageSize = 400

func normalizeFractions(line string) string {
	replacements := map[string]string{
		"1/2": "½",
		"1/3": "⅓",
		"1/4": "¼",
		"2/3": "⅔",
		"3/4": "¾",
		"2/5": "⅖",
		"3/5": "⅗",
		"4/5": "⅘",
	}

	r := regexp.MustCompile(`([123]/[234])`)
	return r.ReplaceAllStringFunc(
		line, func(match string) string {
			if fraction, ok := replacements[match]; ok {
				return fraction
			}

			return match
		},
	)
}

//goland:noinspection GoUnusedExportedFunction
func Handler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	if !queryParams.Has("url") || !queryParams.Has("browserlessToken") {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("the query params 'url' and 'browserlessToken' are required"))
		return
	}

	url := queryParams.Get("url")
	browserlessToken := queryParams.Get("browserlessToken")

	recipe, _ := internal.GetRecipe(browserlessToken, url)

	response := strings.Builder{}
	response.WriteString(
		fmt.Sprintf(
			"# %s\n\n> original recipe: %s\n\n<img src=\"%s\" width=\"%d\"/>\n\n> prep time: %v\n\n## ingredients\n",
			recipe.Name,
			url,
			recipe.Image,
			ImageSize,
			recipe.PrepTime,
		),
	)

	for _, ingredient := range recipe.Ingredients {
		response.WriteString(
			fmt.Sprintf(
				"* %s\n",
				normalizeFractions(ingredient),
			),
		)
	}

	response.WriteString("\n## instructions\n")
	for i, instruction := range recipe.Instructions {
		response.WriteString(
			fmt.Sprintf(
				"%d. %s\n",
				i+1,
				normalizeFractions(instruction),
			),
		)
	}

	if recipe.Notes != "" {
		response.WriteString("\n> notes: " + recipe.Notes)
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Add("Cache-Control", "max-age=0, s-maxage=86400")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(response.String()))
}
