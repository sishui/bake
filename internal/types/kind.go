package types

const (
	KindNumeric = "NUMERIC"
	KindString  = "STRING"
	KindBytes   = "BYTES"
	KindTime    = "TIME"
	KindEnum    = "ENUM"
	KindBoolean = "BOOL"
	KindStruct  = "STRUCT"
	KindArray   = "ARRAY"
)

func TypeToKind(typ string) string {
	switch typ {
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64",
		"*int", "*int8", "*int16", "*int32", "*int64", "*uint", "*uint8", "*uint16", "*uint32", "*uint64", "*float32", "*float64":
		return KindNumeric
	case "string", "*string":
		return KindString
	case "[]byte", "*[]byte":
		return KindBytes
	case "time.Time", "*time.Time":
		return KindTime
	case "bool", "*bool":
		return KindBoolean
	default:
		return KindStruct
	}
}
