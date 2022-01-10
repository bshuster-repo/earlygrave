package earlygrave

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func decorator1(f Filter) Filter {
	return Filter(func(r *http.Request) (*http.Request, error) {
		rr, err := http.NewRequest(http.MethodHead, r.URL.String(), r.Body)
		if err != nil {
			return rr, err
		}
		return f(rr)
	})
}

func decorator2(f Filter) Filter {
	return Filter(func(r *http.Request) (*http.Request, error) {
		rr, err := http.NewRequest(r.Method, fmt.Sprintf("%s/world", r.URL.String()), r.Body)
		if err != nil {
			return rr, err
		}
		return f(rr)
	})
}

func TestNew(t *testing.T) {
	funct := New(decorator2, decorator1)
	rr, err := http.NewRequest(http.MethodGet, "/hello", nil)
	if err != nil {
		t.Errorf("creating a new request failed: %s", err)
	}
	req, err2 := funct(rr)
	if err2 != nil {
		t.Errorf("funct returned error: %s", err)
	}

	expectedURL := "/hello/world"
	expectedMethod := http.MethodHead

	if req.URL.String() != expectedURL {
		t.Errorf("Expected request URL to be %s but got %s", expectedURL, req.URL.String())
	}
	if req.Method != expectedMethod {
		t.Errorf("Expected request method to be %s but got %s", expectedMethod, req.Method)
	}
}

func TestSortValidator(t *testing.T) {
	tt := []struct {
		Url         string
		ExpectedErr error
	}{
		{Url: "/", ExpectedErr: nil},
		{Url: "/?sort=name", ExpectedErr: nil},
		{Url: "/?sort=-name", ExpectedErr: nil},
		{Url: "/?sort=rank", ExpectedErr: fmt.Errorf("rank is not sortable")},
		{Url: "/?sort=-rank", ExpectedErr: fmt.Errorf("rank is not sortable")},
	}

	for _, te := range tt {
		funct := New(SortValidator([]string{"name", "role"}))
		rr, err := http.NewRequest(http.MethodGet, te.Url, nil)
		if err != nil {
			t.Errorf("Failed creating request: %q", err)
		}
		_, pErr := funct(rr)
		if te.ExpectedErr == nil && pErr != nil {
			t.Errorf("Expected no error but got %q", pErr)
		}
		if te.ExpectedErr != nil && pErr == nil {
			t.Errorf("Expected error %q but got nothing", te.ExpectedErr)
		}
		if te.ExpectedErr != nil && pErr != nil && te.ExpectedErr.Error() != pErr.Error() {
			t.Errorf("Expected error %q but got %q", te.ExpectedErr, pErr)
		}
	}
}

func TestPaginationValidator(t *testing.T) {
	tt := []struct {
		Url         string
		ExpectedErr error
	}{
		{Url: "/", ExpectedErr: nil},
		{Url: "/?limit=100", ExpectedErr: nil},
		{Url: "/?offset=3", ExpectedErr: nil},
		{Url: "/?offset=3&limit=34", ExpectedErr: nil},
		{Url: "/?limit=s200", ExpectedErr: fmt.Errorf("strconv.Atoi: parsing \"s200\": invalid syntax")},
		{Url: "/?offset=blabla", ExpectedErr: fmt.Errorf("strconv.Atoi: parsing \"blabla\": invalid syntax")},
	}

	for _, te := range tt {
		funct := New(PaginationValidator())
		rr, err := http.NewRequest(http.MethodGet, te.Url, nil)
		if err != nil {
			t.Errorf("Failed creating request: %q", err)
		}
		_, pErr := funct(rr)
		if te.ExpectedErr == nil && pErr != nil {
			t.Errorf("Expected no error but got %q", pErr)
		}
		if te.ExpectedErr != nil && pErr == nil {
			t.Errorf("Expected error %q but got nothing", te.ExpectedErr)
		}
		if te.ExpectedErr != nil && pErr != nil && te.ExpectedErr.Error() != pErr.Error() {
			t.Errorf("Expected error %q but got %q", te.ExpectedErr, pErr)
		}
	}
}

func TestPaginationExtractor(t *testing.T) {
	tt := []struct {
		Url    string
		Result Pagination
	}{
		{Url: "/", Result: Pagination{Offset: "0", Limit: "30"}},
		{Url: "/?offset=3&limit=34", Result: Pagination{Offset: "3", Limit: "34"}},
		{Url: "/?offset=34", Result: Pagination{Offset: "34", Limit: "30"}},
		{Url: "/?limit=20", Result: Pagination{Offset: "0", Limit: "20"}},
	}

	for _, te := range tt {
		funct := New(PaginationExtractor(Pagination{Limit: "30", Offset: "0"}))
		rr, err := http.NewRequest(http.MethodGet, te.Url, nil)
		if err != nil {
			t.Errorf("Failed creating request: %q", err)
		}
		res, pErr := funct(rr)
		if pErr != nil {
			t.Errorf("Expected no error but got %q", pErr)
		}
		v, eErr := GetPaginationContext(res)
		if eErr != nil {
			t.Errorf("Expected to have pagination data in context but found none")
		}
		if v.Limit != te.Result.Limit {
			t.Errorf("Expected limit to be %q but got %q", te.Result.Limit, v.Limit)
		}
		if v.Offset != te.Result.Offset {
			t.Errorf("Expected offset to be %q but got %q", te.Result.Offset, v.Offset)
		}
	}
}

func TestSortExtractor(t *testing.T) {
	tt := []struct {
		Url    string
		Result Sort
	}{
		{Url: "/", Result: Sort{Column: "name", Direction: "DESC"}},
		{Url: "/?sort=user", Result: Sort{Column: "user", Direction: "ASC"}},
		{Url: "/?sort=-user", Result: Sort{Column: "user", Direction: "DESC"}},
	}

	for _, te := range tt {
		funct := New(SortExtractor(Sort{Column: "name", Direction: "DESC"}))
		rr, err := http.NewRequest(http.MethodGet, te.Url, nil)
		if err != nil {
			t.Errorf("Failed creating request: %q", err)
		}
		res, pErr := funct(rr)
		if pErr != nil {
			t.Errorf("Expected no error but got %q", pErr)
		}

		v, eErr := GetSortContext(res)
		if eErr != nil {
			t.Errorf("Expected to have pagination data in context but found none")
		}
		if v.Direction != te.Result.Direction {
			t.Errorf("Expected direction to be %q but got %q", te.Result.Direction, v.Direction)
		}
		if v.Column != te.Result.Column {
			t.Errorf("Expected column to be %q but got %q", te.Result.Column, v.Column)
		}
	}
}

func TestGetPaginationContextFailure(t *testing.T) {
	rr, err := http.NewRequest(http.MethodGet, "/?sort=-name", nil)
	if err != nil {
		t.Errorf("Failed creating request: %q", err)
	}
	_, err = GetPaginationContext(rr)
	if err == nil {
		t.Errorf("Expected to have error but got none")
	}
	if _, ok := err.(ECtxNotExist); !ok {
		t.Errorf("Expected error to be %q but got %q", "No pagination was found", err)
	}
}

func TestGetSortContextFailure(t *testing.T) {
	rr, err := http.NewRequest(http.MethodGet, "/?limit=10", nil)
	if err != nil {
		t.Errorf("Failed creating request: %q", err)
	}
	_, err = GetSortContext(rr)
	if err == nil {
		t.Errorf("Expected to have error but got none")
	}
	if _, ok := err.(ECtxNotExist); !ok {
		t.Errorf("Expected error to be %q but got %q", "No sort was found", err)
	}
}

func extractor1(r *http.Request) (*http.Request, error) {
	return nil, fmt.Errorf("Oops!")
}

func TestExtractParamFailure(t *testing.T) {
	funct := New(ExtractParam(extractor1))
	rr, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Errorf("creating a new request failed: %s", err)
	}
	_, err = funct(rr)
	if err == nil {
		t.Errorf("Expected to have error but got none")
	}
}

func TestChoiceValidator(t *testing.T) {
	funct := New(ChoiceValidator("currency", map[string]bool{"USD": true, "NIS": true}))
	tcs := []struct {
		name        string
		expectedErr error
		url         string
	}{
		{
			name:        "invalid choice",
			expectedErr: errors.New("BLA is an invalid option for currency"),
			url:         "/?currency=BLA",
		},
	}

	for _, tc := range tcs {
		rr, err := http.NewRequest(http.MethodGet, tc.url, nil)
		if err != nil {
			t.Errorf("%s: creating a new request failed: %s", tc.name, err)
		}
		_, err = funct(rr)
		if tc.expectedErr != nil && err != nil {
			if err.Error() != tc.expectedErr.Error() {
				t.Errorf("%s: expected error to be %s but got %s", tc.name, tc.expectedErr, err)
			}
		}
		if err == nil && tc.expectedErr != nil {
			t.Errorf("%s: expected to have error but got none", tc.name)
		}
		if err != nil && tc.expectedErr == nil {
			t.Errorf("%s: expected to have no errors but got: %s", tc.name, err)
		}
	}
}
