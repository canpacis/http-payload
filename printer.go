package httppayload

import (
	"fmt"
	"net/http"
	"reflect"

	ende "github.com/canpacis/http-payload/internal/structende"
)

type Printer interface {
	Print(any) error
}

type HeaderPrinter struct {
	w http.ResponseWriter
}

func (p *HeaderPrinter) Set(key string, value any) {
	p.w.Header().Set(key, fmt.Sprintf("%s", value))
}

func (p *HeaderPrinter) Print(v any) error {
	return ende.NewEncoder(p, "header").Encode(v)
}

func NewHeaderPrinter(w http.ResponseWriter) *HeaderPrinter {
	return &HeaderPrinter{w: w}
}

type CookiePrinter struct {
	w http.ResponseWriter
}

func (p *CookiePrinter) Set(key string, value any) {
	// No-op
}

func (p *CookiePrinter) SetField(key string, value any, field reflect.StructField) {
	path, ok := field.Tag.Lookup("cookie-path")
	if !ok {
		path = "/"
	}
	var secure bool
	securestr, ok := field.Tag.Lookup("cookie-secure")
	if ok && securestr == "true" {
		secure = true
	}
	var samesite http.SameSite
	samesitestr, ok := field.Tag.Lookup("cookie-samesite")
	if ok {
		switch samesitestr {
		case "lax":
			samesite = http.SameSiteLaxMode
		case "strict":
			samesite = http.SameSiteStrictMode
		case "none":
			samesite = http.SameSiteNoneMode
		default:
			samesite = http.SameSiteDefaultMode
		}
	}

	cookieval := fmt.Sprintf("%s", value)
	http.SetCookie(p.w, &http.Cookie{
		Name:     key,
		Value:    cookieval,
		Path:     path,
		Secure:   secure,
		SameSite: samesite,
	})
}

func (p *CookiePrinter) Print(v any) error {
	return ende.NewEncoder(p, "cookie").Encode(v)
}

func NewCookiePrinter(w http.ResponseWriter) *CookiePrinter {
	return &CookiePrinter{w: w}
}

type PipePrinter []Printer

// Runs given printers in sequence
func (s *PipePrinter) Print(v any) error {
	value := v

	for _, printer := range *s {
		if err := printer.Print(value); err != nil {
			return err
		}
	}

	return nil
}

func NewPipePrinter(printers ...Printer) *PipePrinter {
	s := PipePrinter(printers)
	return &s
}
