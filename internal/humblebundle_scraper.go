package internal

import (
	"bytes"
	"fmt"

	"github.com/PuerkitoBio/goquery"
)

type Bundle struct {
	Name  string
	Items []string
}

func GetBundleData(browserlessToken string, url string) (
	Bundle,
	error,
) {
	htmlContent, err := GrabContent(browserlessToken, url)
	if err != nil {
		return Bundle{}, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(htmlContent))
	if err != nil {
		return Bundle{}, err
	}

	bundleName, _ := doc.Find(".bundle-logo").First().Attr("alt")

	var itemNames []string

	items := doc.Find(".item-title")
	for _, bundleItem := range items.Nodes {
		itemNames = append(
			itemNames,
			fmt.Sprintf("- %s", bundleItem.FirstChild.Data),
		)
	}

	return Bundle{
		Name:  bundleName,
		Items: itemNames,
	}, nil
}
