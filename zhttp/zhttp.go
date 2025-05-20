package zhttp

import (
	"net/http"
	"net/url"
	"reflect"
	"strings"

	p "github.com/Oudwins/zog/internals"
	"github.com/Oudwins/zog/parsers/zjson"
	"github.com/Oudwins/zog/zconst"
)

type ParserFunc = func(r *http.Request) p.DpFactory

var (
	formTag      string = "form"
	queryParam   string = "query"
	multipartTag string = "multipart"
)

var Config = struct {
	Parsers struct {
		JSON      ParserFunc
		Form      ParserFunc
		Multipart ParserFunc
		Query     ParserFunc
	}
}{
	Parsers: struct {
		JSON      ParserFunc
		Form      ParserFunc
		Multipart ParserFunc
		Query     ParserFunc
	}{
		JSON: func(r *http.Request) p.DpFactory {
			return zjson.Decode(r.Body)
		},
		Form: func(r *http.Request) p.DpFactory {
			return func() (p.DataProvider, *p.ZogIssue) {
				err := r.ParseForm()
				if err != nil {
					return nil, &p.ZogIssue{Code: zconst.IssueCodeZHTTPInvalidForm, Err: err}
				}
				return form(r.Form, &formTag), nil
			}
		},
		Multipart: func(r *http.Request) p.DpFactory {
			return func() (p.DataProvider, *p.ZogIssue) {
				err := r.ParseMultipartForm(32 << 20)
				if err != nil {
					return nil, &p.ZogIssue{Code: zconst.IssueCodeZHTTPInvalidForm, Err: err}
				}

				data := make(map[string]any)
				for k, v := range r.MultipartForm.Value {
					data[k] = v[0]
				}

				for k, v := range r.MultipartForm.File {
					data[k] = v[0]
				}
				return multipartDataProvider{Data: data, tag: &multipartTag}, nil
			}
		},
		Query: func(r *http.Request) p.DpFactory {
			return func() (p.DataProvider, *p.ZogIssue) {
				// This handles generic GET request from browser. We treat it as url.Values
				return form(r.URL.Query(), &queryParam), nil
			}
		},
	},
}

type urlDataProvider struct {
	Data url.Values
	tag  *string
}

var _ p.DataProvider = urlDataProvider{}

func (u urlDataProvider) Get(key string) any {
	// if query param ends with [] its  always a slice
	if len(key) > 2 && key[len(key)-2:] == "[]" {
		return u.Data[key]
	}

	if len(u.Data[key]) > 1 {
		return u.Data[key]
	} else {
		return u.Data.Get(key)
	}
}

func (u urlDataProvider) GetByField(field reflect.StructField, fallback string) (any, string) {
	key := p.GetKeyFromField(field, fallback, u.tag)
	return u.Get(key), key
}

func (u urlDataProvider) GetNestedProvider(key string) p.DataProvider {
	return u
}
func (u urlDataProvider) GetUnderlying() any {
	return u.Data
}

type multipartDataProvider struct {
	Data map[string]any
	tag  *string
}

var _ p.DataProvider = multipartDataProvider{}

func (u multipartDataProvider) Get(key string) any {
	return u.Data[key]
}

func (u multipartDataProvider) GetByField(field reflect.StructField, fallback string) (any, string) {
	key := p.GetKeyFromField(field, fallback, u.tag)
	return u.Get(key), key
}

func (u multipartDataProvider) GetNestedProvider(key string) p.DataProvider {
	return u
}
func (u multipartDataProvider) GetUnderlying() any {
	return u.Data
}

// Parses JSON, Form & Query data from request based on Content-Type header
// Usage:
// schema.Parse(zhttp.Request(r), &dest)
// WARNING: FOR JSON PARSING DOES NOT SUPPORT JSON ARRAYS OR PRIMITIVES
func Request(r *http.Request) p.DpFactory {
	switch r.Method {
	case "GET", "HEAD":
		return Config.Parsers.Query(r)

	default:
		// Content-Type follows this format: Content-Type: <media-type>; parameter=value
		contentType := r.Header.Get("Content-Type")
		typ, _, _ := strings.Cut(contentType, ";")
		typ = strings.TrimSpace(typ)

		switch {
		case typ == "application/json":
			return Config.Parsers.JSON(r)
		case typ == "application/x-www-form-urlencoded":
			return Config.Parsers.Form(r)
		case strings.HasPrefix(contentType, "multipart/form-data"):
			return Config.Parsers.Multipart(r)
		default:
			return Config.Parsers.Query(r)
		}
	}
}

func form(data url.Values, tag *string) p.DataProvider {
	return urlDataProvider{Data: data, tag: tag}
}

// func params(data url.Values) p.DataProvider {
// 	return form(data)
// }
