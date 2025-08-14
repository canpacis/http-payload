# Install

```shell
go get github.com/canpacis/http-payload
```

# Scanner

Scanner is a utility package to extract certain values and cast them to usable values in the context of http servers. It can extract request bodies, form values, url queries, cookies and headers. Scanner defines a `Scanner` interface for you to extend it to your own needs.

```go
type Scanner interface {
  Scan(any) error
}
```

## Example

An example that extracts header values from an `http.Header` object.

```go
type Params struct {
  // provide the header name in the field tag
  Language string `header:"accept-language"`
}

// ...

// Create your instance
p := &Params{}
// Provide the request headers
s := payload.NewHeaderScanner(r.Headers())
if err := s.Scan(p); err != nil {
  // handle error
}
// p.Language -> r.Headers().Get("Accept-Language")
```

You can compose your scanners. There are a handful of pre-built scanners for the most common of use cases.

- `payload.JSONScanner`: Scans json data from an `io.Reader`
- `payload.HeaderScanner`: Scans header data from an `*http.Header`
- `payload.QueryScanner`: Scans url query values from a `*url.Values`
- `payload.FormScanner`: Scans form data from a `*url.Values`
- `payload.CookieScanner`: Scans cookies from a `http.CookieJar`
- `payload.MultipartScanner`: Scans multipart form data from a `payload.MultipartValues`
- `payload.ImageScanner`: Scans multipart images from a `payload.MultipartValues`

eg.:

```go
type Params struct {
  IDs      []string `query:"ids"`
  Language string   `header:"accept-language"`
  Token    string   `cookie:"token"`
}

// ...

// Create your instance
p := &Params{}
var s payload.Scanner

s = payload.NewQueryScanner(/* url.Values */)
s.Scan(p) // Don't forget to handle errors

s = payload.NewHeaderScanner(/* http.Header */)
s.Scan(p) // Don't forget to handle errors

s = payload.NewCookieScanner(/* http.CookieJar */)
s.Scan(p) // Don't forget to handle errors
```

Or, alternatively, you can use a `payload.PipeScanner` to streamline the process.

```go
type Params struct {
  IDs      []string `query:"ids"`
  Language string   `header:"accept-language"`
  Token    string   `cookie:"token"`
}

// ...

// Create your instance
p := &Params{}
s := payload.NewPipeScanner(
  payload.NewQueryScanner(/* url.Values */),
  payload.NewHeaderScanner(/* http.Header */),
  payload.NewCookieScanner(/* http.CookieJar */),
)
s.Scan(p) // Don't forget to handle errors
```

This will populate your struct's fields with available values.