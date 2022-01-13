See `cmds/main.go` for example usage.

To run example do something like this:

```
go run cmd/main.go --query=query.graphql --variables=var.json --uri=marketplace-service:deployed --user=marketplace-tf
```

Note: Must be authenticated in the dev environment (`lifeomic-aws login`)
