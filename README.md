This is a simple library used to wrap common behavior when interacting with lifeomic services.
The goal is support more of a public API through the graphql proxy, but initially it only
supports direct lambda invocations.


See `cmds/main.go` for example usage.

To run example do something like this:

```
go run cmd/main.go --query=query.graphql --variables=var.json --uri=marketplace-service:deployed/v1/marketplace/authenticated/graphql --user=marketplace-tf
```