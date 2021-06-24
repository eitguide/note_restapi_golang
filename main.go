package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type SQLModel struct {
	ID int `gorm:"column:id;primaryKey" json:"id" form:"id"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"create_at" form:"create_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoCreateTime" json:"updated_at" form:"updated_at"`
	Status int `gorm:"column:status;default:1" json:"status" form:"status"`
}

type Note struct {
	SQLModel
	Title string `gorm:"column:title;" json:"title" form:"title"`
}

func (Note)TableName() string  {
	return "notes"
}

func main()  {
	db, err := gorm.Open(sqlite.Open("rest_api.db"), &gorm.Config{})

	if err != nil {
		panic("fail to connect database")
	}

	err = db.AutoMigrate(&Note{})
	if err != nil {
		log.Println("Cant no migrate db", err.Error())
	}

	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.LoggerWithFormatter(func(params gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			params.ClientIP,
			params.TimeStamp.Format(time.RFC1123),
			params.Method,
			params.Path,
			params.Request.Proto,
			params.StatusCode,
			params.Latency,
			params.Request.UserAgent(),
			params.ErrorMessage,
		)
	}))

	router.MaxMultipartMemory = 8 << 20 //8 MiB

	responseNote := Note{
		Title: "Demo Title",
	}
	//Return Type in gin
	router.GET("json", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": responseNote})
	})

	router.GET("xml", func(c *gin.Context) {
		c.XML(http.StatusOK, gin.H{"data": responseNote})
	})

	router.GET("yaml", func(c *gin.Context) {
		c.YAML(http.StatusOK, gin.H{"data": responseNote})
	})

	router.GET("jsonsecure", func(c *gin.Context) {
		c.SecureJSON(http.StatusOK, gin.H{"data": responseNote})
	})

	router.GET("jsonp", func(c *gin.Context) {
		c.JSONP(http.StatusOK, gin.H{"data": responseNote})
	})

	router.GET("jsonascii", func(c *gin.Context) {
		c.AsciiJSON(http.StatusOK, gin.H{"data": responseNote})
	})


	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	router.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	router.Static("/assets", "./assets")

	v1 := router.Group("v1")
	{
		notes := v1.Group("notes")

		notes.GET("", func(c *gin.Context) {
			var results []Note

			if err := db.Find(&results).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"data": results})
		})

		notes.POST("", func(c *gin.Context) {
			var note Note
			if err := c.ShouldBind(&note); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			if err := db.Create(&note).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"data": note.ID})
		})

		notes.PUT("", func(c *gin.Context) {
			var note Note
			if err := c.ShouldBind(&note); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			if err := db.Model(&note).Where("id = ?", note.ID).Updates(&note).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"data": note.ID})
		})

		notes.GET(":id", func(c *gin.Context) {
			id := c.Param("id")
			var note Note

			if err := db.First(&note, id).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"data": note})
		})

		notes.DELETE(":id", func(c *gin.Context) {
			id := c.Param("id")
			var note Note

			if err := db.Delete(&note, id).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"data": "OK"})
		})
	}

	router.Run()

}
