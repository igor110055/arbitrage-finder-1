package filters

import "fmt"

const (
	Asc  SortDirection = "asc"
	Desc SortDirection = "desc"
)

// SortDirection is enum of SQL "order by" directions: asc and desc
type SortDirection string

// NewSortDirection transforms string => SortDirection with enum value check
func NewSortDirection(str string) (SortDirection, error) {
	sd := SortDirection(str)
	if sd != Asc && sd != Desc {
		return "", fmt.Errorf("invalid SortDirection enum %q: must be any of [%q, %q]", str, Asc, Desc)
	}
	return sd, nil
}
