# go-compose

[![Build Status](https://api.travis-ci.org/ibrt/go-compose.svg?branch=master)](https://travis-ci.org/ibrt/go-compose?branch=master)
[![Coverage Status](https://coveralls.io/repos/github/ibrt/go-compose/badge.svg?branch=master)](https://coveralls.io/github/ibrt/go-compose?branch=master)

Go wrapper around Docker Compose, useful for integration testing.
Check out the [GoDoc](https://godoc.org/github.com/ibrt/go-compose/compose) for more information.

Example:

```go
var composeYML =`
test_mockserver:
  container_name: ms
  image: jamesdbloom/mockserver
  ports:
    - "10000:1080"
    - "1090"
test_postgres:
  container_name: pg
  image: postgres
  ports:
    - "5432"`

compose, err := compose.Start(composeYML, true, true)
if err != nil {
	panic(err)
}
defer compose.Kill()
```
