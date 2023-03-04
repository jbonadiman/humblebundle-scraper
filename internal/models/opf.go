package models

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TitleType int

const (
	Main TitleType = iota
	Subtitle
	Short
	Collection
	Edition
	Expanded
)

func (s TitleType) String() string {
	return [...]string{
		"main",
		"subtitle",
		"short",
		"collection",
		"edition",
		"expanded",
	}[s]
}

func NewOPF(epubVersion, language, mainTitle string) OPF {
	opf := OPF{
		Version:       epubVersion,
		XMLNamespace:  "http://www.idpf.org/2007/opf",
		XMLLanguage:   language,
		UniqueID:      "uuid_id",
		TextDirection: "ltr",
		Metadata: MetadataElement{
			XMLNamespace: "http://purl.org/dc/elements/1.1/",

			Identifier: []IdentifierElement{
				{
					Value:  uuid.New().String(),
					ID:     "uuid_id",
					Scheme: "uuid",
				},
			},
		},
	}

	opf.AddTitle(mainTitle, Main)

	if epubVersion == "3.0" {
		opf.SetModificationDate(time.Now())
	}

	return opf
}

func (opf *OPF) AddTitle(title string, titleType TitleType) {
	titleCount := len(opf.Metadata.Titles)
	titleId := fmt.Sprintf("title%d", titleCount+1)

	opf.Metadata.Titles = append(
		opf.Metadata.Titles, TitleElement{
			Value: title,
			ID:    titleId,
		},
	)

	opf.Metadata.Metas = append(
		opf.Metadata.Metas, MetaElement{
			Value:    titleType.String(),
			Refines:  fmt.Sprintf("#%s", titleId),
			Property: "title-type",
		},
	)
}

func (opf *OPF) SetModificationDate(date time.Time) {
	opf.Metadata.Metas = append(
		opf.Metadata.Metas, MetaElement{
			Value:    date.UTC().Format(time.RFC3339),
			Property: "dcterms:modified",
		},
	)
}

type OPF struct {
	XMLName      xml.Name `xml:"package"`
	Version      string   `xml:"version,attr"`
	XMLNamespace string   `xml:"xmlns,attr"`
	UniqueID     string   `xml:"unique-identifier,attr"`

	Metadata MetadataElement `xml:"metadata"`

	XMLLanguage   string `xml:"xml:lang,attr,omitempty"`
	TextDirection string `xml:"dir,attr,omitempty"`

	// Manifest ManifestElement `xml:"manifest"`
	// Spine    SpineElement    `xml:"spine"`
}

type MetadataElement struct {
	XMLNamespace string `xml:"xmlns:dc,attr"`

	Identifier []IdentifierElement `xml:"dc:identifier"`
	Titles     []TitleElement      `xml:"dc:title"`
	Language   string              `xml:"dc:language"`

	Creator string `xml:"dc:creator,omitempty"`
	Date    string `xml:"dc:date,omitempty"`
	Source  string `xml:"dc:source,omitempty"`

	Metas []MetaElement `xml:"meta,omitempty"`
}

type MetaElement struct {
	Value string `xml:",chardata"`

	Refines  string `xml:"refines,attr,omitempty"`
	Property string `xml:"property,attr"`
	ID       string `xml:"id,attr,omitempty"`
}

type TitleElement struct {
	Value         string `xml:",chardata"`
	ID            string `xml:"id,attr"`
	Language      string `xml:"xml:lang,attr,omitempty"`
	TextDirection string `xml:"dir,attr,omitempty"`
}

type IdentifierElement struct {
	Value  string `xml:",chardata"`
	ID     string `xml:"id,attr"`
	Scheme string `xml:"scheme,attr"`
}

type ManifestElement struct {
	Items []ItemElement `xml:"item"`
}

type SpineElement struct {
	ItemRefs []ItemRefElement `xml:"itemref"`
}

type ItemElement struct {
	ID        string `xml:"id,attr"`
	Href      string `xml:"href,attr"`
	MediaType string `xml:"media-type,attr"`
}

type ItemRefElement struct {
	IDRef string `xml:"idref,attr"`
}
