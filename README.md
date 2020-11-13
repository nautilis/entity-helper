## entity-helper 
entity-helper 是一个用于生成数据库表实体的go tool, 通过定义空实体即可自动填充与数据表一致的golang字段。
### 使用方式
- 配置
在 ~/.entity-helper/conf.toml配置数据库信息
```toml
DbAddress = "username:password@tcp(127.0.0.1:3306)/yourdb?charset=utf8mb4"
Db = "yourdb"
```

- 定义实体并添加go:generate注释，指明目标结构体 指明表名
```go
//go:generate entity-helper -target User -table user_dating
type User struct {
}
```
- 在定义实体的package 目录下执行 go generate
- 如果需要重新生成entity 需要删除struct的现有字段
