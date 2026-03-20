# bake

# bake init
    初始化 一个简易的 bake 配置文件`bake.gen.yaml`到当前文件夹
# bake version
    查看当前bake版本
# bake
    根据配置文件生成model

# 配置文件说明:
```yaml
log:
  file: "" #生成日志，默认为空

uncountables: ["sms", "mms", "rls"] #不复数

initialisms: ["ID", "URL", "URI", "UUID", "IP"] #命名转换成

timezone: "" #created_at, updated_at, deleted_at hook的时区 

template:
  dir: "" # 自定义模板目录
  model: "model" # 自定义模板文件名
output:
  dir: "model" # 生成文件目录
  package: "model" # 生成文件包名
  module: "github.com/username/project" # import 分组时的local
db:
  - driver: "postgres" 
    dsn: "postgres://postgres:123456@127.0.0.1:5432/postgres?sslmode=disable"
    schema: "public"
    included: [] # 包含的表(默认全部表)
    excluded: [] # 排除的表(默认全部表) (优先级高于 included) 
    customs: # 自定义表设置
      test_all_types: # 表名
        comment: "test all types comment" #自定义表的注释
        tags: # 自定义tag (key, []options)
          - key: "form"
            name: "$SnakeCase"
          - key: "xml"
            name: "$CamelCase"
      mail:
        tags:
          - key: "form"
        fields:
          has_attachment: #数据库字段名
          name: "CustomHasAttachment" #自定义字段名
          tags:
            - key: "json"
              name: "#has_attachment" #自定义tag名
          attachments: # 自定义字段名
            type: "[]*MailAttachment" #自定义字段类型
            comment: "email attachments" 
            relation: true #是否关联
            tags: # 自定义tag (key, name ,[]options)
              - key: "bun"
                options: [ "rel:has-many", "join:id=mail_id" ]
```

# 生成文件说明: 可以不使用内置的模板生成model,自己在配置里指定模板,传进模板的对象结构是：
```go
type Model struct {
	Version            string     // bake version
	Package            string     // package name
	Imports            [][]string // imports
	BunModel           string     // bun.BaseModel
	Table              string     // table name
	Model              string     // model name
	Alias              string     // model alias
	Comments           []string   // model comments
	Fields             []*Field   // fields
	Timezone           string     // timezone
	CreatedAtType      string     // created_at type
	UpdatedAtType      string     // updated_at type
	DeletedAtType      string     // deleted_at type
	MaxFieldLength     int        // max field length
	MaxNullableLength  int        // max nullable length
	MaxStringLength    int        // max string length
	MaxNumericLength   int        // max numeric length
	MaxOrderedLength   int        // max ordered length
	MaxEquatableLength int        // max equatable length
	MaxRelationLength  int        // max relation length
}
```
