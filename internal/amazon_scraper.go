package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/language"
)

type BookInfo struct {
	Title       string
	Authors     []string
	CoverUrl    string
	Language    language.Tag
	Publisher   string
	PublishedAt time.Time
	Description string
	Asin        string
	Isbn        string
}

const AmazonUrl = "https://www.amazon.com.br/dp/%s"

func (b BookInfo) String() string {
	bookInfo, _ := json.Marshal(b)
	return string(bookInfo)
}

func getTextElement(
	document *goquery.Document,
	selector, elementName string,
) (string, error) {
	rawText := document.Find(selector).Text()

	if rawText == "" {
		return "", errors.New(
			fmt.Sprintf(
				"could not find %s using selector %q",
				elementName,
				selector,
			),
		)
	}

	return strings.Trim(rawText, " "), nil
}

func getBookCover(document *goquery.Document) (string, error) {
	selector := "#ebooksImgBlkFront"
	imageUnparsedUrls, _ := document.Find(selector).Attr("data-a-dynamic-image")

	var imageMapping map[string][]int
	err := json.Unmarshal([]byte(imageUnparsedUrls), &imageMapping)
	if err != nil {
		return "", err
	}

	var biggestCoverUrl string
	var biggestCoverHeight int

	for coverUrl, imageSize := range imageMapping {
		if imageSize[1] > biggestCoverHeight {
			biggestCoverHeight = imageSize[1]
			biggestCoverUrl = coverUrl
		}
	}

	return biggestCoverUrl, nil
}

func getBookPublishedDate(document *goquery.Document) (time.Time, error) {
	publishedAtText, err := getTextElement(
		document,
		"#rpi-attribute-book_details-publication_date div.rpi-attribute-value",
		"publishedAt",
	)
	if err != nil {
		return time.Time{}, err
	}

	return parseDate(publishedAtText)
}

func getBookTitle(document *goquery.Document) (string, error) {
	return getTextElement(document, "#productTitle", "title")
}

func getBookPublisher(document *goquery.Document) (string, error) {
	return getTextElement(
		document,
		"#rpi-attribute-book_details-publisher div.rpi-attribute-value",
		"publisher",
	)
}

func getBookAuthors(document *goquery.Document) ([]string, error) {
	selector := ".author > a"

	var authors []string
	document.Find("").Each(
		func(_ int, s *goquery.Selection) {
			authors = append(authors, strings.Trim(s.Text(), " "))
		},
	)

	for _, name := range authors {
		if name == "" {
			return nil, errors.New(
				fmt.Sprintf(
					"could not find some/all authors using selector %q",
					selector,
				),
			)
		}
	}

	return authors, nil
}

func getBookDescription(document *goquery.Document) (string, error) {
	selector := "#bookDescription_feature_div span:not(.a-expander-prompt)"

	var descriptionBuilder strings.Builder
	document.Find(selector).Each(
		func(_ int, s *goquery.Selection) {
			descriptionBuilder.WriteString(strings.Trim(s.Text(), " "))
		},
	)

	description := descriptionBuilder.String()

	if description == "" {
		return "", errors.New(
			fmt.Sprintf(
				"could not find description using selector %q",
				selector,
			),
		)
	}

	return description, nil
}

func getBookLanguage(document *goquery.Document) (language.Tag, error) {
	languageText, err := getTextElement(
		document,
		"#rpi-attribute-language div.rpi-attribute-value",
		"language",
	)
	if err != nil {
		return language.Tag{}, err
	}

	return languageToIso(languageText)
}

func isbn10ToIsbn13(isbn10 string) string {
	isbn13 := "978" + isbn10[:9]

	checksum := 0
	for i, char := range isbn13 {
		if i%2 == 0 {
			checksum += 1 * int(char-'0')
		} else {
			checksum += 3 * int(char-'0')
		}
	}

	checksum = 10 - checksum%10
	if checksum == 10 {
		checksum = 0
	}
	isbn13 += strconv.Itoa(checksum)

	return isbn13
}

func parseBookCode(asin, isbn string) (string, error) {
	if asin == "" && isbn == "" {
		return "", errors.New("ASIN or ISBN-13 codes are mandatory")
	}

	if asin != "" {
		if !strings.HasPrefix(strings.ToLower(asin), "b") {
			return "", errors.New("invalid ASIN code")
		}

		return asin, nil
	}

	isbnPattern := regexp.MustCompile("^(?:\\d[\\d-]{0,4}){3}[\\dX]$")
	if !isbnPattern.MatchString(isbn) {
		return "", errors.New("invalid ISBN code")
	}

	if len(isbn) == 10 {
		isbn = isbn10ToIsbn13(isbn)
	}

	return isbn, nil
}

func parseDate(date string) (time.Time, error) {
	var months = map[string]string{
		"janeiro":   "January",
		"fevereiro": "February",
		"março":     "March",
		"abril":     "April",
		"maio":      "May",
		"junho":     "June",
		"julho":     "July",
		"agosto":    "August",
		"setembro":  "September",
		"outubro":   "October",
		"novembro":  "November",
		"dezembro":  "December",
	}

	for portugueseMonth, englishMonth := range months {
		date = strings.ReplaceAll(date, portugueseMonth, englishMonth)
	}

	return time.Parse("2 January 2006", date)
}

func languageToIso(languageText string) (language.Tag, error) {
	var languages = map[string]language.Tag{
		"Português": language.BrazilianPortuguese,
		"Inglês":    language.AmericanEnglish,
		"Espanhol":  language.Spanish,
		"Francês":   language.French,
		"Alemão":    language.German,
		"Italiano":  language.Italian,
	}

	if tag, ok := languages[languageText]; ok {
		return tag, nil

	}

	return language.Tag{}, errors.New(
		fmt.Sprintf(
			"could not parse language %q to its ISO equivalent",
			languageText,
		),
	)
}

func GetBookInfo(browserlessToken, asin, isbn string) (BookInfo, error) {
	bookCode, err := parseBookCode(asin, isbn)
	if err != nil {
		return BookInfo{}, err
	}

	htmlContent, err := GrabContent(
		browserlessToken,
		fmt.Sprintf(AmazonUrl, bookCode),
	)
	if err != nil {
		return BookInfo{}, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(htmlContent))
	if err != nil {
		return BookInfo{}, err
	}

	title, err := getBookTitle(doc)
	if err != nil {
		return BookInfo{}, err
	}

	authors, err := getBookAuthors(doc)
	if err != nil {
		return BookInfo{}, err
	}

	description, err := getBookDescription(doc)
	if err != nil {
		return BookInfo{}, err
	}

	publisher, err := getBookPublisher(doc)
	if err != nil {
		return BookInfo{}, err
	}

	publishedAt, err := getBookPublishedDate(doc)
	if err != nil {
		return BookInfo{}, err
	}

	languageTag, err := getBookLanguage(doc)
	if err != nil {
		return BookInfo{}, err
	}

	coverUrl, err := getBookCover(doc)
	if err != nil {
		return BookInfo{}, err
	}

	var mobiAsin string
	var isbn13 string

	if asin != "" {
		mobiAsin = bookCode
		// TODO: get isbn13
		isbn13 = ""
		// isbn13 = doc.Find("#rpi-attribute-book_details-isbn13 .rpi-attribute-value").Text()
	} else {
		isbn13 = bookCode
		// TODO: get asin
		mobiAsin = ""
	}

	return BookInfo{
		Title:       title,
		Authors:     authors,
		CoverUrl:    coverUrl,
		Language:    languageTag,
		Publisher:   publisher,
		PublishedAt: publishedAt,
		Description: description,
		Asin:        mobiAsin,
		Isbn:        isbn13,
	}, nil
}
