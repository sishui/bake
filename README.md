# bake

A database model code generator for Go. Generate ORM models from database schema using [bun](https://github.com/uptrace/bun) framework.

## Features

- Supports MySQL and PostgreSQL databases
- Automatic Go type mapping for database columns
- Customizable templates
- Custom field names, types, comments, and tags
- Table relationship support (has-many, has-one)
- Naming conventions (snake_case, camelCase)
- Timezone support for created_at, updated_at, deleted_at hooks

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

| Command | Description |
|---------|-------------|
| `bake init` | Initialize a configuration file in current directory |
| `bake version` | Show current version |
| `bake` | Generate models based on configuration |

## Configuration

### Basic Configuration

```yaml
log:
  level: "info"  # debug, info, warn, error
  file: ""       # log file path (optional)

initialisms: ["ID", "URL", "URI", "UUID", "IP"]  # naming conventions

timezone: "Asia/Shanghai"  # timezone for time hooks

template:
  dir: ""        # custom template directory
  model: "model" # template filename

output:
  dir: "model"                         # output directory
  package: "model"                     # package name
  module: "github.com/username/project" # module path for imports

db:
  - driver: "postgres"
    dsn: "postgres://user:pass@localhost:5432/db?sslmode=disable"
    schema: "public"
    included: []    # tables to include (default: all)
    excluded: []    # tables to exclude (default: none)
```

### Custom Table Settings

```yaml
db:
  - driver: "postgres"
    # ...
    customs:
      users:
        comment: "User table"  # custom table comment
        tags:
          - key: "form"
            name: "$SnakeCase"  # convert to snake_case
          - key: "xml"
            name: "$CamelCase"  # convert to camelCase
        fields:
          created_at:  # database column name
            name: "CreatedAt"   # custom field name
            tags:
              - key: "json"
                name: "created_at"  # custom json tag name
      posts:
        fields:
          author_id:
            name: "Author"         # custom field name
            type: "*User"          # custom field type
            relation: true        # is a relation
            tags:
              - key: "bun"
                options: ["rel:belongs", "join:author_id=id"]
```

### Tag Name Transformations

| Value | Description |
|-------|-------------|
| `$SnakeCase` | Convert to snake_case |
| `$CamelCase` | Convert to camelCase |
| `#field_name` | Use literal value |

## Template Data Structure

You can use custom templates. The following data is passed to templates:

```go
type Model struct {
    Version            string     // bake version
    Package            string     // package name
    Imports            [][]string // grouped imports
    BunModel           string     // bun.BaseModel
    Table              string     // table name
    Model              string     // model struct name
    Alias              string     // model alias
    Comments           []string   // model comments
    Fields             []*Field   // model fields
    Timezone           string     // timezone
    CreatedAtType      string     // created_at field type
    UpdatedAtType      string     // updated_at field type
    DeletedAtType      string     // deleted_at field type
    MaxFieldLength     int        // max field name length
    MaxNullableLength  int        // max nullable field length
    MaxStringLength    int        // max string field length
    MaxNumericLength   int        // max numeric field length
    MaxOrderedLength   int        // max ordered field length
    MaxEquatableLength int        // max equatable field length
    MaxRelationLength  int        // max relation field length
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

| MySQL Type | Go Type |
|------------|---------|
| tinyint(1) | bool |
| tinyint | int8 / uint8 |
| smallint | int16 / uint16 |
| int | int32 / uint32 |
| bigint | int64 / uint64 |
| float | float32 |
| double | float64 |
| decimal | decimal.Decimal |
| varchar, text | string |
| blob | []byte |
| datetime, timestamp | time.Time |
| json | json.RawMessage |
| enum, set | string |

### PostgreSQL

| PostgreSQL Type | Go Type |
|-----------------|---------|
| int2 | int16 |
| int4 | int32 |
| int8 | int64 |
| float4 | float32 |
| float8 | float64 |
| numeric, decimal | decimal.Decimal |
| bool | bool |
| text, varchar | string |
| bytea | []byte |
| timestamp, date, time | time.Time |
| json, jsonb | json.RawMessage |
| uuid | uuid.UUID |
| inet, cidr | net.IP |
| interval | time.Duration |
| ARRAY | []string, []int32, []int64, []uuid.UUID |

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
