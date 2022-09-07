package gorm2x

import (
	"encoding/json"
	"fmt"
	zero_logger "github.com/rs/zerolog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	gorm_logger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"testing"
	"time"
)

/**
gorm 2.x

GORM 倾向于约定，而不是配置，GORM 是"傻瓜式"的字符串拼接，结构体仅用来做字段映射

关联： 主要注意驱动表是哪个。关联方式的不同，驱动表就不同（相同的两张表，关联方式不同，驱动表就不同）

预加载：预加载的机制不同
Preload预加载 是 拆分多次查询 (先查主表，在in查关联子表数据)；
Joins预加载 是 一次性查询，默认用 inner join 一次性拖出来再拆分处理

注意哪些方法是准备SQL语句阶段，哪些方法是触发SQL语句执行

结构体 参数 存在 零值，因无法区分是传的零值，还是结构体默认的零值，存在二义性，统一忽略掉
要传值零值，只能使用map[string]interface{} 来实现
*/

func GetDB() *gorm.DB {

	// refer https://github.com/go-sql-driver/mysql#dsn-data-source-name for details
	dsn := "imake:12315@tcp(127.0.0.1:3306)/a1gorm?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,  // 使用单数表名，启用该选项后，`Student` 表将是`student`, 禁用则为`students`
			TablePrefix:   "tb_", // 表名前缀，`User`表为`pre_users`
			//NameReplacer:  strings.NewReplacer("AbcID", "Abcid"), // 在转为数据库名称之前，使用NameReplacer更改结构/字段名称。
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	db.Logger.LogMode(gorm_logger.LogLevel(zero_logger.InfoLevel))

	return db
}

type Time2 time.Time

type Model_ struct {
	ID        int   `gorm:"primaryKey;autoIncrement"`
	CreatedAt Time2 `gorm:"column:ctime;" json:"ctime"` //<= 手动设置字段名
	UpdatedAt Time2 `gorm:"column:utime;"`
}

type Model1 struct {
	Model_
	Name     string
	RealName string
}

type ModelCond struct {
	ID        int    `gorm:"primaryKey;autoIncrement"`
	CreatedAt Time2  `gorm:"column:ctime" json:"ctime"` //<= 手动设置字段名
	UpdatedAt Time2  `gorm:"column:utime" json:"utime"`
	C1        string `gorm:"c1"`
	C2        string `gorm:"c2"`
	C3        string `gorm:"c3"`
}

func TestGorm1(t *testing.T) {
	db := GetDB()

	var m Model1
	t.Run("t1", func(t *testing.T) {
		db.First(&m)
		b, _ := json.Marshal(&m)
		fmt.Println(string(b))

	})
	//t.Run("t2", func(t *testing.T) {
	//	// 通过传入外围的引用变量来获取数据，而不是通过返回值获取数据 ！！！ import
	//	txWrapper(db, func(tx *gorm.DB) error {
	//		tx.First(&m)
	//		return nil
	//	})
	//	fmt.Println(m.Name)
	//})

	type Result struct {
		Name  string `gorm:"name"`
		Bname string `gorm:"bname"`
	}

	var results []Result
	var result Result

	t.Run("t3", func(t *testing.T) {
		//db.Raw("select a.name ,b.name as bname from tb_model1 a left join tb_model2 b on a.id = b.id").Scan(&results)
		//db.Raw("select a.name ,b.name as bname from tb_model1 a left join tb_model2 b on a.id = b.id").Find(&results)
		_db := db.Table("tb_model1").Select("tb_model1.name,tb_model2.name bname").Joins("left join tb_model2 on tb_model1.id = tb_model2.id")
		var total int64
		_db.Count(&total)
		_db.Find(&results)
		println("xxx")
	})

	t.Run("t31", func(t *testing.T) {
		_db := db.Table("tb_model1 a").Select("a.name, b.name bname").Joins("left join tb_model2 b on a.id = b.id ").Order("a.id asc")
		var total int64
		_db.Count(&total)
		_db.Find(&results)
		println("xxx")
	})

	t.Run("t4", func(t *testing.T) {
		result := map[string]interface{}{}

		//db.Table("tb_model_cond").Find(&result)
		//
		//db.Debug().Model(&ModelCond{})
		//db.Where("c1 = ?", "c1")
		//db.Where("c2 = ?", "c2")
		//db.First(&result)

		db.Model(&ModelCond{}).Where("c1 = ?", "c1").Where("c2 = ?", "c2")
		db.First(&result)

		//_db := db.Debug().Table("tb_model_cond")
		//_db = _db.Where("c1 = ?", "c1")
		//_db = _db.Where("c2 = ?", "c2")
		//_db = _db.Where("c3 = ?", "c3")
		//var total int64
		//_db.Count(&total)
		//_db.Find(&result)
		println("xxx")
	})

	t.Run("sql", func(t *testing.T) {
		stat := db.Model(&ModelCond{}).Where("c1 = ?", "c1").Where("c2 = ?", "c2").First(&result).Statement
		t.Log(stat.SQL.String())
		t.Log(stat.Vars)
		// DryRun 模式, 在不执行的情况下生成 SQL 及其参数，可以用于准备或测试生成的 SQL
		stat = db.Session(&gorm.Session{DryRun: true}).First(&result, 1).Statement
		t.Log(stat.SQL.String())
		t.Log(stat.Vars)

		sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return tx.Find(&[]ModelCond{}, 1)
		})
		t.Log("ToSQL >>> ", sql)
	})

}
