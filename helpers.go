package earlygrave

import (
	"fmt"
	"net/http"
)

type ECtxNotExist error

func GetPaginationContext(r *http.Request) (Pagination, error) {
	v, ok := r.Context().Value(ctxKeyPagination).(Pagination)
	if !ok {
		return Pagination{}, ECtxNotExist(fmt.Errorf("No pagination was found"))
	}
	return v, nil
}

func GetSortContext(r *http.Request) (Sort, error) {
	v, ok := r.Context().Value(ctxKeySort).(Sort)
	if !ok {
		return Sort{}, ECtxNotExist(fmt.Errorf("No sort was found"))
	}
	return v, nil
}
