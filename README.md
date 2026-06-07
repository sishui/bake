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

| Category   | Expressions                                                                            | Description           |
| ---------- | -------------------------------------------------------------------------------------- | --------------------- |
| Column     | `Eq`, `Neq`                                                                            | Equality comparisons  |
| Ordered    | `Gt`, `Gte`, `Lt`, `Lte`                                                               | Ordering comparisons  |
| String     | `Like`, `LikePrefix`, `LikeSuffix`, `LikeContain`, `NotLike`, `ConcatExpr`             | Pattern matching      |
| String     | `LengthExpr`, `LowerExpr`, `UpperExpr`                                                 | String functions      |
| Numeric    | `In`, `NotIn`, `Between`, `NotBetween`                                                 | Range operations      |
| Aggregate  | `SUMExpr`, `AVGExpr`, `MINExpr`, `MAXExpr`                                             | Aggregate functions   |
| Arithmetic | `AddExpr`, `SubExpr`, `MulExpr`, `DivExpr`                                             | Arithmetic assignment |
| Arithmetic | `AddLeastExpr`, `SubGreatestExpr`, `ClampExpr`                                         | Bounded arithmetic    |
| Distinct   | `DistinctExpr`, `CountDistinctExpr`                                                    | Deduplication         |
| Nullable   | `IsNull`, `IsNotNull`, `CoalesceExpr`                                                  | NULL handling         |
| Time       | `DateExpr`, `YearExpr`, `MonthExpr`, `DayExpr`, `HourExpr`, `MinuteExpr`, `SecondExpr` | Time extraction       |
| Ordering   | `Asc`, `Desc`                                                                          | Sort order            |
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
    file: "" # log file path (optional; level defaults to "info")

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
    MaxFieldLength            int        // max field length
    MaxNullableLength         int        // max nullable length
    MaxStringLength           int        // max string length
    MaxNumericLength          int        // max numeric length
    MaxOrderedLength          int        // max ordered length
    MaxOrderedNonStringLength int        // max ordered non-string length (numeric + time)
    MaxEquatableLength        int        // max equatable length
    MaxRelationLength         int        // max relation length
    MaxArithmeticLength       int        // max arithmetic length (non-pk numeric)
    MaxTimeLength             int        // max time length
}

type Field struct {
    Imports     []string // field imports
    Name        string   // field name
    AlignedName string   // aligned field name
    Type        string   // Go type
    AlignedType string   // aligned type
    Tag         string   // field tags
    AlignedTag  string   // aligned tag
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

When using the `custom` template (default: `custom.tmpl`), the following data is passed:

```go
type CustomStruct struct {
    Version string         // bake version
    Package string         // package name
    Module  string         // module path
    Imports [][]string     // grouped imports
    Name    string         // struct name (PascalCase)
    Comment []string       // struct-level comment lines
    Fields  []*StructField // fields in this struct
}

type StructField struct {
    Name        string   // Go field name (PascalCase)
    AlignedName string   // Name padded to max width
    GoType      string   // Go type
    AlignedType string   // GoType padded to max width
    Tag         string   // Struct tag (including backticks)
    AlignedTag  string   // Tag padded to max width
    Comment     []string // Multi-line comment
}

Generated features:

- **Field alignment** — Names, types, and tags are aligned with padding
- **Multi-line comments** — Rendered before the field, tag alignment is skipped for fields with multi-line comments
- **Comment groups** — Multi-line comment fields separate struct fields into alignment groups
- **Scan/Value** — Each struct gets `Scan(src any)` and `Value() (driver.Value, error)` methods for `database/sql` compatibility

## Type Mappings

### MySQL

| MySQL Type          | Go Type         |
| ------------------- | --------------- |
| tinyint(1)          | bool            |
| tinyint             | int8 / uint8    |
| smallint            | int16 / uint16  |
| int                 | int32 / uint32  |
| bigint              | int64 / uint64  |
| float               | float32         |
| double              | float64         |
| decimal             | decimal.Decimal |
| varchar, text       | string          |
| blob                | []byte          |
| datetime, timestamp | time.Time       |
| json                | json.RawMessage |
| enum, set           | string          |
| geometry, point, linestring, polygon, etc. | []byte (WKB)    |

### PostgreSQL

| PostgreSQL Type       | Go Type                                 |
| --------------------- | --------------------------------------- |
| int2                  | int16                                   |
| int4                  | int32                                   |
| int8                  | int64                                   |
| float4                | float32                                 |
| float8                | float64                                 |
| numeric, decimal      | decimal.Decimal                         |
| bool                  | bool                                    |
| text, varchar         | string                                  |
| bytea                 | []byte                                  |
| timestamp, date, time | time.Time                               |
| json, jsonb           | json.RawMessage                         |
| uuid                  | uuid.UUID                               |
| inet, cidr            | net.IP                                  |
| interval              | time.Duration                           |
| ARRAY                 | []string, []int32, []int64, []uuid.UUID |
| geometry, geography, point, etc. (PostGIS) | []byte (WKB)   |

## Generated Bun Tags

bake automatically generates bun struct tags for each column based on the database schema.

### Index Tags

| Scenario | Generated Tag | Description |
|----------|--------------|-------------|
| Primary key | `pk,autoincrement` | Primary key column |
| Single-column unique index | `unique` | Column has a unique constraint |
| Composite unique index | `unique:index_name` | Column is part of a multi-column unique index |

Example for a composite unique index on `(start_at, end_at)`:

```go
StartAt time.Time `bun:"start_at,unique:idx_unique_range,notnull"`
EndAt   time.Time `bun:"end_at,unique:idx_unique_range,notnull"`
```

### Column Tags

| Property | Tag | Condition |
|----------|-----|-----------|
| `notnull` | Non-nullable column | `IS_NOT_NULL` or `NOT NULL` |
| `nullzero` | Nullable column | `IS_NULL` or nullable |
| `default:value` | Has default value | Column has a default |
| `soft_delete` | Soft delete column | Column name is `deleted_at` |
| `type:decimal(M,N)` | Decimal type | MySQL/PostgreSQL `decimal` columns |

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
// Post struct
type Post struct {
    bun.BaseModel `bun:"table:posts"`

    ID        int64     `bun:"id,pk,autoincrement"`
    UserID    int64     `bun:"user_id,notnull"`
    Title     string    `bun:"title,notnull"`
    User      *User     `bun:"user,rel:belongs-to,join:user_id=id"`
}

// User struct (auto-generated reverse relation)
type User struct {
    bun.BaseModel `bun:"table:users"`

    ID    int64     `bun:"id,pk,autoincrement"`
    Name  string    `bun:"name,notnull"`
    Posts []*Post   `bun:"posts,rel:has-many,join:id=user_id"`
}
```

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
// version: v0.4.0

package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// Application configuration stored as JSONB
type Config struct {
	// A detailed description
	// with multiple lines
	Description   string `json:"description,omitempty"`
	Notifications bool   `json:"notifications,omitempty"` // Enable push notifications
	Theme         string `json:"theme,omitempty"`          // UI theme (light/dark)
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

### Field Alignment

Fields within the same comment group are aligned for readability:

| Group condition | Alignment behavior |
|----------------|-------------------|
| Consecutive single-line comments | Names/types/tags are aligned to the widest value in the group |
| Multi-line comment field | The field's tag is not padded; creates a new alignment group |
| No comment | Rendered with `//` suffix and aligned normally |

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
