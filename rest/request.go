package rest

import (
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/grupawp/tensorflow-deploy/app"
	"github.com/grupawp/tensorflow-deploy/exterr"

	"github.com/go-chi/chi"
	"gopkg.in/go-playground/validator.v9"
)

var (
	logInvalidURLFieldErrorCode = 1003

	errorInvalidURLField = exterr.NewErrorWithMessage("couldn't parse an url, invalid parameters").WithComponent(app.ComponentRest).WithCode(logInvalidURLFieldErrorCode)
)

// URLParams is dedicated struct to hold parameters given in URL
type URLParams struct {
	Team            string `validate:"required,max=32,min=1"`
	Project         string `validate:"required,max=32,min=1"`
	Name            string `validate:"omitempty,max=32,min=1"`
	Version         int64  `validate:"omitempty,gte=1,lte=999"`
	Label           string `validate:"omitempty,max=32,min=1"`
	Status          string `validate:"omitempty,max=32,min=1"`
	SkipShortConfig bool   `validate:"omitempty"`
}

// ServableID returns a ServableID struct based on
// current URLParams
func (p *URLParams) ServableID() app.ServableID {
	return app.ServableID{Team: p.Team, Project: p.Project, Name: p.Name}
}

// QueryParameters returns a QueryParameters struct based on
// current URLParams
func (p *URLParams) QueryParameters() app.QueryParameters {
	params := app.QueryParameters{}

	s := reflect.ValueOf(p).Elem()

	for i := 0; i < s.NumField(); i++ {
		fieldName := s.Type().Field(i).Name
		field := s.FieldByName(fieldName)

		switch field.Kind() {
		case reflect.String:
			if len(field.String()) > 0 {
				params[strings.ToLower(fieldName)] = field.String()
			}

		case reflect.Int64:
			if field.Int() != 0 {
				params[strings.ToLower(fieldName)] = field.Int()
			}
		}
	}

	return params
}

const (
	urlTeam            = "Team"
	urlProject         = "Project"
	urlName            = "Name"
	urlVersion         = "Version"
	urlLabel           = "Label"
	urlSkipShortConfig = "SkipShortConfig"
)

func parseAndValidateParamsFromRequest(r *http.Request, allowQueryStrings bool, fields ...string) (*URLParams, error) {
	var urlParams URLParams
	var ommitedFields []string

	if allowQueryStrings {
		fields = nil
	}

	s := reflect.ValueOf(&urlParams).Elem()

	for i := 0; i < s.NumField(); i++ {
		ommitedFields = append(ommitedFields, s.Type().Field(i).Name)
		if allowQueryStrings {
			fields = append(fields, s.Type().Field(i).Name)
		}
	}

	for _, field := range fields {
		urlParam := getParamFromRequest(r, strings.ToLower(field))
		if len(urlParam) == 0 && allowQueryStrings {
			continue
		}

		for i, f := range ommitedFields {
			if strings.Compare(f, field) == 0 {
				ommitedFields = append(ommitedFields[:i], ommitedFields[i+1:]...)
			}
		}

		if s.FieldByName(field) == (reflect.Value{}) {
			return nil, errorInvalidURLField
		}

		switch s.FieldByName(field).Kind() {
		case reflect.String:
			s.FieldByName(field).SetString(strings.ToLower(urlParam))

		case reflect.Int64:
			parsed, _ := strconv.ParseInt(urlParam, 10, 64)
			s.FieldByName(field).SetInt(parsed)

		case reflect.Bool:
			parsed, _ := strconv.ParseBool(urlParam)
			s.FieldByName(field).SetBool(parsed)
		}
	}

	if err := validator.New().StructExcept(urlParams, ommitedFields...); err != nil {
		return nil, exterr.WrapWithFrame(err)
	}

	return &urlParams, nil
}

func getParamFromRequest(r *http.Request, name string) string {
	param := chi.URLParam(r, name)
	if len(param) > 0 {
		return param
	}

	return r.URL.Query().Get(name)
}
