# entity-helper [中文](./README_CN.md) #
Useful go tool to generate struct fields according to database schema

### installation

```shell
go get -u github.com/nautilis/entity-helper 
```

### Quick Start

- fill database configuration on ~/.entity-hepler/conf.toml
```toml
DbAddress = "username:password@tcp(127.0.0.1:3306)/yourdb?charset=utf8mb4"
Db = "yourdb"
```
- declare the empty entity struct and comment go:generate with target struct name and table name .
```go
//go:generate entity-helper -target User -table user_table_name
type User struct {
}
```
- execute ```go generate ``` under package.
- must delete the existed fields of the struct if you need to regenerate it.
