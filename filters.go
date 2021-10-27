package earlygrave

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type ctxKey int

const (
	ctxKeyPagination = ctxKey(1)
	ctxKeySort       = ctxKey(2)
)

type Pagination struct {
	Limit  string
	Offset string
}

func PaginationValidator() ConfigFilter {
	validate := Validate(func(r *http.Request) error {
		values := r.URL.Query()
		limit := values.Get("limit")
		offset := values.Get("offset")

		if _, err := strconv.Atoi(limit); limit != "" && err != nil {
			return err
		}

		if _, err := strconv.Atoi(offset); offset != "" && err != nil {
			return err
		}

		return nil
	})
	return ValidateParam(validate)
}

func PaginationExtractor(defaultPagination Pagination) ConfigFilter {
	extract := Extract(func(r *http.Request) (*http.Request, error) {
		result := Pagination{Offset: defaultPagination.Offset, Limit: defaultPagination.Limit}
		values := r.URL.Query()
		if limit := values.Get("limit"); limit != "" {
			result.Limit = limit
		}
		if offset := values.Get("offset"); offset != "" {
			result.Offset = offset
		}
		return r.WithContext(context.WithValue(r.Context(), ctxKeyPagination, result)), nil
	})
	return ExtractParam(extract)
}

type Sort struct {
	Column    string
	Direction string
}

func SortValidator(sortedColumns []string) ConfigFilter {
	validate := Validate(func(r *http.Request) error {
		values := r.URL.Query()
		column := values.Get("sort")
		if column == "" {
			return nil
		}
		if strings.HasPrefix(column, "-") {
			column = column[1:]
		}
		found := false
		for _, sortedCol := range sortedColumns {
			if sortedCol == column {
				found = true
				break
			}
		}
		if found == false {
			return fmt.Errorf("%s is not sortable", column)
		}
		return nil
	})
	return ValidateParam(validate)
}

func SortExtractor(defaultSort Sort) ConfigFilter {
	extract := Extract(func(r *http.Request) (*http.Request, error) {
		result := Sort{Column: defaultSort.Column, Direction: defaultSort.Direction}
		values := r.URL.Query()
		if sort := values.Get("sort"); sort != "" {
			if strings.HasPrefix(sort, "-") {
				result.Column = sort[1:]
				result.Direction = "DESC"
			} else {
				result.Column = sort
				result.Direction = "ASC"
			}
		}
		return r.WithContext(context.WithValue(r.Context(), ctxKeySort, result)), nil
	})
	return ExtractParam(extract)
}
