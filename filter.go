package filterxorm

import (
	"fmt"
	"strings"

	"github.com/go-xorm/xorm"
)

type OperType int

const (
	CONDITION_AND OperType = 1
	CONDITION_OR  OperType = 2
)

func And(conditions ...Condition) Condition {
	if len(conditions) == 1 {
		return conditions[0]
	}
	return NewConditionOper(CONDITION_AND, conditions)
}

func Or(conditions ...Condition) Condition {
	if len(conditions) == 1 {
		return conditions[0]
	}
	return NewConditionOper(CONDITION_OR, conditions)
}

type Condition interface {
	Build(*xorm.Engine) Condition
	CondiStr() string
	Values() []interface{}
	And(vs ...Condition) Condition
	Or(vs ...Condition) Condition
	Where(sess *xorm.Session) *xorm.Session
	Do(func(sess *xorm.Session))
}

func NewCondition(conditionStr string, fields []*Field, values []interface{}) Condition {
	return &condition{
		conditionStr: conditionStr,
		fields:       fields,
		values:       values,
	}
}

func NewConditionOper(oper OperType, conditions []Condition) Condition {
	return &conditionOper{
		oper:       oper,
		conditions: conditions,
	}
}

type condition struct {
	engine       *xorm.Engine
	conditionStr string
	fields       []*Field
	values       []interface{}
}

func (condi *condition) Build(engine *xorm.Engine) Condition {
	condi.engine = engine
	return condi
}
func (condi *condition) CondiStr() string {
	names := make([]interface{}, len(condi.fields))
	for i, field := range condi.fields {
		names[i] = field.Name(condi.engine)
	}
	return fmt.Sprintf(condi.conditionStr, names...)
}
func (condi *condition) Values() []interface{} { return condi.values }
func (condi *condition) And(vs ...Condition) Condition {
	conditions := make([]Condition, 1, len(vs)+1)
	conditions[0] = condi
	return And(append(conditions, vs...)...)
}
func (condi *condition) Or(vs ...Condition) Condition {
	conditions := make([]Condition, 1, len(vs)+1)
	conditions[0] = condi
	return Or(append(conditions, vs...)...)
}
func (condi *condition) Where(sess *xorm.Session) *xorm.Session {
	condi.Build(sess.Engine)
	sess.Where(condi.CondiStr(), condi.Values()...)
	return sess
}
func (condi *condition) Do(f func(sess *xorm.Session)) {
	sess := condi.engine.NewSession()
	defer sess.Close()
	sess.Where(condi.CondiStr(), condi.Values()...)
	f(sess)
}

type conditionOper struct {
	engine     *xorm.Engine
	oper       OperType
	conditions []Condition
}

func (oper *conditionOper) Build(engine *xorm.Engine) Condition {
	oper.engine = engine
	return oper
}
func (oper *conditionOper) CondiStr() string {
	r := make([]string, len(oper.conditions))
	for i, condi := range oper.conditions {
		condi.Build(oper.engine)
		r[i] = fmt.Sprintf("(%s)", condi.CondiStr())
	}
	var op string
	switch oper.oper {
	case CONDITION_AND:
		op = oper.engine.Dialect().AndStr()
	case CONDITION_OR:
		op = oper.engine.Dialect().OrStr()
	}
	return strings.Join(r, fmt.Sprintf(" %s ", op))
}
func (oper *conditionOper) Values() (vs []interface{}) {
	vs = make([]interface{}, 0, len(oper.conditions)<<1)
	for _, condi := range oper.conditions {
		condi.Build(oper.engine)
		vs = append(vs, condi.Values()...)
	}
	return
}
func (oper *conditionOper) And(vs ...Condition) Condition { return And(vs...) }
func (oper *conditionOper) Or(vs ...Condition) Condition  { return Or(vs...) }
func (oper *conditionOper) Where(sess *xorm.Session) *xorm.Session {
	oper.Build(sess.Engine)
	sess.Where(oper.CondiStr(), oper.Values()...)
	return sess
}
func (oper *conditionOper) Do(f func(sess *xorm.Session)) {
	sess := oper.engine.NewSession()
	defer sess.Close()
	sess.Where(oper.CondiStr(), oper.Values()...)
	f(sess)
}

func NewField(table *Table, fieldName string) *Field {
	return &Field{table: table, fieldName: fieldName}
}

type Field struct {
	table     *Table
	fieldName string
	name      string
}

func sliceFields(fs ...*Field) []*Field { return fs }

func (f *Field) Eq(v interface{}) Condition {
	return f.condition(v, "=")
}

func (f *Field) Ne(v interface{}) Condition {
	return f.condition(v, "<>")
}

func (f *Field) Gt(v interface{}) Condition {
	return f.condition(v, ">")
}

func (f *Field) Gte(v interface{}) Condition {
	return f.condition(v, ">=")
}

func (f *Field) Lt(v interface{}) Condition {
	return f.condition(v, "<")
}

func (f *Field) Lte(v interface{}) Condition {
	return f.condition(v, "<=")
}

func (f *Field) Between(vs ...interface{}) Condition {
	vs = paramsToSlice(vs)
	v1, v2 := vs[0], vs[1]
	return NewCondition(
		"(%s >= ? AND %s <= ?)",
		sliceFields(f, f),
		[]interface{}{v1, v2},
	)
}

func (f *Field) Outside(vs ...interface{}) Condition {
	vs = paramsToSlice(vs)
	v1, v2 := vs[0], vs[1]
	return NewCondition(
		"(%s < ? OR %s > ?)",
		sliceFields(f, f),
		[]interface{}{v1, v2},
	)
}

func (f *Field) Like(v string) Condition {
	return f.condition(v, "LIKE")
}

func (f *Field) Startswith(v string) Condition {
	return f.condition(fmt.Sprintf("%s%%", v), "LIKE")
}

func (f *Field) Endswith(v string) Condition {
	return f.condition(fmt.Sprintf("%%%s", v), "LIKE")
}

func (f *Field) Contains(v string) Condition {
	return f.condition(fmt.Sprintf("%%%s%%", v), "LIKE")
}

func (f *Field) In(vs ...interface{}) Condition {
	vs = paramsToSlice(vs...)
	total := len(vs)
	if total == 0 {
		return NewCondition("1 = 0", sliceFields(), vs)
	}
	cstrs := make([]string, total)
	for i := 0; i < total; i++ {
		cstrs[i] = "?"
	}
	return NewCondition("%s IN ("+strings.Join(cstrs, ", ")+")", sliceFields(f), vs)
}

func (f *Field) FindInSet(v interface{}) Condition {
	return NewCondition("FIND_IN_SET(?, %s)", sliceFields(f), []interface{}{v})
}

func (f *Field) Name(engine *xorm.Engine) string {
	if f.name == "" {
		names := []string{
			f.table.getTableName(engine),
			f.table.getColumenName(engine, f.fieldName),
		}
		for i, name := range names {
			names[i] = engine.Dialect().Quote(name)
		}
		f.name = strings.Join(names, ".")
	}
	return f.name
}

func (f *Field) condition(v interface{}, op string) (condi Condition) {
	switch v.(type) {
	case *Field:
		condi = NewCondition("%s "+op+" %s", sliceFields(f, v.(*Field)), []interface{}{})
	default:
		condi = NewCondition("%s "+op+" ?", sliceFields(f), []interface{}{v})
	}
	return condi
}

type Table struct {
	bean        interface{}
	engine      *xorm.Engine
	tableName   string
	columnNames map[string]string
	fields      map[string]*Field
}

func NewTable(bean interface{}) *Table {
	return &Table{bean: bean, fields: make(map[string]*Field)}
}

func (tb *Table) bindEngine(engine *xorm.Engine) {
	if tb.engine != nil {
		return
	}
	tb.engine = engine
	tableInfo := engine.TableInfo(tb.bean)
	tb.tableName = tableInfo.Name
	columns := tableInfo.Columns()
	tb.columnNames = make(map[string]string, len(columns))
	for _, col := range columns {
		tb.columnNames[col.FieldName] = col.Name
	}
}

func (tb *Table) GetTableName() string {
	return tb.tableName
}

func (tb *Table) getTableName(engine *xorm.Engine) string {
	tb.bindEngine(engine)
	return tb.tableName
}

func (tb *Table) getColumenName(engine *xorm.Engine, fieldName string) string {
	tb.bindEngine(engine)
	name, ok := tb.columnNames[fieldName]
	if !ok {
		panic(fmt.Sprintf(`Field "%s" is not exist in table "%s"`, fieldName, tb.tableName))
	}
	return name
}

func (tb *Table) Field() *Field {
	return tb.getField(funcName(2))
}

func (tb *Table) GetField(fieldName string) *Field {
	return tb.getField(fieldName)
}

func (tb *Table) getField(fieldName string) *Field {
	field, ok := tb.fields[fieldName]
	if !ok {
		field = NewField(tb, fieldName)
		tb.fields[fieldName] = field
	}
	return field
}
