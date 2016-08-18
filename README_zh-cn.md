# FilterXorm

FilterXorm 用来构建Xorm的过滤条件

如过滤条件为 `(col1 = v1 AND (col2 >= v2 OR col2 < v3) AND (col3 = v4 OR col3 = v5))`：

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

也支持链式 `((col1 = v1 AND col2 = v2) AND col3 = v3)`：

```go
col1.Eq(v1).And(col2.Eq(v2)).And(col3.Eq(v3))
```

或 `(col1 = v1 AND col2 = v2 AND col3 = v3)`:

```go
col1.Eq(v1).And(col2.Eq(v2), col3.Eq(v3))
```

或者在两个不同的字段之间进行比较 `col1 = col2`：

```go
col1.Eq(col2)
```

## 默认条件

- col.Eq(v): `col = v`
- col.Ne(v): `col <> v`
- col.Gt(v): `col > v`
- col.Gte(v): `col >= v`
- col.Lt(v): `col < v`
- col.Lte(v): `col <= v`
- col.In(ve): `col in (v)`
- col.Between(v1, v2) 或 col.Between([]T{v1, v2}): `col >= v1 AND col <= v2`
- col.Outside(v1, v2) 或 col.Outside([]T{v1, v2}): `col < v1 AND col > v2`
- string only: col.Like(s): `col LIKE s`
- string only: col.Startswith(s): `col LIKE s%`
- string only: col.Endswith(s): `col LIKE %s`
- string only: col.Contains(s): `col LIKE %s%`

## 示例

假设ORM对象的结构为：

```go
type Sample struct{
    Id   int    `xorm:"pk autoincr"`
    Name string `xorm:"user_name VARCHAR(128)"`
}
```

首先，获取字段：

```go
tSample := filterxorm.NewTable(new(Sample))
fId := tSample.GetField("Id")
fName := tSample.GetField("Name")
```

也可以专门建立一个结构体：

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

然后，构建条件：

```go
condi := filterxorm.Or(
    fId.Between(1, 10),
    fName.Eq("tom"),
).Build(engine)
```

不要忘记 `Build(*xorm.Engine)`，需要通过xorm.Engine来获取某些关键字。

完成以上几步，就可以使用到Xorm里了：

```go
sess := engine.NewSession()
defer sess.Close()

items := make([]Sample, 0)
err := sess.Where(condi.CondiStr(), condi.Values()...).
    Find(&items)
```

或者：

```go
sess := engine.NewSession()
defer sess.Close()

err := condi.Where(sess).Find(&items)
```

或者（这个会通过传入的engine自动产生一个xorm.Session并应用过滤条件，执行结束后会自动关闭）：

```go
var err error
condi.Do(func(sess *xorm.Session) {
    err = sess.Find(&items)
})
```
