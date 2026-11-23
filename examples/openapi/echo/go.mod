module example/openapi/echo

go 1.26.6

replace github.com/natalie-o-perret/go-api-docs => ../../..

require (
	github.com/labstack/echo/v5 v5.4.7
	github.com/natalie-o-perret/go-api-docs v0.0.0-00010101000000-000000000000
)

require golang.org/x/time v0.15.0 // indirect
