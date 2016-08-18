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
	sess := engine.NewSession()
	defer sess.Close()

	sample1 := new(Sample1)
	tSample1 := NewTable(sample1)
	fId := tSample1.GetField("Id")
	fName := tSample1.GetField("Name")

	items := make([]Sample1, 0)
	condi = fId.In([]int{1, 2}).And(fName.Endswith("y")).Build(engine)
	if err := condi.Where(sess).Find(&items); err != nil {
		t.Error(err.Error())
	}
	debug(items)

	condi = Or(
		fId.Between(1, 2),
		fName.Eq("Kenny"),
	).Build(engine)
	if _, err := condi.Where(sess).Get(sample1); err != nil {
		t.Error(err.Error())
	}
	debug(sample1)
}

// Condition from struct

type Sample1Table struct {
	*Table
}

func NewSample1Table() *Sample1Table {
	return &Sample1Table{
		Table: NewTable(new(Sample1)),
	}
}

func (tb *Sample1Table) Id() *Field       { return tb.Field() }
func (tb *Sample1Table) Name() *Field     { return tb.Field() }
func (tb *Sample1Table) CreateAt() *Field { return tb.Field() }

type Sample2Table struct {
	*Table
}

func NewSample2Table() *Sample2Table {
	return &Sample2Table{
		Table: NewTable(new(Sample2)),
	}
}

func (tb *Sample2Table) Id() *Field       { return tb.Field() }
func (tb *Sample2Table) Value() *Field    { return tb.Field() }
func (tb *Sample2Table) CreateAt() *Field { return tb.Field() }

func TestStruct(t *testing.T) {
	var condi Condition
	var err error

	sample1 := new(Sample1)
	tSample1 := NewSample1Table()

	items := make([]Sample1, 0)
	condi = tSample1.Id().Eq(1).Build(engine)
	condi.Do(func(sess *xorm.Session) {
		err = sess.Find(&items)
	})
	if err != nil {
		t.Error(err.Error())
	}
	debug(items)

	condi = Or(
		tSample1.Id().Between(1, 2),
		tSample1.Name().Eq("Kenny"),
	).Build(engine)
	condi.Do(func(sess *xorm.Session) {
		_, err = sess.Get(sample1)
	})
	if err != nil {
		t.Error(err.Error())
	}
	debug(sample1)
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
var models []interface{}

func init() {
	engine, _ = xorm.NewEngine("mysql", "root:@/test?charset=utf8")
	engine.ShowSQL = true
	models = append(models, new(Sample1), new(Sample2))
	if err := engine.Sync2(models...); err != nil {
		panic(err.Error())
		return
	}
	items := []interface{}{
		Sample1{Name: "Kenny", CreateAt: time.Date(2016, 1, 1, 0, 0, 0, 0, time.Local)},
		Sample1{Name: "Sindy", CreateAt: time.Date(2016, 2, 1, 0, 0, 0, 0, time.Local)},
		Sample1{Name: "Sam", CreateAt: time.Date(2016, 3, 1, 0, 0, 0, 0, time.Local)},
		Sample2{Value: 1, CreateAt: time.Date(2016, 4, 1, 0, 0, 0, 0, time.Local)},
		Sample2{Value: 2, CreateAt: time.Date(2016, 5, 1, 0, 0, 0, 0, time.Local)},
		Sample2{Value: 3, CreateAt: time.Date(2016, 6, 1, 0, 0, 0, 0, time.Local)},
	}
	if _, err := engine.NoAutoTime().Insert(items...); err != nil {
		panic(err.Error())
		return
	}
}

func TestFinish(t *testing.T) {
	if err := engine.DropTables(models...); err != nil {
		t.Error(err.Error())
	}
}
