package config

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/3WDeveloper-GM/json-endpoints/internal/validator"
)

func (app *Application) ReadStrings(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	return s
}

func (app *Application) ReadCSV(qs url.Values, key string, defaulValue []string) []string {
	csv := qs.Get(key)

	if csv == "" {
		return defaulValue
	}

	return strings.Split(csv, ",")
}

func (app *Application) ReadInt(qs url.Values, key string, defaultvalue int, v *validator.Validator) int {

	s := qs.Get(key)

	if s == "" {
		return defaultvalue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultvalue
	}

	return i

}
