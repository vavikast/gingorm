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

func IndexGet(c *gin.Context) {
	var (
		pageIndex int
		pageSize  =  system.GetConfiguration().PageSize  //最终调用的是PageSize = DEFAULT_PAGESIZE=10
		total     int
		page      string
		err       error
		posts     []*models.Post
		policy    *bluemonday.Policy
	)
	//无法查询page，所以page=0
	page = c.Query("page")
	pageIndex, _ = strconv.Atoi(page)
	if pageIndex <= 0 {
		pageIndex = 1
	}
	//posts, err = models.ListPublishedPost("", 1, 10)
	//查询所有满足条件的post页面
	posts, err = models.ListPublishedPost("", pageIndex, pageSize)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	//查询所有发布页面标签数量
	total, err = models.CountPostByTag("")
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	//设置严格html策略
	policy = bluemonday.StrictPolicy()
	//标签和文章内容
	for _, post := range posts {
		post.Tags, _ = models.ListTagByPostId(strconv.FormatUint(uint64(post.ID), 10))
		post.Body = policy.Sanitize(string(blackfriday.Run([]byte(post.Body))))
	}
	user, _ := c.Get(CONTEXT_USER_KEY)
	c.HTML(http.StatusOK, "index/index.html", gin.H{
		"posts":           posts,
		"tags":            models.MustListTag(),
		"archives":        models.MustListPostArchives(),
		"links":           models.MustListLinks(),
		"user":            user,
		"pageIndex":       pageIndex,
		"totalPage":       int(math.Ceil(float64(total) / float64(pageSize))),
		"path":            c.Request.URL.Path,
		"maxReadPosts":    models.MustListMaxReadPost(),
		"maxCommentPosts": models.MustListMaxCommentPost(),
	})
}

func AdminIndex(c *gin.Context) {
	user, _ := c.Get(CONTEXT_USER_KEY)
	c.HTML(http.StatusOK, "admin/index.html", gin.H{
		"pageCount":    models.CountPage(),
		"postCount":    models.CountPost(),
		"tagCount":     models.CountTag(),
		"commentCount": models.CountComment(),
		"user":         user,
		"comments":     models.MustListUnreadComment(),
	})
}
