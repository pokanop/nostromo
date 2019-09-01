package log

type fieldStyle int

const (
	keyFieldStyle fieldStyle = iota
	valueFieldStyle
	headerFieldStyle
	cellFieldStyle
)

// FieldMapper type returns fields as a map for logging
type FieldMapper interface {
	Fields() map[string]interface{}
}

// Fields logs key value pairs formatted on a single line
func Fields(mapper FieldMapper) {

}

// Table logs key value pairs as a table with keys for the header
func Table(mapper FieldMapper) {

}
