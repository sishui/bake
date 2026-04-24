package naming_test

import (
	"reflect"
	"testing"

	"github.com/jinzhu/inflection"
	"github.com/sishui/bake/internal/naming"
)

var initialisms = []string{"ID", "URL", "URI", "UUID", "IP"}

func TestSingular(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Should get normal singular",
			args: args{"dogs"},
			want: "dog",
		},
		{
			name: "Should get irregular singular",
			args: args{"children"},
			want: "child",
		},
		{
			name: "Should get non-countable",
			args: args{"fish"},
			want: "fish",
		},
		{
			name: "Should get added non-countable",
			args: args{"sms"},
			want: "sms",
		},
		{
			name: "Should ignore non plural",
			args: args{"test"},
			want: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inflection.AddUncountable("sms", "mms", "rls")
			if got := naming.Singular(tt.args.input); got != tt.want {
				t.Errorf("Singular() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSanitize(t *testing.T) {
	type args struct {
		raw string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "should sanitize string contains special chars",
			args: args{raw: "te$t-Str1ng0§"},
			want: "tet_Str1ng0",
		},
		{
			name: "should keep letters and numbers and dash",
			args: args{raw: "abcdef_12345-67890"},
			want: "abcdef_12345_67890",
		},
		{
			name: "should add prefix if starting with number",
			args: args{raw: "1234abcdef"},
			want: "T1234abcdef",
		},
		{
			name: "should add prefix if starting with number after sanitize",
			args: args{raw: "#1234abcdef"},
			want: "T1234abcdef",
		},
		{
			name: "should add prefix if starting with dash",
			args: args{raw: "#-1234abcdef"},
			want: "T_1234abcdef",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := naming.NormalizeIdentifier(tt.args.raw); got != tt.want {
				t.Errorf("Sanitize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCamelCase(t *testing.T) {
	type args struct {
		raw string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Should convert word to Word",
			args: args{raw: "word"},
			want: "Word",
		},
		{
			name: "Should convert word_word to WordWord",
			args: args{raw: "word_word"},
			want: "WordWord",
		},
		{
			name: "Should convert word_WORD to WordWORD",
			args: args{raw: "word_WORD"},
			want: "WordWORD",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := naming.ToCamelCase(tt.args.raw); got != tt.want {
				t.Errorf("CamelCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSnakeCased(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Should convert Word to word",
			args: args{s: "Word"},
			want: "word",
		},
		{
			name: "Should convert WordWord to word_word",
			args: args{s: "WordWord"},
			want: "word_word",
		},
		{
			name: "Should convert word_WORD to word_word",
			args: args{s: "word_WORD"},
			want: "word_word",
		},
		{
			name: "Should convert SimpleHTTPServer to simple_http_server",
			args: args{s: "SimpleHTTPServer"},
			want: "simple_http_server",
		},
		{
			name: "Should convert HTTPServer to http_server",
			args: args{s: "HTTPServer"},
			want: "http_server",
		},
		{
			name: "Should convert XMLParser to xml_parser",
			args: args{s: "XMLParser"},
			want: "xml_parser",
		},
		{
			name: "Should convert APIResponse to api_response",
			args: args{s: "APIResponse"},
			want: "api_response",
		},
		{
			name: "Should convert UserID to user_id",
			args: args{s: "UserID"},
			want: "user_id",
		},
		{
			name: "Should convert getHTTPResponse to get_http_response",
			args: args{s: "getHTTPResponse"},
			want: "get_http_response",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := naming.ToSnakeCase(tt.args.s); got != tt.want {
				t.Errorf("SnakeCased() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTableNameToStructName(t *testing.T) {
	type args struct {
		raw string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Should generate from simple word",
			args: args{"users"},
			want: "User",
		},
		{
			name: "Should generate from simple word end with es",
			args: args{"companies"},
			want: "Company",
		},
		{
			name: "Should generate from simple word end with es",
			args: args{"glasses"},
			want: "Glass",
		},
		{
			name: "Should generate from non-countable",
			args: args{"audio"},
			want: "Audio",
		},
		{
			name: "Should generate from underscored",
			args: args{"user_orders"},
			want: "UserOrder",
		},
		{
			name: "Should generate from camelCased",
			args: args{"userOrders"},
			want: "UserOrder",
		},
		{
			name: "Should generate from plural in last place",
			args: args{"usersWithOrders"},
			want: "UsersWithOrder",
		},
		{
			name: "Should generate from abracadabra",
			args: args{"abracadabra"},
			want: "Abracadabra",
		},
		{
			name: "Should generate from abracadabra_users",
			args: args{"abracadabra_users"},
			want: "AbracadabraUser",
		},
		{
			name: "Should generate from children",
			args: args{"children"},
			want: "Child",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := naming.TableToStruct(tt.args.raw); got != tt.want {
				t.Errorf("TableNameToStructName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColumnNameToFieldName(t *testing.T) {
	type args struct {
		raw string
	}
	n := naming.New(initialisms...)
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Should generate from simple word",
			args: args{"title"},
			want: "Title",
		},
		{
			name: "Should generate from underscored",
			args: args{"short_title"},
			want: "ShortTitle",
		},
		{
			name: "Should generate from camelCased",
			args: args{"shortTitle"},
			want: "ShortTitle",
		},
		{
			name: "Should generate with underscored id",
			args: args{"location_id"},
			want: "LocationID",
		},
		{
			name: "Should generate with camelCased id",
			args: args{"locationId"},
			want: "LocationID",
		},
		{
			name: "Should generate with underscored ids",
			args: args{"location_ids"},
			want: "LocationIDs",
		},
		{
			name: "Should generate with camelCased Urls",
			args: args{"shortUrls"},
			want: "ShortURLs",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := n.ColumnToField(tt.args.raw); got != tt.want {
				t.Errorf("ColumnNameToFieldName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertCommentToLines(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Should convert comment to lines",
			args: args{s: "This is a comment\nWith multiple lines"},
			want: []string{"This is a comment", "With multiple lines"},
		},
		{
			name: "Should convert comment to lines with leading and trailing spaces",
			args: args{s: "  This is a comment\n  With multiple lines  "},
			want: []string{"This is a comment", "With multiple lines"},
		},
		{
			name: "Should convert comment to lines with leading and trailing tabs",
			args: args{s: "\tThis is a comment\n\tWith multiple lines\t"},
			want: []string{"This is a comment", "With multiple lines"},
		},
		{
			name: "Should convert comment to lines with leading and trailing tabs and spaces",
			args: args{s: "  \tThis is a comment\n  \tWith multiple lines\t  "},
			want: []string{"This is a comment", "With multiple lines"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := naming.SplitCommentLines(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertCommentToLines() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPadRight(t *testing.T) {
	type args struct {
		s string
		n int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Should pad right with spaces",
			args: args{"test", 10},
			want: "test      ",
		},
		{
			name: "Should pad right with spaces when len < n",
			args: args{"test", 6},
			want: "test  ",
		},
		{
			name: "Not pad if already longer than n",
			args: args{"test", 4},
			want: "test",
		},
		{
			name: "Not pad if n is negative",
			args: args{"test", -1},
			want: "test",
		},
		{
			name: "Should align by rune width for multibyte chars",
			args: args{"用户", 6},
			want: "用户    ",
		},
		{
			name: "Should align mixed ascii and multibyte chars",
			args: args{"ab名", 6},
			want: "ab名   ",
		},
		{
			name: "Should not pad if rune count equals maxN",
			args: args{"用户名", 3},
			want: "用户名",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := naming.Align(tt.args.s, tt.args.n); got != tt.want {
				t.Errorf("PadRight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTableToAlias(t *testing.T) {
	tests := []struct {
		name  string
		table string
		want  string
	}{
		{
			name:  "Should add alias suffix",
			table: "users",
			want:  "users_alias",
		},
		{
			name:  "Should handle already aliased",
			table: "user_orders",
			want:  "user_orders_alias",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := naming.TableToAlias(tt.table); got != tt.want {
				t.Errorf("TableToAlias() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStructToReceiver(t *testing.T) {
	tests := []struct {
		name       string
		structName string
		want       string
	}{
		{
			name:       "Should convert to lowercase first char",
			structName: "User",
			want:       "u",
		},
		{
			name:       "Should handle multi-char name",
			structName: "UserOrder",
			want:       "u",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := naming.StructToReceiver(tt.structName); got != tt.want {
				t.Errorf("StructToReceiver() = %v, want %v", got, tt.want)
			}
		})
	}
}
