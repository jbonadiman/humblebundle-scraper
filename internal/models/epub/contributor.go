package epub

type ContributorRole int

const (
	Author ContributorRole = iota
	Translator
	Editor
	Illustrator
)

func (s ContributorRole) String() string {
	return [...]string{
		"aut",
		"trl",
		"edt",
		"ill",
	}[s]
}

type ContributorElement struct {
	Value string `xml:",chardata"`
	ID    string `xml:"id,attr,omitempty"`
}
