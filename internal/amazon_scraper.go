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
	Language    string
	Publisher   string
	PublishedAt time.Time
	Description string
	Subjects    []string
	Asin        string
	Isbn        string
}

const AmazonUrl = "https://www.amazon.com.br/dp/%s"

func (b BookInfo) String() string {
	bookInfo, _ := json.Marshal(b)
	return string(bookInfo)
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

func getBookCode(asin, isbn string) (string, error) {
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

func languageToIso(languageText string) language.Tag {
	var languages = map[string]language.Tag{
		"Português": language.BrazilianPortuguese,
		"Inglês":    language.AmericanEnglish,
		"Espanhol":  language.Spanish,
		"Francês":   language.French,
		"Alemão":    language.German,
		"Italiano":  language.Italian,
	}

	return languages[languageText]
}

func GetBookInfo(browserlessToken, asin, isbn string) (BookInfo, error) {
	bookCode, err := getBookCode(asin, isbn)
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

	title := strings.Trim(doc.Find("#productTitle").Text(), " ")

	var authors []string
	doc.Find(".author > a").Each(
		func(_ int, s *goquery.Selection) {
			authors = append(authors, s.Text())
		},
	)

	var description strings.Builder
	doc.Find("#bookDescription_feature_div span:not(.a-expander-prompt)").Each(
		func(_ int, s *goquery.Selection) {
			description.WriteString(strings.Trim(s.Text(), " "))
		},
	)

	publisher := strings.Trim(
		doc.Find("#rpi-attribute-book_details-publisher div.rpi-attribute-value").Text(),
		" ",
	)

	publishedAtText := strings.Trim(
		doc.Find("#rpi-attribute-book_details-publication_date div.rpi-attribute-value").Text(),
		" ",
	)
	publishedAt, _ := parseDate(publishedAtText)

	languageText := strings.Trim(
		doc.Find("#rpi-attribute-language div.rpi-attribute-value").Text(),
		" ",
	)
	languageTag := languageToIso(languageText)

	imageUnparsedUrls, _ := doc.Find("#ebooksImgBlkFront").Attr("data-a-dynamic-image")
	var imageMapping map[string][]int

	err = json.Unmarshal([]byte(imageUnparsedUrls), &imageMapping)
	if err != nil {
		return BookInfo{}, err
	}

	var biggestCoverUrl string
	var biggestCoverHeight int

	for coverUrl, imageSize := range imageMapping {
		if imageSize[1] > biggestCoverHeight {
			biggestCoverHeight = imageSize[1]
			biggestCoverUrl = coverUrl
		}
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
		CoverUrl:    biggestCoverUrl,
		Language:    languageTag.String(),
		Publisher:   publisher,
		PublishedAt: publishedAt,
		Description: description.String(),
		Asin:        mobiAsin,
		Isbn:        isbn13,
	}, nil
}
