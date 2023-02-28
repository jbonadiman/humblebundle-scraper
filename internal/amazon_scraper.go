package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
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

func validateBookCode(asin, isbn string) (string, error) {
	if asin == "" && isbn == "" {
		return "", errors.New("ASIN or ISBN-13 codes are mandatory")
	}

	if asin != "" {
		if !strings.HasPrefix(strings.ToLower(asin), "b") {
			return "", errors.New("invalid ASIN code")
		}

		return asin, nil
	}

	isbnPattern := regexp.MustCompile("^(?=(?:\\D*\\d){10}(?:(?:\\D*\\d){3})?$)[\\d-]+$")
	if !isbnPattern.MatchString(isbn) {
		return "", errors.New("invalid ISBN code")
	}

	// converts ISBN-10 to ISBN-13
	if len(isbn) == 10 {
		isbn = isbn10ToIsbn13(isbn)
	}

	return isbn, nil
}

func GetBookInfo(browserlessToken, asin, isbn string) (BookInfo, error) {
	bookCode, err := validateBookCode(asin, isbn)
	if err != nil {
		return BookInfo{}, err
	}

	log.Println(bookCode)
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
	doc.Find(".author").Each(
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

	return BookInfo{
		Title:       title,
		Authors:     authors,
		CoverUrl:    "",
		Language:    "",
		Publisher:   "",
		PublishedAt: time.Time{},
		Description: description.String(),
		Subjects:    nil,
		Asin:        "",
		Isbn:        "",
	}, nil
}
