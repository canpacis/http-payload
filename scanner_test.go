package httppayload_test

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"image"
	"image/draw"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"

	payload "github.com/canpacis/http-payload"
	"github.com/stretchr/testify/assert"
)

type Role struct {
	Name string
}

func (r *Role) UnmarshalString(s string) error {
	r.Name = s
	return nil
}

type Params struct {
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`

	Language string `json:"-" header:"accept-language"`

	Page  uint32 `json:"-" query:"page" form:"page"`
	Done  bool   `json:"-" query:"done"`
	Role  Role   `json:"-" query:"role"`
	Roles []Role `json:"-" query:"roles"`

	Filters []string `json:"-" form:"filters"`
	Numbers []int    `json:"-" form:"numbers"`

	Token string `json:"-" cookie:"token"`

	Document multipart.File `json:"-" multipart:"document"`

	ID   string `json:"-" path:"id"`
	Slug string `json:"-" path:"slug"`
}

type Expectation struct {
	Expected any
	Actual   any
}

type ScannerCase struct {
	Scanner      payload.Scanner
	Expectations func(p *Params) []Expectation
}

func (c ScannerCase) Run(t *testing.T) {
	assert := assert.New(t)
	p := &Params{}

	err := c.Scanner.Scan(p)
	assert.NoError(err)

	for _, e := range c.Expectations(p) {
		assert.Equal(e.Expected, e.Actual)
	}
}

func TestJsonScanner(t *testing.T) {
	body := bytes.NewBuffer([]byte(`{ "email": "test@example.com", "name": "John Doe" }`))

	c := ScannerCase{
		Scanner: payload.NewJSONScanner(body),
		Expectations: func(p *Params) []Expectation {
			return []Expectation{
				{"test@example.com", p.Email},
				{"John Doe", p.Name},
			}
		},
	}
	c.Run(t)
}

func TestHeaderScanner(t *testing.T) {
	header := &http.Header{}
	header.Set("Accept-Language", "en")

	c := ScannerCase{
		Scanner: payload.NewHeaderScanner(header),
		Expectations: func(p *Params) []Expectation {
			return []Expectation{
				{"en", p.Language},
			}
		},
	}
	c.Run(t)
}

func TestQueryScanner(t *testing.T) {
	values := url.Values{}
	values.Set("page", "2")
	values.Set("done", "true")
	values.Set("role", "admin")
	values.Set("roles", "admin,user")

	c := ScannerCase{
		Scanner: payload.NewQueryScanner(values),
		Expectations: func(p *Params) []Expectation {
			return []Expectation{
				{uint32(2), p.Page},
				{true, p.Done},
				{"admin", p.Role.Name},
				{2, len(p.Roles)},
				{"admin", p.Roles[0].Name},
				{"user", p.Roles[1].Name},
			}
		},
	}
	c.Run(t)
}

func TestFormScanner(t *testing.T) {
	form := &url.Values{}
	form.Set("filters", "sepia,monochrome")
	form.Set("numbers", "6,7,8")

	c := ScannerCase{
		Scanner: payload.NewFormScanner(form),
		Expectations: func(p *Params) []Expectation {
			return []Expectation{
				{2, len(p.Filters)},
				{3, len(p.Numbers)},
				{"sepia", p.Filters[0]},
				{"monochrome", p.Filters[1]},
				{6, p.Numbers[0]},
				{7, p.Numbers[1]},
				{8, p.Numbers[2]},
			}
		},
	}
	c.Run(t)
}

func TestPathScanner(t *testing.T) {
	req := &http.Request{}
	req.SetPathValue("id", "this_is_id")
	req.SetPathValue("slug", "this-is-slug")

	c := ScannerCase{
		Scanner: payload.NewPathScanner(req),
		Expectations: func(p *Params) []Expectation {
			return []Expectation{
				{"this_is_id", p.ID},
				{"this-is-slug", p.Slug},
			}
		},
	}
	c.Run(t)
}

func TestCookieScanner(t *testing.T) {
	jar, _ := cookiejar.New(nil)
	url, _ := url.Parse("http://url.net")
	jar.SetCookies(url, []*http.Cookie{
		{
			Name:  "token",
			Value: "cookie-token",
		},
	})

	c := ScannerCase{
		Scanner: payload.NewCookieScanner(jar.Cookies(url)),
		Expectations: func(p *Params) []Expectation {
			return []Expectation{
				{"cookie-token", p.Token},
			}
		},
	}
	c.Run(t)
}

type file struct {
	io.Reader
	io.ReaderAt
	io.Seeker
	io.Closer
}

func TestMultipartScanner(t *testing.T) {
	multipart := &payload.MultipartValues{
		Files: map[string]multipart.File{
			"document": file{
				Reader: bytes.NewBuffer([]byte("text document")),
			},
		},
	}

	c := ScannerCase{
		Scanner: payload.NewMultipartScanner(multipart),
		Expectations: func(p *Params) []Expectation {
			file, err := io.ReadAll(p.Document)

			return []Expectation{
				{nil, err},
				{"text document", string(file)},
			}
		},
	}
	c.Run(t)
}

func hash(img image.Image) string {
	var rgba *image.RGBA
	var ok bool

	rgba, ok = img.(*image.RGBA)
	if !ok {
		rgba = image.NewRGBA(img.Bounds())
		draw.Draw(rgba, img.Bounds(), img, image.Pt(0, 0), draw.Over)
	}

	return fmt.Sprintf("%x", md5.Sum(rgba.Pix))
}
