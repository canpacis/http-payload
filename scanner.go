package httppayload

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"

	ende "github.com/canpacis/http-payload/internal/structende"
)

// Scanner interface resembles a json parser, it populates the given struct with available values based on its field tags. It should return an error when v is not a struct.
type Scanner interface {
	Scan(any) error
}

// A scanner to scan json value from an `io.Reader` to a struct
type JSONScanner struct {
	r io.Reader
}

// Scans the json onto v
func (s *JSONScanner) Scan(v any) error {
	return json.NewDecoder(s.r).Decode(v)
}

func NewJSONScanner(r io.Reader) *JSONScanner {
	return &JSONScanner{
		r: r,
	}
}

func NewJSONScannerFromBytes(b []byte) *JSONScanner {
	return &JSONScanner{
		r: bytes.NewBuffer(b),
	}
}

// A scanner to scan header values from an `http.Header` to a struct
type HeaderScanner struct {
	*http.Header
}

func (h *HeaderScanner) Get(key string) any {
	return h.Header.Get(key)
}

// Scans the headers onto v
func (s *HeaderScanner) Scan(v any) error {
	return ende.NewDecoder(s, "header").Decode(v)
}

func NewHeaderScanner(h *http.Header) *HeaderScanner {
	return &HeaderScanner{
		Header: h,
	}
}

// A scanner to scan url query values from a `*url.Values` to a struct
type QueryScanner struct {
	*url.Values
}

func (v QueryScanner) Get(key string) any {
	return v.Values.Get(key)
}

func (v QueryScanner) Cast(from any, to reflect.Type) (any, error) {
	return ende.DefaultCast(from, to)
}

// Scans the query values onto v
func (s *QueryScanner) Scan(v any) error {
	return ende.NewDecoder(s, "query").Decode(v)
}

func NewQueryScanner(v url.Values) *QueryScanner {
	return &QueryScanner{
		Values: &v,
	}
}

// A scanner to scan http cookies for a url from a `http.CookieJar` to a struct
type CookieScanner struct {
	cookies []*http.Cookie
}

func (v CookieScanner) Get(key string) any {
	for _, cookie := range v.cookies {
		if cookie.Name == key {
			return cookie.Value
		}
	}

	return nil
}

// Scans the cookie values onto v
func (s *CookieScanner) Scan(v any) error {
	return ende.NewDecoder(s, "cookie").Decode(v)
}

func NewCookieScanner(cookies []*http.Cookie) *CookieScanner {
	return &CookieScanner{
		cookies: cookies,
	}
}

// A scanner to scan form values from a `*url.Values` to a struct
type FormScanner struct {
	*url.Values
}

func (v FormScanner) Get(key string) any {
	return v.Values.Get(key)
}

func (v FormScanner) Cast(from any, to reflect.Type) (any, error) {
	return ende.DefaultCast(from, to)
}

// Scans the form data onto v
func (s *FormScanner) Scan(v any) error {
	return ende.NewDecoder(s, "form").Decode(v)
}

func NewFormScanner(v *url.Values) *FormScanner {
	return &FormScanner{
		Values: v,
	}
}

// A scanner to scan path parameters from a `*http.Request` to a struct
type PathScanner struct {
	*http.Request
}

func (v PathScanner) Get(key string) any {
	return v.PathValue(key)
}

func (v PathScanner) Cast(from any, to reflect.Type) (any, error) {
	return ende.DefaultCast(from, to)
}

// Scans the path parameters onto v
func (s *PathScanner) Scan(v any) error {
	return ende.NewDecoder(s, "path").Decode(v)
}

func NewPathScanner(req *http.Request) *PathScanner {
	return &PathScanner{
		Request: req,
	}
}

type MultipartValues struct {
	Files map[string]multipart.File
}

func (v MultipartValues) Get(key string) any {
	return v.Files[key]
}

type MultipartParser interface {
	ParseMultipartForm(int64) error
	FormFile(string) (multipart.File, *multipart.FileHeader, error)
}

// MultipartValuesFromParser takes a generic parser that is usually an `*http.Request` and
// returns `*scanner.MultipartValues` to use it with a `scanner.MultipartScanner` or `scanner.ImageScanner`
func MultipartValuesFromParser(p MultipartParser, size int64, names ...string) (*MultipartValues, error) {
	if err := p.ParseMultipartForm(size); err != nil {
		return nil, err
	}

	files := map[string]multipart.File{}

	for _, name := range names {
		file, _, err := p.FormFile(name)
		if err != nil {
			return nil, err
		}
		files[name] = file
	}

	return &MultipartValues{Files: files}, nil
}

// A scanner to scan multipart form values, files, from a `*scanner.MultipartValues` to a struct
// You can create a `*scanner.MultipartValues` instance with the `scanner.MultipartValuesFromParser` function.
type MultipartScanner struct {
	v *MultipartValues
}

// Scans the multipart form data onto v
func (s *MultipartScanner) Scan(v any) error {
	return ende.NewDecoder(s.v, "multipart").Decode(v)
}

func NewMultipartScanner(v *MultipartValues) *MultipartScanner {
	return &MultipartScanner{
		v: v,
	}
}

type PipeScanner []Scanner

// Runs given scanners in sequence
func (s *PipeScanner) Scan(v any) error {
	value := v

	for _, scanner := range *s {
		if err := scanner.Scan(value); err != nil {
			return err
		}
	}

	return nil
}

func NewPipeScanner(scanners ...Scanner) *PipeScanner {
	s := PipeScanner(scanners)
	return &s
}
