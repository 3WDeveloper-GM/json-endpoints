package data

import (
	"fmt"
	"math"
	"strings"

	"github.com/3WDeveloper-GM/json-endpoints/internal/validator"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafeList []string
}

type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

func CalculateMetadata(totalrecords, page, pagesize int) Metadata {
	if totalrecords == 0 {
		return Metadata{}
	}
	return Metadata{
		CurrentPage:  page,
		PageSize:     pagesize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalrecords) / float64(pagesize))),
		TotalRecords: totalrecords,
	}
}

func ValidateFilters(v *validator.Validator, f Filters) {
	minimumPageQuantity, maximumPageQUantity := 0, 10_000_000

	var moreThanMsg = "must be greater than"
	var lessthanMsg = "must be less than"
	v.Check(f.Page > minimumPageQuantity, "page", fmt.Sprintf(moreThanMsg+" %v", minimumPageQuantity))
	v.Check(f.Page < maximumPageQUantity, "page", fmt.Sprintf(lessthanMsg+" %v", maximumPageQUantity))

	minimumPageSize, maximumPageSize := 0, 100
	v.Check(f.PageSize > minimumPageSize, "page_size", fmt.Sprintf(moreThanMsg+" %v", minimumPageSize))
	v.Check(f.PageSize < maximumPageSize, "page_size", fmt.Sprintf(lessthanMsg+" %v", maximumPageSize))

	var sortMsg = "invalid sort value"
	v.Check(validator.In(f.Sort, f.SortSafeList...), "sort", sortMsg)
}

func (f Filters) SortColumn() string {
	for _, safevalue := range f.SortSafeList {
		if f.Sort == safevalue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	panic("unsafe sort parameter: " + f.Sort)
}

func (f Filters) SortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

func (f Filters) Limit() int {
	return f.PageSize
}

func (f Filters) Offset() int {
	return (f.Page - 1) * f.PageSize
}
