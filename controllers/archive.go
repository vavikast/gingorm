package controllers

import (
	"net/http"
	"strconv"
	"math"
	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
	"gingorm/models"
	"gingorm/system"
)

func ArchiveGet(c *gin.Context) {
	var (
		year      string
		month     string
		page      string
		pageIndex int
		pageSize  = system.GetConfiguration().PageSize //获取configuration结构体中的pageSize
		total     int
		err       error
		posts     []*models.Post
		policy    *bluemonday.Policy //声明bluemonday中policy结构体
	)
	year = c.Param("year")
	month = c.Param("month")
	page = c.Query("page")
	pageIndex, _ = strconv.Atoi(page)
	if pageIndex <= 0 {
		pageIndex = 1
	}
	posts, err = models.ListPostByArchive(year, month, pageIndex, pageSize)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	total, err = models.CountPostByArchive(year, month)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	policy = bluemonday.StrictPolicy() //采用的是严格策略，剥离所有html元素和属性，适用于博客标题
	for _, post := range posts {
		post.Tags, _ = models.ListTagByPostId(strconv.FormatUint(uint64(post.ID), 10))
		post.Body = policy.Sanitize(string(blackfriday.Run([]byte(post.Body))))  //blackfriday在v2中更换为run
	}
	c.HTML(http.StatusOK, "index/index.html", gin.H{
		"posts":           posts,
		"tags":            models.MustListTag(),
		"archives":        models.MustListPostArchives(),
		"links":           models.MustListLinks(),
		"pageIndex":       pageIndex,
		"totalPage":       int(math.Ceil(float64(total) / float64(pageSize))),
		"maxReadPosts":    models.MustListMaxReadPost(),
		"maxCommentPosts": models.MustListMaxCommentPost(),
	})

}
