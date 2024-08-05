# RequestFilter Plugin for Traefik

The RequestFilter plugin is a middleware for Traefik that allows filtering of HTTP requests based on multiple regular expressions. It can filter requests by examining both the URL and the request body (for POST and PUT requests).

## Features

- Filter requests based on multiple regex patterns
- Apply filters to both URL and request body, or only to the URL/body
- Support for GET, POST, PUT, and PATCH methods
- Customizable HTTP response for blocked requests

## Configuration

### Static Configuration

To enable the plugin in your Traefik static configuration:

```yaml
experimental:
  plugins:
    requestFilter:
      moduleName: github.com/aveq-research/requestfilter
      version: v1.0.0  # Replace with the actual version
```

### Dynamic Configuration

The following options are available for the RequestFilter plugin:

| Option           | Description                                        | Default value     |
| ---------------- | -------------------------------------------------- | ----------------- |
| filterRegexes    | List of regex patterns to filter requests.         | []                |
| httpErrorMessage | Custom HTTP response message for blocked requests. | "Request blocked" |
| pathOnly         | Apply filters only to the URL.                     | false             |
| bodyOnly         | Apply filters only to the request body.            | false             |

To use the plugin in your dynamic configuration:

```yaml
http:
  middlewares:
    my-request-filter:
      plugin:
        requestFilter:
          filterRegexes:
            - "sensitive"
            - "confidential"
            - "\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b" # Email pattern
          httpErrorMessage: "Access denied!"
          bodyOnly: true
```

## Example Usage

Here's an example of how to use the RequestFilter in your Traefik configuration:

```yaml
http:
  routers:
    my-router:
      rule: host(`example.com`)
      service: my-service
      middlewares:
        - my-request-filter

  services:
    my-service:
      loadBalancer:
        servers:
          - url: http://internal-service:8080

  middlewares:
    my-request-filter:
      plugin:
        requestFilter:
          filterRegexes:
            - "password"
            - "credit_card"
            - "/admin/"
            - "\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b"
```

This configuration will block requests that:

- Contain "password" or "credit_card" in the URL or body
- Try to access paths containing "/admin/"
- Contain what looks like an email address

## Development

To set up the development environment:

1. Clone the repository
2. Use `go test ./...` to run the tests

## License

Copyright 2024, AVEQ GmbH.

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
