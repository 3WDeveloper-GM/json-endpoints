package data

import (
	"fmt"

	"github.com/3WDeveloper-GM/json-endpoints/internal/validator"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafeList []string
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
