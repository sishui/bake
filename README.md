# bake

A database model code generator for Go. Generate ORM models from database schema using [bun](https://github.com/uptrace/bun) framework.

## Features

- Supports MySQL and PostgreSQL databases
- Automatic Go type mapping for database columns
- Relation fields declared with bun tags (`rel:belongs-to`, `rel:has-one`, `rel:has-many`, `m2m`)
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

### Index Constants

Each non-primary index is emitted once, under its own name, so composite
indexes are covered as well:

```go
const (
	UserIndexSeq      = "idx_seq"
	UserIndexStartEnd = "idx_start_end"
	UserIndexUkEmail  = "uk_email"
)
```

The constant suffix is the index name in CamelCase, with a leading `idx_`
dropped so the name does not repeat "Index". The index backing the primary key
is skipped, whatever the driver calls it.

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

imports: [] # type -> package mapping for custom field types (see Imports)

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
                      tags:
                          - key: "bun"
                            options: ["rel:belongs-to", "join:author_id=id"]
```

A field is treated as a relation when its `bun` tag carries one of
`rel:belongs-to`, `rel:has-one`, `rel:has-many`, or an `m2m:<join_table>`
option. There is no separate flag to set.

### Imports

A custom field type that comes from another package needs an import. Declare it
once at the top level; bake matches it against the field `type` and adds the
import to the generated file.

```yaml
imports:
    - package: "github.com/google/uuid"
      types: ["uuid.UUID"]
    - package: "github.com/hashicorp/go-version"
      types: ["version.Version"]
    - package: "github.com/shopspring/decimal"
      alias: "dec" # emits: dec "github.com/shopspring/decimal"
      types: ["dec.Decimal"]
```

- `types` lists the qualified type names exactly as they appear in a field
  `type`; pointers, slices and maps around them still match (`*version.Version`,
  `[]uuid.UUID`).
- `alias` only changes the emitted import line. When it is set, every entry in
  `types` must use that alias.
- Without `alias` the import is emitted as-is, so the qualifier in `types` must
  be the package's own name, which can differ from the last path element
  (`github.com/hashicorp/go-version` is used as `version`, not `go-version`).

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
    Version             string     // bake version
    Module              string     // module path
    Package             string     // package name
    Imports             [][]Import // imports grouped by std / third-party / local
    BunModel            string     // bun.BaseModel
    Driver              string     // database driver: mysql, postgres
    Table               string     // table name
    Model               string     // model name
    Alias               string     // model alias
    Comments            []string   // model comments
    Fields              []*Field   // fields
    Timezone            string     // timezone
    CreatedAtType       string     // created_at type
    UpdatedAtType       string     // updated_at type
    DeletedAtType       string     // deleted_at type
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

type Import struct {
    Path  string // import path, e.g. "github.com/google/uuid"
    Alias string // import alias, empty unless the package is renamed
}

type Field struct {
    Imports     []Import // imports required by the field type
    Name        string   // field name
    Type        string   // Go type
    Tag         string   // field tags
    Comments    []string // field comments
    ColumnName  string   // database column name
    Kind        string   // field kind: NUMERIC, STRING, TIME, etc.
    IsPrimary   bool     // is primary key
    IsNullable  bool     // is nullable
    IsCustom    bool     // is custom field
    IsRelation  bool     // is a relation (from the bun rel:/m2m option)
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

Values are `pg_type.typname`, i.e. `int4` rather than `integer`, `float8` rather
than `double precision`. Array columns use the element type prefixed with `_`
(`_text` for `text[]`), so they are listed here by their element type.

| PostgreSQL Type                               | Go Type                                 |
|-----------------------------------------------|-----------------------------------------|
| int2                                          | int16                                   |
| int4                                          | int32                                   |
| int8                                          | int64                                   |
| float4                                        | float32                                 |
| float8                                        | float64                                 |
| numeric                                       | decimal.Decimal                         |
| bool                                          | bool                                    |
| text, varchar, char, bpchar                   | string                                  |
| bytea, bit, varbit                            | []byte                                  |
| date, timestamp, timestamptz, time, timetz    | time.Time                               |
| json, jsonb                                   | json.RawMessage                         |
| uuid                                          | uuid.UUID                               |
| inet, cidr                                    | net.IP                                  |
| interval                                      | time.Duration                           |
| text[], varchar[], int4[], int8[], uuid[]     | []string, []int32, []int64, []uuid.UUID |
| geometry, geography (PostGIS), point, polygon | []byte (WKB)                            |

### Unsupported PostgreSQL Types

These built-in types have no mapping, so a column using one fails generation
with `unsupported type` (or `unsupported array type` for arrays):

- **Arrays** — any element type other than `text`, `varchar`, `int4`, `int8` or `uuid`,
  e.g. `bool[]`, `int2[]`, `float4[]`, `float8[]`, `numeric[]`, `date[]`, `timestamptz[]`
- **Ranges** — `int4range`, `int8range`, `numrange`, `daterange`, `tsrange`, `tstzrange`
  and their `*multirange` forms
- **Other** — `money`, `xml`, `jsonpath`, `macaddr`, `macaddr8`, `tsvector`, `tsquery`
- **Native geometry** — `line`, `lseg`, `box`, `path`, `circle`
- **System / internal** — `aclitem`, `cid`, `tid`, `xid`, `xid8`, `oid`, `name`,
  `int2vector`, `oidvector`, `refcursor`, `gtsvector`, `pg_lsn`, `pg_snapshot`, `reg*`
  and the remaining `pg_*` catalog types

`geometry` and `geography` come from PostGIS and are mapped when the extension
is installed.

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

## Relations

Relations are declared in the configuration; bake does not inspect foreign keys
to create them. A field counts as a relation when its `bun` tag carries

- `rel:belongs-to`, `rel:has-one` or `rel:has-many`, or
- `m2m:<join_table>`.

Field name, Go type and `join:` option are part of the same declaration:

```yaml
db:
    - driver: "postgres"
      custom:
          posts:
              fields:
                  user: # field name on the Post model
                      type: "*User"
                      comment: "post author"
                      tags:
                          - key: "bun"
                            options: ["rel:belongs-to", "join:user_id=id"]
          users:
              fields:
                  posts:
                      type: "[]*Post"
                      tags:
                          - key: "bun"
                            options: ["rel:has-many", "join:id=user_id"]
                  profile:
                      type: "*Profile"
                      tags:
                          - key: "bun"
                            options: ["rel:has-one", "join:id=user_id"]
                  roles:
                      type: "[]*Role"
                      tags:
                          - key: "bun"
                            options: ["m2m:user_roles"]
```

A declaration is ignored when the field name matches a real column. A `json`
tag is added automatically as `json:"<field>,omitempty"`.

`type` is required on a declared field: it is not backed by a column, so there
is no database type to infer.

The generated models contain exactly the relations you declared:

```go
type Post struct {
	bun.BaseModel `bun:"table:posts,alias:posts_alias"`

	// ...
	User *User `bun:"user,rel:belongs-to,join:user_id=id" json:"user,omitempty"`
}

type User struct {
	bun.BaseModel `bun:"table:users,alias:users_alias"`

	// ...
	Posts []*Post `bun:"posts,rel:has-many,join:id=user_id" json:"posts,omitempty"`
}
```

### Join Helpers

Alongside each model, `<table>.alias.gen.go` carries the alias and join helpers:

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
