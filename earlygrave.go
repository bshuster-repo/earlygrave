package earlygrave

import (
	"net/http"
)

type Filter func(r *http.Request) (*http.Request, error)
type ConfigFilter func(funct Filter) Filter
type Validate func(*http.Request) error
type Extract func(*http.Request) (*http.Request, error)

func identity(r *http.Request) (*http.Request, error) {
	return r, nil
}

func New(cfgs ...ConfigFilter) Filter {
	res := Filter(identity)
	for _, decorator := range cfgs {
		res = decorator(res)
	}
	return res
}

func ValidateParam(validate Validate) ConfigFilter {
	return ConfigFilter(func(f Filter) Filter {
		return Filter(func(r *http.Request) (*http.Request, error) {
			if err := validate(r); err != nil {
				return r, err
			}
			return f(r)
		})
	})
}

func ExtractParam(extract Extract) ConfigFilter {
	return ConfigFilter(func(f Filter) Filter {
		return Filter(func(r *http.Request) (*http.Request, error) {
			nr, err := extract(r)
			if err != nil {
				return r, err
			}
			return f(nr)
		})
	})
}
