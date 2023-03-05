package epub

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/text/language"
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

func NewOPF(
	epubVersion string,
	lang language.Tag,
	mainTitle string,
	mainAuthor string,
) OPF {
	uniqueIdentifierID := "pub_id"

	opf := OPF{
		Version:       epubVersion,
		XMLNamespace:  "http://www.idpf.org/2007/opf",
		XMLLanguage:   lang.String(),
		UniqueID:      uniqueIdentifierID,
		TextDirection: "ltr",
		Metadata: MetadataElement{
			XMLNamespace: "http://purl.org/dc/elements/1.1/",

			Identifier: []IdentifierElement{
				{
					Value: fmt.Sprintf("urn:uuid:%s", uuid.New().String()),
					ID:    uniqueIdentifierID,
				},
			},
			Language: lang.String(),
			Metas: []MetaElement{
				{
					Refines:  fmt.Sprintf("#%s", uniqueIdentifierID),
					Property: "identifier-type",
					BaseElement: BaseElement{
						Value: "uuid",
					},
				},
			},
		},
	}

	opf.AddTitle(mainTitle, Main)
	opf.AddContributor(mainAuthor, Author)

	if epubVersion == "3.0" {
		opf.SetModificationDate(time.Now())
	}

	return opf
}

func (opf *OPF) AddTitle(title string, titleType TitleType) {
	titleCount := len(opf.Metadata.Titles)
	titleId := fmt.Sprintf("title%02d", titleCount+1)

	opf.Metadata.Titles = append(
		opf.Metadata.Titles, BaseElement{
			Value: title,
			ID:    titleId,
		},
	)

	opf.Metadata.Metas = append(
		opf.Metadata.Metas, MetaElement{
			BaseElement: BaseElement{
				Value: titleType.String(),
			},
			Refines:  fmt.Sprintf("#%s", titleId),
			Property: "title-type",
		},
	)
}

func (opf *OPF) SetModificationDate(date time.Time) {
	opf.Metadata.Metas = append(
		opf.Metadata.Metas, MetaElement{
			BaseElement: BaseElement{
				Value: date.Format(time.RFC3339),
			},
			Property: "dcterms:modified",
		},
	)
}

func (opf *OPF) AddContributor(
	contributorName string,
	role ContributorRole,
) *ContributorElement {
	contributorRole := role.String()
	var id string
	var sliceToModify *[]ContributorElement

	if role == Author {
		id = fmt.Sprintf("creator%02d", len(opf.Metadata.Creators)+1)
		sliceToModify = &opf.Metadata.Creators
	} else {
		id = fmt.Sprintf("contributor%02d", len(opf.Metadata.Contributors)+1)
		sliceToModify = &opf.Metadata.Contributors
	}

	contributor := ContributorElement{
		Value: contributorName,
		ID:    id,
	}

	*sliceToModify = append(
		*sliceToModify, contributor,
	)

	opf.Metadata.Metas = append(
		opf.Metadata.Metas, MetaElement{
			BaseElement: BaseElement{
				Value: contributorRole,
			},
			Refines:  fmt.Sprintf("#%s", id),
			Property: "role",
			Scheme:   "marc:relators",
		},
	)

	return &contributor
}

func (opf *OPF) AddAlternateNameToContributor(
	contributor *ContributorElement,
	script string,
	lang language.Tag,
) {
	opf.Metadata.Metas = append(
		opf.Metadata.Metas, MetaElement{
			BaseElement: BaseElement{
				Value:       script,
				XMLLanguage: lang.String(),
			},
			Refines:  fmt.Sprintf("#%s", contributor.ID),
			Property: "alternate-script",
		},
	)
}

func (opf *OPF) AddSortNameToContributor(
	contributor *ContributorElement,
	sortName string,
) {
	opf.Metadata.Metas = append(
		opf.Metadata.Metas, MetaElement{
			BaseElement: BaseElement{
				Value: sortName,
			},
			Property: "file-as",
			Refines:  fmt.Sprintf("#%s", contributor.ID),
			Scheme:   "",
		},
	)
}

func (opf *OPF) SetDescription(description string) {
	opf.Metadata.Description = BaseElement{
		Value: description,
	}
}

func (opf *OPF) SetPublisher(publisher string) {
	opf.Metadata.Publisher = BaseElement{
		Value: publisher,
	}
}

func (opf *OPF) SetPublicationDate(date time.Time) {
	opf.Metadata.Date = date.Format(time.RFC3339)
}

func (opf *OPF) UpdateBookInfo(bookInfo *BookInfo) {
	opf.SetDescription(bookInfo.Description)
	opf.SetPublisher(bookInfo.Publisher)
	opf.SetPublicationDate(bookInfo.PublicationDate)
}

type OPF struct {
	XMLName      xml.Name `xml:"package"`
	UniqueID     string   `xml:"unique-identifier,attr"`
	Version      string   `xml:"version,attr"`
	XMLNamespace string   `xml:"xmlns,attr"`

	XMLLanguage   string `xml:"xml:lang,attr,omitempty"`
	ID            string `xml:"id,attr,omitempty"`
	Prefix        string `xml:"prefix,attr,omitempty"`
	TextDirection string `xml:"dir,attr,omitempty"`

	Metadata MetadataElement `xml:"metadata"`
}

type BaseElement struct {
	Value         string `xml:",chardata"`
	XMLLanguage   string `xml:"xml:lang,attr,omitempty"`
	ID            string `xml:"id,attr,omitempty"`
	TextDirection string `xml:"dir,attr,omitempty"`
}

type MetadataElement struct {
	XMLNamespace string `xml:"xmlns:dc,attr"`

	Identifier []IdentifierElement `xml:"dc:identifier"`
	Titles     []BaseElement       `xml:"dc:title"`
	Language   string              `xml:"dc:language"`

	Creators     []ContributorElement `xml:"dc:creator,omitempty"`
	Contributors []ContributorElement `xml:"dc:contributor,omitempty"`
	Date         string               `xml:"dc:date,omitempty"`
	Description  BaseElement          `xml:"dc:description,omitempty"`
	Publisher    BaseElement          `xml:"dc:publisher,omitempty"`

	Metas []MetaElement `xml:"meta,omitempty"`
}

type MetaElement struct {
	BaseElement
	Property string `xml:"property,attr"`

	Refines string `xml:"refines,attr,omitempty"`
	Scheme  string `xml:"scheme,attr,omitempty"`
}

type IdentifierElement struct {
	Value string `xml:",chardata"`
	ID    string `xml:"id,attr"`
}
