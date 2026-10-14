# bake

A database model code generator for Go. Generate ORM models from database schema using [bun](https://github.com/uptrace/bun) framework.

## Features

- Supports MySQL and PostgreSQL databases
- Automatic Go type mapping for database columns
- Automatic foreign key detection and relation generation (belongs-to, has-many)
- Composite unique index detection (`unique`, `unique:index_name` tags)
- Customizable templates
- Custom struct generation — define non-DB structs for JSON columns, value objects, configuration objects
- Custom field names, types, comments, tags, and multi-line comments
- Naming conventions (snake_case, camelCase, consecutive acronyms)
- Timezone support for created_at, updated_at, deleted_at hooks
- Environment-based configuration (e.g., bake.gen.dev.yaml for env=dev)
- Concurrent table processing
- Auto-detect database driver from DSN scheme

### Generated Expressions

For each table, bake generates two files:
- `<table>.gen.go` — base constants, struct, time hooks
- `<table>.alias.gen.go` — alias-prefixed constants for joined queries

| Category   | Expressions                                                                                | Description           |
|------------|--------------------------------------------------------------------------------------------|-----------------------|
| Column     | `Eq`, `Neq`                                                                                | Equality comparisons  |
| Ordered    | `Gt`, `Gte`, `Lt`, `Lte`                                                                   | Ordering comparisons  |
| String     | `Like`, `LikePrefix`, `LikeSuffix`, `LikeContain`, `NotLike`, `ConcatExpr`                 | Pattern matching      |
| String     | `LengthExpr`, `LowerExpr`, `UpperExpr`                                                     | String functions      |
| Numeric    | `In`, `NotIn`, `Between`, `NotBetween`                                                     | Range operations      |
| Aggregate  | `SUMExpr`, `AVGExpr`, `MINExpr`, `MAXExpr`                                                 | Aggregate functions   |
| Arithmetic | `AddExpr`, `SubExpr`, `MulExpr`, `DivExpr`                                                 | Arithmetic assignment |
| Arithmetic | `AddLeastExpr`, `SubGreatestExpr`, `ClampExpr`                                             | Bounded arithmetic    |
| Distinct   | `DistinctExpr`, `CountDistinctExpr`                                                        | Deduplication         |
| Nullable   | `IsNull`, `IsNotNull`, `CoalesceExpr`                                                      | NULL handling         |
| Time       | `DateExpr`, `YearExpr`, `MonthExpr`, `DayExpr`, `HourExpr`, `MinuteExpr`, `SecondExpr`     | Time extraction       |
| Ordering   | `Asc`, `Desc`                                                                              | Sort order            |
| Join       | `InnerJoin`, `LeftJoin`, `RightJoin`, `FullJoin` (alias file, `FullJoin` skipped on MySQL) | JOIN helpers          |

## Installation

```shell
go install github.com/sishui/bake/cmd/bake@latest
```

## Quick Start

1. Initialize a configuration file:

```shell
bake init
```

2. Edit `bake.gen.yaml` with your database connection details

3. Generate models:

```shell
bake
```

## Commands

| Command        | Description                                          |
| -------------- | ---------------------------------------------------- |
| `bake init`    | Initialize a configuration file in current directory |
| `bake version` | Show current version                                 |
| `bake`         | Generate models based on configuration               |

## Configuration

### Basic Configuration

```yaml
log:
    level: "info" # debug, info, warn, error (default: info)
    file: "" # log file path (optional)

uncountables: ["sms", "mms", "rls"] # words that should not be pluralized

initialisms: ["ID", "URL", "URI", "UUID", "IP"] # naming conventions

timezone: "Asia/Shanghai" # timezone for time hooks

template:
    dir: "" # custom template directory

output:
    dir: "model" # output directory
    package: "model" # package name
    module: "github.com/username/project" # module path for imports

custom: [] # custom struct definitions (not tied to db tables)

db:
    - driver: "postgres"
      dsn: "postgres://user:pass@localhost:5432/db?sslmode=disable"
      schema: "public" # required for postgres, omitted for mysql
      include: [] # tables to include (default: all)
      exclude: [] # tables to exclude (default: none)
      custom: {} # per-table custom field/tag overrides
```

### Environment-based Configuration

When `.env` contains `env=dev`, bake looks for `bake.gen.dev.yaml` first, then falls back to `bake.gen.yaml`.

### Custom Table Settings

```yaml
db:
    - driver: "postgres"
      # ...
      custom:
          users:
              comment: "User table" # custom table comment
              tags:
                  - key: "form"
                    name: "$SnakeCase" # convert to snake_case
                  - key: "xml"
                    name: "$CamelCase" # convert to camelCase
              fields:
                  created_at: # database column name
                      name: "CreatedAt" # custom field name
                      tags:
                          - key: "json"
                            name: "created_at" # custom json tag name
          posts:
              fields:
                  author_id:
                      name: "Author" # custom field name
                      type: "*User" # custom field type
                      relation: true # is a relation
                      tags:
                          - key: "bun"
                            options: ["rel:belongs", "join:author_id=id"]
```

### Tag Name Transformations

| Value         | Description           |
|---------------|-----------------------|
| `$SnakeCase`  | Convert to snake_case |
| `$CamelCase`  | Convert to camelCase  |
| `#field_name` | Use literal value     |

## Template Data Structure

You can use custom templates. The following data is passed to templates:

```go
type Model struct {
    Version                   string     // bake version
    Module                    string     // module path
    Package                   string     // package name
    Imports                   [][]string // imports
    BunModel                  string     // bun.BaseModel
    Driver                    string     // database driver: mysql, postgres
    Table                     string     // table name
    Model                     string     // model name
    Alias                     string     // model alias
    Comments                  []string   // model comments
    Fields                    []*Field   // fields
    Timezone                  string     // timezone
    CreatedAtType             string     // created_at type
    UpdatedAtType             string     // updated_at type
    DeletedAtType             string     // deleted_at type
    HasField            bool       // has at least one field
    HasNullable         bool       // has a nullable field
    HasString           bool       // has a STRING field
    HasOrdered          bool       // has an ordered field (string/numeric/time)
    HasOrderedNonString bool       // has a numeric or time field
    HasEquatable        bool       // has an equatable field (numeric/string/bool/time/enum)
    HasRelation         bool       // has a relation field
    HasArithmetic       bool       // has a non-primary numeric field
    HasTime             bool       // has a TIME field
}

type Field struct {
    Imports     []string // field imports
    Name        string   // field name
    Type        string   // Go type
    Tag         string   // field tags
    Comments    []string // field comments
    ColumnName  string   // database column name
    Kind        string   // field kind: NUMERIC, STRING, TIME, etc.
    IsPrimary   bool     // is primary key
    IsNullable  bool     // is nullable
    IsCustom    bool     // is custom field
    IsRelation  bool     // is relation field
}
```

### Custom Struct Template Data

The `custom` template (default: `custom.tmpl`) receives the same `Model` and
`Field` structures as the model template. `Model.Model` holds the struct name
and `Model.Comments` the struct-level comment lines.

Generated features:

- **Multi-line comments** — Rendered before the field
- **Scan/Value** — Each struct gets `Scan(src any)` and `Value() (driver.Value, error)` methods for `database/sql` compatibility

## Type Mappings

Nullable columns are emitted as pointers (`*T`) unless the Go type's zero value
already represents SQL NULL (`[]byte`, `json.RawMessage`).

### MySQL

| MySQL Type                                  | Go Type         |
| ------------------------------------------- | --------------- |
| tinyint(1)                                  | bool            |
| tinyint                                     | int8 / uint8    |
| smallint                                    | int16 / uint16  |
| int, integer, mediumint                     | int32 / uint32  |
| bigint                                      | int64 / uint64  |
| float                                       | float32         |
| double, real                                | float64         |
| decimal                                     | decimal.Decimal |
| char, varchar, text (all sizes)             | string          |
| binary, varbinary, blob (all sizes), bit    | []byte          |
| date, time, datetime, timestamp             | time.Time       |
| year                                        | int             |
| json                                        | json.RawMessage |
| enum, set                                   | string          |
| geometry (all spatial types report this)    | []byte (WKB)    |

`unsigned` variants map to the unsigned Go type.

### PostgreSQL

| PostgreSQL Type                                                                            | Go Type                                 |
|--------------------------------------------------------------------------------------------|-----------------------------------------|
| int2                                                                                       | int16                                   |
| int4                                                                                       | int32                                   |
| int8                                                                                       | int64                                   |
| float4                                                                                     | float32                                 |
| float8                                                                                     | float64                                 |
| numeric, decimal                                                                           | decimal.Decimal                         |
| bool                                                                                       | bool                                    |
| text, varchar                                                                              | string                                  |
| bytea                                                                                      | []byte                                  |
| timestamp, date, time                                                                      | time.Time                               |
| json, jsonb                                                                                | json.RawMessage                         |
| uuid                                                                                       | uuid.UUID                               |
| inet, cidr                                                                                 | net.IP                                  |
| interval                                                                                   | time.Duration                           |
| ARRAY                                                                                      | []string, []int32, []int64, []uuid.UUID |
| geometry, geography, point, linestring, polygon, multipoint, multilinestring, multipolygon | []byte (WKB)                            |

## Generated Tags

bake generates a `bun` and a `json` struct tag for each column based on the database schema.

### Index Tags

| Scenario                   | Generated Tag       | Description                                   |
|----------------------------|---------------------|-----------------------------------------------|
| Primary key                | `pk`                | Primary key column                            |
| Auto-increment column      | `autoincrement`     | Auto-increment / identity column              |
| Single-column unique index | `unique`            | Column has a unique constraint                |
| Composite unique index     | `unique:index_name` | Column is part of a multi-column unique index |

Example for a composite unique index on `(start_at, end_at)`:

```go
StartAt time.Time `bun:"start_at,unique:idx_unique_range,notnull"`
EndAt   time.Time `bun:"end_at,unique:idx_unique_range,notnull"`
```

### Column Tags

| Property            | Tag                 | Condition                          |
|---------------------|---------------------|------------------------------------|
| `notnull`           | Non-nullable column | `IS_NOT_NULL` or `NOT NULL`        |
| `nullzero`          | Nullable column     | `IS_NULL` or nullable              |
| `default:value`     | Has default value   | Column has a default               |
| `soft_delete`       | Soft delete column  | Column name is `deleted_at`        |
| `type:decimal(M,N)` | Decimal type        | MySQL/PostgreSQL `decimal` columns |

### JSON Tags

Every column also gets a `json` tag: `json:"<column_name>,omitempty"`, with
`json:"-"` for `deleted_at`. Override it per column with a `custom` tag using
`key: "json"`.

## Foreign Key Relations

bake automatically detects foreign key relationships from your database schema and generates the appropriate bun relation tags.

### How It Works

When a column has a foreign key constraint:

```sql
-- posts.user_id references users.id
ALTER TABLE posts ADD CONSTRAINT fk_posts_user_id
  FOREIGN KEY (user_id) REFERENCES users(id);
```

bake will automatically:

1. On `posts` table: Generate `User *User` field with `rel:belongs-to`
2. On `users` table: Generate `Posts []*Post` field with `rel:has-many`

### Generated Example

For the `posts` table with `user_id` foreign key:

```go
type Post struct {
	bun.BaseModel `bun:"table:posts,alias:posts_alias"`

	ID        int64     `bun:"id,pk,autoincrement,notnull" json:"id,omitempty"`                             //
	UserID    int64     `bun:"user_id,notnull" json:"user_id,omitempty"`                                    // 用户ID
	Title     string    `bun:"title,notnull" json:"title,omitempty"`                                        // 标题
	Content   *string   `bun:"content,nullzero" json:"content,omitempty"`                                   // 内容
	Status    int8      `bun:"status,notnull,default:'0'" json:"status,omitempty"`                          // 状态: 0=草稿, 1=发布
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp(3)" json:"created_at,omitempty"` // 创建时间
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp(3)" json:"updated_at,omitempty"` // 更新时间
	User      *User     `bun:"user,rel:belongs-to,join:user_id=id" json:"user,omitempty"`                   //
}

type User struct {
	bun.BaseModel `bun:"table:users,alias:users_alias"`

	ID        int64     `bun:"id,pk,autoincrement,notnull" json:"id,omitempty"`                             //
	Name      string    `bun:"name,notnull" json:"name,omitempty"`                                          // 用户名
	Email     string    `bun:"email,unique,notnull" json:"email,omitempty"`                                 // 邮箱
	StartAt   time.Time `bun:"start_at,notnull" json:"start_at,omitempty"`                                  //
	EndAt     time.Time `bun:"end_at,notnull" json:"end_at,omitempty"`                                      //
	Seq       int64     `bun:"seq,notnull" json:"seq,omitempty"`                                            //
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp(3)" json:"created_at,omitempty"` // 创建时间
	Posts     []*Post   `bun:"posts,rel:has-many,join:id=user_id" json:"posts,omitempty"`                   //
}
```

`rel:belongs-to` is generated on the table holding the foreign key, and the
matching `rel:has-many` slice on the referenced table. Alongside each model,
`<table>.alias.gen.go` carries the join helpers:

```go
const PostTableAlias = "posts_alias"

const PostTableNameAsAlias = "posts AS posts_alias"

const (
	InnerJoinPostAsAlias = "INNER JOIN posts AS posts_alias" // same as JOIN
	LeftJoinPostAsAlias  = "LEFT JOIN posts AS posts_alias"  // same as LEFT OUTER JOIN
	RightJoinPostAsAlias = "RIGHT JOIN posts AS posts_alias" // same as RIGHT OUTER JOIN
)
```

`FullJoin` is emitted for PostgreSQL only, since MySQL has no `FULL JOIN`.

### Manual Override

You can still manually configure relations using the `custom` configuration. Manual configuration takes precedence over automatic detection.

## Custom Structs

Custom structs allow you to define Go structs that are not tied to database tables. They are generated with `Scan` and `Value` methods for `database/sql` compatibility, making them ideal for JSON/JSONB column types.

### Usage

Define custom structs in your `bake.gen.yaml`:

```yaml
output:
  dir: "model"
  package: "model"
  module: "github.com/user/project"

custom:
  - name: "Config"
    comment: "Application configuration stored as JSONB"
    fields:
      - name: "Theme"
        type: "string"
        comment: "UI theme (light/dark)"
      - name: "Notifications"
        type: "bool"
        comment: "Enable push notifications"
      - name: "Description"
        type: "string"
        comment: "A detailed description\nwith multiple lines"
```

### Generated Output

Running `bake` generates one file per custom struct (e.g., `model/config.gen.go`):

```go
// Code generated by bake. DO NOT EDIT.
// version: v0.4.2
// source: config

// Package model defines database models generated from schema.
package model

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

var (
	_ sql.Scanner   = (*Config)(nil)
	_ driver.Valuer = (*Config)(nil)
)

// Application configuration stored as JSONB
type Config struct {
	// A detailed description
	// with multiple lines
	Description   string `json:"description,omitempty"`
	Notifications bool   `json:"notifications,omitempty"` // Enable push notifications
	Theme         string `json:"theme,omitempty"`         // UI theme (light/dark)
}

func (o *Config) Scan(src any) error {
	if src == nil {
		return nil
	}
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, o)
	case string:
		return json.Unmarshal([]byte(v), o)
	default:
		return fmt.Errorf("unsupported scan type %T for Config", src)
	}
}

func (o Config) Value() (driver.Value, error) {
	return json.Marshal(o)
}
```

### Formatting

Generated source is passed through `go/format` before being written, so field
names, types and tags are aligned exactly as `gofmt` would align them. There is
no template-level padding to configure.

### Example

See [examples/mysql/](examples/mysql/) or [examples/postgres/](examples/postgres/) for complete working examples with custom structs integrated alongside database models.


## Development

### Running Tests

```shell
go test ./...
```

### Building

```shell
go build ./cmd/bake
```

## License

Apache License 2.0 - see [LICENSE](LICENSE) file
