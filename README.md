# bake

A database model code generator for Go. Generate ORM models from database schema using [bun](https://github.com/uptrace/bun) framework.

## Features

- Supports MySQL and PostgreSQL databases
- Automatic Go type mapping for database columns
- Automatic foreign key detection and relation generation (belongs-to, has-many)
- Customizable templates
- Custom field names, types, comments, and tags
- Naming conventions (snake_case, camelCase, consecutive acronyms)
- Timezone support for created_at, updated_at, deleted_at hooks
- Environment-based configuration (e.g., bake.gen.dev.yaml for env=dev)
- Concurrent table processing
- Auto-detect database driver from DSN scheme

### Generated Expressions

| Category   | Expressions                                                | Description           |
| ---------- | ---------------------------------------------------------- | --------------------- |
| Comparison | `Eq`, `Neq`, `Gt`, `Gte`, `Lt`, `Lte`                      | Basic comparisons     |
| String     | `Like`, `LikePrefix`, `LikeSuffix`, `LikeContain`          | Pattern matching      |
| Numeric    | `In`, `NotIn`, `Between`, `NotBetween`                     | Range operations      |
| Aggregate  | `SUM`, `AVG`, `MIN`, `MAX`                                 | Aggregate functions   |
| Arithmetic | `Add`, `Sub`, `AddLeast`, `SubGreatest`, `Clamp`           | Arithmetic operations |
| Time       | `Date`, `Year`, `Month`, `Day`, `Hour`, `Minute`, `Second` | Time extraction       |
| Nullable   | `IsNull`, `IsNotNull`, `Coalesce`                          | NULL handling         |
| Ordering   | `Asc`, `Desc`                                              | Sort order            |

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
    level: "info" # debug, info, warn, error
    file: "" # log file path (optional)

uncountables: ["sms", "mms", "rls"] # words that should not be pluralized

initialisms: ["ID", "URL", "URI", "UUID", "IP"] # naming conventions

timezone: "Asia/Shanghai" # timezone for time hooks

template:
    dir: "" # custom template directory
    model: "model" # template filename

output:
    dir: "model" # output directory
    package: "model" # package name
    module: "github.com/username/project" # module path for imports

db:
    - driver: "postgres"
      dsn: "postgres://user:pass@localhost:5432/db?sslmode=disable"
      schema: "public" # required for postgres, optional for mysql
      included: [] # tables to include (default: all)
      excluded: [] # tables to exclude (default: none)
```

### Environment-based Configuration

When `.env` contains `env=dev`, bake looks for `bake.gen.dev.yaml` first, then falls back to `bake.gen.yaml`.

### Custom Table Settings

```yaml
db:
    - driver: "postgres"
      # ...
      customs:
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
	Driver                    string     // database driver (mysql, postgres)
	Module                    string     // module path
	Package                   string     // package name
	Imports                   [][]string // imports
	BunModel                  string     // bun.BaseModel
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

You can still manually configure relations using the `customs` configuration. Manual configuration takes precedence over automatic detection.

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
