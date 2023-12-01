package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

/*
结构体
*/
type List struct {
	gorm.Model        // 主键
	Name       string `gorm:"type:varchar(20); not null" json:"name" binding:"required"`
	State      string `gorm:"type:varchar(20); not null" json:"state" binding:"required"`
	Phone      string `gorm:"type:varchar(20); not null" json:"phone" binding:"required"`
	Email      string `gorm:"type:varchar(20); not null" json:"email" binding:"required"`
	Address    string `gorm:"type:varchar(40); not null" json:"address" binding:"required"`
}

func main() {

	/*
		连接数据库
		如何连接数据库 ? MySQL + Navicat
		需要更改的内容：用户名，密码，数据库名称
	*/
	dsn := "root:123456@tcp(127.0.0.1:3306)/crud-list?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true},
	})

	if err != nil {
		fmt.Println("err", err.Error())
	}
	fmt.Println("db = ", db)
	fmt.Println("err = ", err)

	/*
		连接池
	*/
	sqlDB, err := db.DB()
	if err != nil {
		fmt.Println("err", err.Error())
	}
	/*SetMaxIdleConns 设置空闲连接池中连接的最大数量*/
	sqlDB.SetMaxIdleConns(10)
	/*SetMaxOpenConns 设置打开数据库连接的最大数量。*/
	sqlDB.SetMaxOpenConns(100)
	/* SetConnMaxLifetime 设置了连接可复用的最大时间。10秒钟*/
	sqlDB.SetConnMaxLifetime(10 * time.Second)

	/*
		迁移
	*/
	db.AutoMigrate(&List{})

	/*
		接口
	*/
	r := gin.Default()

	//测试
	r.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "请求成功",
		})
	})

	/*
		CRUD
			creat
	*/
	r.POST("/add", func(ctx *gin.Context) {
		var data List
		err := ctx.ShouldBind(&data)
		if err != nil {
			ctx.JSON(400, gin.H{
				"msg": "添加失败",
				"data": gin.H{
					"code": "400",
				},
			})
		} else {
			db.Create(&data)
			ctx.JSON(200, gin.H{
				"msg":  "添加成功",
				"date": data,
				"code": 200,
			})
		}
	})

	/*
		 curd
			delete
	*/
	r.DELETE("/delete/:id", func(ctx *gin.Context) {
		var data []List
		// 接收id
		id := ctx.Param("id") // 如果有键值对形式的话用Query()
		// 判断id是否存在
		db.Where("id = ? ", id).Find(&data)
		if len(data) == 0 {
			ctx.JSON(200, gin.H{
				"msg":  "id没有找到，删除失败",
				"code": 400,
			})
		} else {
			// 操作数据库删除（删除id所对应的那一条）
			// db.Where("id = ? ", id).Delete(&data) <- 其实不需要这样写，因为查到的data里面就是要删除的数据
			db.Delete(&data)

			ctx.JSON(200, gin.H{
				"msg":  "删除成功",
				"code": 200,
			})
		}

	})

	/*
		 curd
			update
	*/
	r.PUT("/update/:id", func(ctx *gin.Context) {
		// 1. 找到对应的id所对应的条目
		// 2. 判断id是否存在
		// 3. 修改对应条目 or 返回id没有找到
		var data List
		id := ctx.Param("id")
		db.Where("id=?", id).Find(&data)
		if data.ID == 0 {
			ctx.JSON(200, gin.H{
				"msg":  "没有找到id",
				"code": 400,
			})
		} else {
			err := ctx.ShouldBind(&data)
			if err != nil {
				ctx.JSON(200, gin.H{
					"msg":  "修改失败	",
					"code": 400,
				})
			} else {
				db.Where("id = ?", id).Updates(&data)
				ctx.JSON(200, gin.H{
					"msg":  "修改成功	",
					"code": 200,
				})
			}

		}

	})

	/*
		 curd
			read
	*/
	// 第一种：条件查询，
	r.POST("/read/:name", func(ctx *gin.Context) {
		var data []List
		name := ctx.Param("name")
		db.Where("name=?", name).Find(&data)
		if len(data) == 0 {
			ctx.JSON(http.StatusOK, gin.H{
				"msg":  "没有查询到数据" + name,
				"code": 400,
			})
		} else {
			ctx.JSON(http.StatusOK, gin.H{
				"msg":  "查询成功",
				"data": data,
			})
		}
	})
	// 第二种：全部查询 / 分页查询
	r.POST("/read/list", func(ctx *gin.Context) {
		var datalist []List

		// 查询全部数据 or 查询分页数据
		pageSize, _ := strconv.Atoi(ctx.Query("pageSize"))
		pageNum, _ := strconv.Atoi(ctx.Query("pageNum"))

		// 判断是否需要分页
		if pageSize == 0 {
			pageSize = -1
		}
		if pageNum == 0 {
			pageNum = -1
		}

		offsetVal := (pageNum - 1) * pageSize // 固定写法 记住就行
		if pageNum == -1 && pageSize == -1 {
			offsetVal = -1
		}

		// 返回一个总数
		var total int64

		// 查询数据库
		db.Model(&datalist).Count(&total).Limit(pageSize).Offset(offsetVal).Find(&datalist)
		if len(datalist) == 0 {
			ctx.JSON(200, gin.H{
				"msg":  "没有查询到数据",
				"code": 400,
				"data": gin.H{},
			})
		} else {
			ctx.JSON(200, gin.H{
				"msg":  "查询成功",
				"code": 200,
				"data": gin.H{
					"list":     datalist,
					"total":    total,
					"pageNum":  pageNum,
					"pageSize": pageSize,
				},
			})
		}
	})

	r.Run()

}
