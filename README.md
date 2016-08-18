# FilterXorm

[zh-CN](README_zh-cn.md)

FilterXorm is a tool for xorm to build filter conditions.

Like this `(col1 = v1 AND (col2 >= v2 OR col2 < v3) AND (col3 = v4 OR col3 = v5))`:

```go
filterxorm.And(
    col1.Eq(v1),
    filterxorm.Or(
        col2.Gte(v2),
        col2.Lt(v3),
    )
    filterxorm.Or(
        col3.Eq(v4),
        col3.Eq(v5),
    )
)
```

Of course flow style is supported `((col1 = v1 AND col2 = v2) AND col3 = v3)`:

```go
col1.Eq(v1).And(col2.Eq(v2)).And(col3.Eq(v3))
```

Or `(col1 = v1 AND col2 = v2 AND col3 = v3)`:

```go
col1.Eq(v1).And(col2.Eq(v2), col3.Eq(v3))
```

Or in different columns `col1 = col2`:

```go
col1.Eq(col2)
```

## Default function

- col.Eq(v): `col = v`
- col.Ne(v): `col <> v`
- col.Gt(v): `col > v`
- col.Gte(v): `col >= v`
- col.Lt(v): `col < v`
- col.Lte(v): `col <= v`
- col.In(ve): `col in (v)`
- col.Between(v1, v2) or col.Between([]T{v1, v2}): `col >= v1 AND col <= v2`
- col.Outside(v1, v2) or col.Outside([]T{v1, v2}): `col < v1 AND col > v2`
- string only: col.Like(s): `col LIKE s`
- string only: col.Startswith(s): `col LIKE s%`
- string only: col.Endswith(s): `col LIKE %s`
- string only: col.Contains(s): `col LIKE %s%`

## Usage

Model is:

```go
type Sample struct{
    Id   int    `xorm:"pk autoincr"`
    Name string `xorm:"user_name VARCHAR(128)"`
}
```

First, Get column Field

```go
tSample := filterxorm.NewTable(new(Sample))
fId := tSample.GetField("Id")
fName := tSample.GetField("Name")
```

Or get a table struct:

```go
type TSample struct{
    *filterxorm.Table
}

func NewSampleTable() *TSample {
    return &TSample{Table: filterxorm.NewTable(new(Sample))}
}
func (tb *TSample) Id() *filterxorm.Field   { return tb.Field() }
func (tb *TSample) Name() *filterxorm.Field { return tb.Field() }

tSample := NewSampleTable()
fId := tSample.Id()
fName := tSample.Name()
```

Then, build the condition:

```go
condi := filterxorm.Or(
    fId.Between(1, 10),
    fName.Eq("tom"),
).Build(engine)
```

Don't forget `Build(*xorm.Engine)`

After above step, we get ready to filter:

```go
sess := engine.NewSession()
defer sess.Close()

items := make([]Sample, 0)
err := sess.Where(condi.CondiStr(), condi.Values()...).
    Find(&items)
```

Or:

```go
sess := engine.NewSession()
defer sess.Close()

err := condi.Where(sess).Find(&items)
```

Or use new auto session (the session will close auto after query end):

```go
var err error
condi.Do(func(sess *xorm.Session) {
    err = sess.Find(&items)
})
```
