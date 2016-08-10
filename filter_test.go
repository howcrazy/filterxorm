package filterxorm

import (
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
)

// Models

type Sample1 struct {
	Id       int       `xorm:"pk autoincr"`
	Name     string    `xorm:"user_name VARCHAR(128)"`
	CreateAt time.Time `xorm:"created"`
}

type Sample2 struct {
	Id       int       `xorm:"pk autoincr"`
	Value    float32   `xorm:""`
	CreateAt time.Time `xorm:"created"`
}

// Condition from table

func TestTable(t *testing.T) {
	var condi Condition

	table := NewTable(new(Sample1))
	fId := table.GetField("Id")
	fName := table.GetField("Name")
	fCreateAt := table.GetField("CreateAt")

	condi = fId.In([]int{1, 2, 3}).And(fName.Startswith("begin")).Build(engine)
	showCondi(condi)

	timeBegin, _ := time.Parse("2006-01-02 15:04:05", "2016-01-01 00:00:00")
	condi = And(
		fId.Between(1, 100),
		Or(
			fName.Startswith("Anney"),
			fName.Startswith("Tom"),
		),
		fCreateAt.Gte(timeBegin),
	).Build(engine)
	showCondi(condi)
}

// Condition from struct

type Sample1Table struct {
	*Table
}

var sample1Table *Sample1Table

func NewSample1Table() *Sample1Table {
	if sample1Table == nil {
		sample1Table = &Sample1Table{
			Table: NewTable(new(Sample1)),
		}
	}
	return sample1Table
}

func (tb *Sample1Table) Id() *Field       { return tb.Field() }
func (tb *Sample1Table) Name() *Field     { return tb.Field() }
func (tb *Sample1Table) CreateAt() *Field { return tb.Field() }

type Sample2Table struct {
	*Table
}

var sample2Table *Sample2Table

func NewSample2Table() *Sample2Table {
	if sample2Table == nil {
		sample2Table = &Sample2Table{
			Table: NewTable(new(Sample2)),
		}
	}
	return sample2Table
}

func (tb *Sample2Table) Id() *Field       { return tb.Field() }
func (tb *Sample2Table) Value() *Field    { return tb.Field() }
func (tb *Sample2Table) CreateAt() *Field { return tb.Field() }

func TestStruct(t *testing.T) {
	var condi Condition

	sample1 := NewSample1Table()
	sample2 := NewSample2Table()

	condi = sample1.Id().Eq(1).Build(engine)
	showCondi(condi)

	condi = And(
		sample1.Id().Eq(1),
		sample2.Id().Eq(2),
	).Build(engine)
	showCondi(condi)
}

// Usecase in xorm

func TestDB(t *testing.T) {
	var condi Condition

	sample1 := NewSample1Table()
	sample2 := NewSample2Table()

	condi = sample1.Name().Startswith("name").Build(engine)
	engine.Where(condi.CondiStr(), condi.Values()...).Get(new(Sample1))

	condiJoin := sample1.Id().Eq(sample2.Id()).Build(engine)
	condiOther := sample2.Value().Between(1.3, 5.8).Build(engine)
	engine.Where(condi.CondiStr(), condi.Values()).
		Join("LEFTJOIN", sample2.GetTableName(), condiJoin.CondiStr()).
		Where(condiOther.CondiStr(), condiOther.Values()...).
		Get(new(Sample1))
}

func showCondi(condi Condition) {
	debug("\"%s\" %s", condi.CondiStr(), condi.Values())
}

var engine *xorm.Engine

func init() {
	engine, _ = xorm.NewEngine("mysql", "root:123@/test?charset=utf8")
	engine.ShowSQL = true
}
