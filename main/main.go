package main

import (
	"encoding/xml"
	"log"
	"time"

	"golang.org/x/text/language"
	"webscrapers/internal"
	"webscrapers/internal/models/epub"
)

func main() {
	// bookInfo, _ := internal.GetBookInfo(
	// 	os.Getenv("BROWSERLESS_TOKEN"),
	// 	"B083G6VYBZ",
	// 	"",
	// )
	// log.Println(bookInfo.String())

	testData := internal.BookInfo{
		Title:       "O cheiro do ralo",
		Authors:     []string{"Lourenço Mutarelli"},
		CoverUrl:    "https://images-na.ssl-images-amazon.com/images/I/51ZQYQZQJNL._SX331_BO1,204,203,200_.jpg",
		Language:    language.BrazilianPortuguese,
		Publisher:   "Companhia das Letras",
		PublishedAt: time.Date(2019, 11, 1, 0, 0, 0, 0, time.UTC),
		Description: "Descrição do livro",
		Asin:        "B083G6VYBZ",
		Isbn:        "9788535931008",
	}

	contentOpf := epub.NewOPF(
		"3.0",
		testData.Language,
		testData.Title,
		testData.Authors[0],
	)

	contentOpf.SetDescription(testData.Description)
	contentOpf.SetPublisher(testData.Publisher)
	contentOpf.SetPublicationDate(testData.PublishedAt)

	contentOpf.AddSortNameToContributor(
		&contentOpf.Metadata.Creators[0],
		"Muratelli, Lourenço",
	)

	opfAsXml, err := xml.MarshalIndent(contentOpf, "", "  ")
	if err != nil {
		panic(err)
	}

	log.Println(string(opfAsXml))
}
