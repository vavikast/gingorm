package controllers

import (
	"fmt"

	"github.com/cihub/seelog"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	"gingorm/helpers"
	"gingorm/models"
	"gingorm/system"
)

func RssGet(c *gin.Context) {
	//获取当前时间
	now := helpers.GetCurrentTime()
	//初始化没有赋值，默认为“”
	domain := system.GetConfiguration().Domain
	feed := &feeds.Feed{
		Title:       "Wblog",
		Link:        &feeds.Link{Href: domain},
		Description: "Wblog,talk about golang,k8s and so on.",
		Author:      &feeds.Author{Name: "Felix", Email: "486892195@qq.com"},
		Created:     now,
	}

	feed.Items = make([]*feeds.Item, 0)
	posts, err := models.ListPublishedPost("", 0, 0)
	if err != nil {
		seelog.Error(err)
		return
	}

	for _, post := range posts {
		item := &feeds.Item{
			Id:          fmt.Sprintf("%s/post/%d", domain, post.ID),
			Title:       post.Title,
			Link:        &feeds.Link{Href: fmt.Sprintf("%s/post/%d", domain, post.ID)},
			Description: string(post.Excerpt()),
			Created:     now,
		}
		feed.Items = append(feed.Items, item)
	}
	rss, err := feed.ToRss()
	if err != nil {
		seelog.Error(err)
		return
	}
	c.Writer.WriteString(rss)
}
