package models

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
	"html/template"
	"strconv"
	"time"
)

//db.Create(&user)时， “CreatedAt”用于存储记录的创建时间
//db.Save(&user) // "UpdatedAt"用于存储记录当前时间

//设置base_model基本表
type BaseModel struct {
	ID uint 	`gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
}


// table page 在数据库中创建pages表-页面表
type  Page struct {
	BaseModel
	Title string //title 标题
	Body string //body 主体
	View int //view count 观看次数
	IsPublished bool //是否发表

}



// table posts
type Post struct {
	BaseModel
	Title        string     // title
	Body         string     // body
	View         int        // view count
	IsPublished  bool       // published or not
	Tags         []*Tag     `gorm:"-"` // tags of post
	Comments     []*Comment `gorm:"-"` // comments of post
	CommentTotal int        `gorm:"-"` // count of comment
}

// table tags
type Tag struct {
	BaseModel
	Name  string // tag name
	Total int    `gorm:"-"` // count of post
}

// table post_tags
type PostTag struct {
	BaseModel
	PostId uint // post id
	TagId  uint // tag id
}

// table users
type User struct {
	gorm.Model
	Email         string    `gorm:"unique_index;default:null"` //邮箱
	Telephone     string    `gorm:"unique_index;default:null"` //手机号码
	Password      string    `gorm:"default:null"`              //密码
	VerifyState   string    `gorm:"default:'0'"`               //邮箱验证状态
	SecretKey     string    `gorm:"default:null"`              //密钥
	OutTime       time.Time `gorm:"default:null"` //过期时间 #此处更改 不更改解析出错
	GithubLoginId string    `gorm:"unique_index;default:null"` // github唯一标识
	GithubUrl     string    //github地址
	IsAdmin       bool      //是否是管理员
	AvatarUrl     string    // 头像链接
	NickName      string    // 昵称
	LockState     bool      `gorm:"default:'0'"` //锁定状态
}

// table comments
type Comment struct {
	BaseModel
	UserID    uint   // 用户id
	Content   string // 内容
	PostID    uint   // 文章id
	ReadState bool   `gorm:"default:'0'"` // 阅读状态
	//Replies []*Comment // 评论
	NickName  string `gorm:"-"`
	AvatarUrl string `gorm:"-"`
	GithubUrl string `gorm:"-"`
}

// table subscribe
type Subscriber struct {
	gorm.Model
	Email          string    `gorm:"unique_index"` //邮箱
	VerifyState    bool      `gorm:"default:'0'"`  //验证状态
	SubscribeState bool      `gorm:"default:'1'"`  //订阅状态
	OutTime        time.Time //过期时间
	SecretKey      string    // 秘钥
	Signature      string    //签名
}

// table link
type Link struct {
	gorm.Model
	Name string //名称
	Url  string //地址
	Sort int    `gorm:"default:'0'"` //排序
	View int    //访问次数
}

// query result
type QrArchive struct {
	ArchiveDate time.Time //month
	Total       int       //total
	Year        int       // year
	Month       int       // month
}

type SmmsFile struct {
	BaseModel
	FileName  string `json:"filename"`
	StoreName string `json:"storename"`
	Size      int    `json:"size"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Hash      string `json:"hash"`
	Delete    string `json:"delete"`
	Url       string `json:"url"`
	Path      string `json:"path"`
}

var DB *gorm.DB   //做了一个全局的的DB, initDB函数中把db赋值给DB，同时initDB中的return的db在main函数中被defer db.close了，函数没有被关闭之前全局DB继承了db的属性，所以可以执行下面的函数。

func InitDB() (*gorm.DB, error) {

	db, err := gorm.Open("mysql", "root:Itcen2531,.@(127.0.0.1:3306)/wblog?charset=utf8mb4&parseTime=True&loc=Local")
	if err == nil {
		DB = db
		//db.LogMode(true)
		db.AutoMigrate(&Page{}, &Post{}, &Tag{}, &PostTag{}, &User{}, &Comment{}, &Subscriber{}, &Link{}, &SmmsFile{})
		db.Model(&PostTag{}).AddUniqueIndex("uk_post_tag", "post_id", "tag_id")
		return db, err
	}
	return nil, err
}

// Page  插入页面
func (page *Page) Insert() error {
	return DB.Create(page).Error
}

//更新页面，但是是全局更新，没想通这样操作的意义，等后面的操作看完，再补充
func (page *Page) Update() error {
	return DB.Model(page).Updates(map[string]interface{}{
		"title":        page.Title,
		"body":         page.Body,
		"is_published": page.IsPublished,
	}).Error
}
//更新浏览量，但是是全局更新，没想通这样操作的意义，等后面的操作看完，再补充
func (page *Page) UpdateView() error {
	return DB.Model(page).Updates(map[string]interface{}{
		"view": page.View,
	}).Error
}

//删除page数据，但是是全局更新
func (page *Page) Delete() error {
	return DB.Delete(page).Error
}

//获取指定行的页面信息
func GetPageById(id string) (*Page, error) {
	pid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, err
	}
	var page Page
	err = DB.First(&page, "id = ?", pid).Error
	return &page, err
}

//获取所有公开页面
func ListPublishedPage() ([]*Page, error) {
	return _listPage(true)
}

//获取所有页面
func ListAllPage() ([]*Page, error) {
	return _listPage(false)
}


//[]*page 其实可以看成gorm中的pages的实现，gorm的官方文档只说明了，Find(&pages),下面就说了，怎么实现这个pages
//获取所有页面和获取所有公开页面
func _listPage(published bool) ([]*Page, error) {
	var pages []*Page
	var err error
	if published {
		err = DB.Where("is_published = ?", true).Find(&pages).Error
	} else {
		err = DB.Find(&pages).Error
	}
	return pages, err
}


//获取数量页面数量
func CountPage() int {
	var count int
	DB.Model(&Page{}).Count(&count)
	return count
}

// Post 发布页面 发布页面插入
func (post *Post) Insert() error {
	return DB.Create(post).Error
}

//更新发布页面，但是是更新所有页面，还没有想通这样操作的意义
func (post *Post) Update() error {
	return DB.Model(post).Updates(map[string]interface{}{
		"title":        post.Title,
		"body":         post.Body,
		"is_published": post.IsPublished,
	}).Error
}

//更新发布页面的次数
func (post *Post) UpdateView() error {
	return DB.Model(post).Updates(map[string]interface{}{
		"view": post.View,
	}).Error
}

//删除发布页面
func (post *Post) Delete() error {
	return DB.Delete(post).Error
}


//摘要，估计是主页显示使用，还需要进一步确定
//blackfriday中的MarkdownBasic方法已经换成了run方法，此处使用了严格的html策略


func (post *Post) Excerpt() template.HTML {
	//you can sanitize, cut it down, add images, etc
	policy := bluemonday.StrictPolicy() //remove all html tags
	sanitized := policy.Sanitize(string(blackfriday.Run([]byte(post.Body))))
	runes := []rune(sanitized)
	if len(runes) > 300 {
		sanitized = string(runes[:300])
	}
	excerpt := template.HTML(sanitized + "...")
	return excerpt
}


//列举发布页面
func _listPost(tag string, published bool, pageIndex, pageSize int) ([]*Post, error) {
	var posts []*Post
	var err error
	if len(tag) > 0 {
		tagId, err := strconv.ParseUint(tag, 10, 64)
		if err != nil {
			return nil, err
		}
		//调用gorm的原生函数。定义了rows。
		var rows *sql.Rows
		if published {
			if pageIndex > 0 {
				//使用了内聚函数。 定了posts的别名 p posts_tag的别名pt ，使用了内连接。查询当posts中的id与tag中的id相当的，当post-tag中id等于，p.is_published满足条件的posts，
				rows, err = DB.Raw("select p.* from posts p inner join post_tags pt on p.id = pt.post_id where pt.tag_id = ? and p.is_published = ? order by created_at desc limit ? offset ?", tagId, true, pageSize, (pageIndex-1)*pageSize).Rows()
			} else {
				rows, err = DB.Raw("select p.* from posts p inner join post_tags pt on p.id = pt.post_id where pt.tag_id = ? and p.is_published = ? order by created_at desc", tagId, true).Rows()
			}
		} else {
			rows, err = DB.Raw("select p.* from posts p inner join post_tags pt on p.id = pt.post_id where pt.tag_id = ? order by created_at desc", tagId).Rows()
		}
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var post Post
			DB.ScanRows(rows, &post)
			posts = append(posts, &post)
		}
	} else {
		if published {
			if pageIndex > 0 {
				err = DB.Where("is_published = ?", true).Order("created_at desc").Limit(pageSize).Offset((pageIndex - 1) * pageSize).Find(&posts).Error
			} else {
				err = DB.Where("is_published = ?", true).Order("created_at desc").Find(&posts).Error
			}
		} else {
			err = DB.Order("created_at desc").Find(&posts).Error
		}
	}
	return posts, err
}

//显示所有公开发布
func ListPublishedPost(tag string, pageIndex, pageSize int) ([]*Post, error) {
	return _listPost(tag, true, pageIndex, pageSize)
}

//显示所有发布
func ListAllPost(tag string) ([]*Post, error) {
	return _listPost(tag, false, 0, 0)
}



//显示可公开最多阅读数
func ListMaxReadPost() (posts []*Post, err error) {
	err = DB.Where("is_published = ?", true).Order("view desc").Limit(5).Find(&posts).Error
	return
}

//必须显示可公开最多阅读数
func MustListMaxReadPost() (posts []*Post) {
	posts, _ = ListMaxReadPost()
	return
}


//显示最大评论数
func ListMaxCommentPost() (posts []*Post, err error) {
	var (
		rows *sql.Rows
	)
	rows, err = DB.Raw("select p.*,c.total comment_total from posts p inner join (select post_id,count(*) total from comments group by post_id) c on p.id = c.post_id order by c.total desc limit 5").Rows()
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var post Post
		DB.ScanRows(rows, &post)
		posts = append(posts, &post)
	}
	return
}


//显示必须最大连接数
func MustListMaxCommentPost() (posts []*Post) {
	posts, _ = ListMaxCommentPost()
	return
}


//显示发布标签数量
func CountPostByTag(tag string) (count int, err error) {
	var (
		tagId uint64
	)
	if len(tag) > 0 {
		tagId, err = strconv.ParseUint(tag, 10, 64)
		if err != nil {
			return
		}
		err = DB.Raw("select count(*) from posts p inner join post_tags pt on p.id = pt.post_id where pt.tag_id = ? and p.is_published = ?", tagId, true).Row().Scan(&count)
	} else {
		err = DB.Raw("select count(*) from posts p where p.is_published = ?", true).Row().Scan(&count)
	}
	return
}


//显示发布数量
func CountPost() int {
	var count int
	DB.Model(&Post{}).Count(&count)
	return count
}

//显示第n个发布的信息
func GetPostById(id string) (*Post, error) {
	pid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, err
	}
	var post Post
	err = DB.First(&post, "id = ?", pid).Error
	return &post, err
}

//归档查询
func ListPostArchives() ([]*QrArchive, error) {
	var archives []*QrArchive
	querysql := `select DATE_FORMAT(created_at,'%Y-%m') as month,count(*) as total from posts where is_published = ? group by month order by month desc`
	//querysql := `select strftime('%Y-%m',created_at) as month,count(*) as total from posts where is_published = ? group by month order by month desc`
	rows, err := DB.Raw(querysql, true).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var archive QrArchive
		var month string
		rows.Scan(&month, &archive.Total)
		//DB.ScanRows(rows, &archive)
		archive.ArchiveDate, _ = time.Parse("2006-01", month)
		archive.Year = archive.ArchiveDate.Year()
		archive.Month = int(archive.ArchiveDate.Month())
		archives = append(archives, &archive)
	}
	return archives, nil
}

//归档查询所有信息
func MustListPostArchives() []*QrArchive {
	archives, _ := ListPostArchives()
	return archives
}


//文章查询
func ListPostByArchive(year, month string, pageIndex, pageSize int) ([]*Post, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if len(month) == 1 {
		month = "0" + month
	}
	condition := fmt.Sprintf("%s-%s", year, month)
	if pageIndex > 0 {
		querysql := `select * from posts where date_format(created_at,'%Y-%m') = ? and is_published = ? order by created_at desc limit ? offset ?`
		//querysql := `select * from posts where strftime('%Y-%m',created_at) = ? and is_published = ? order by created_at desc limit ? offset ?`
		rows, err = DB.Raw(querysql, condition, true, pageSize, (pageIndex-1)*pageSize).Rows()
	} else {
		querysql := `select * from posts where date_format(created_at,'%Y-%m') = ? and is_published = ? order by created_at desc`
		//querysql := `select * from posts where strftime('%Y-%m',created_at) = ? and is_published = ? order by created_at desc`
		rows, err = DB.Raw(querysql, condition, true).Rows()
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	posts := make([]*Post, 0)
	for rows.Next() {
		var post Post
		DB.ScanRows(rows, &post)
		posts = append(posts, &post)
	}
	return posts, nil
}


//发布归档数量
func CountPostByArchive(year, month string) (count int, err error) {
	if len(month) == 1 {
		month = "0" + month
	}
	condition := fmt.Sprintf("%s-%s", year, month)
	//querysql := `select count(*) from posts where date_format(created_at,'%Y-%m') = ? and is_published = ? order by created_at desc`
	querysql := `select count(*) from posts where strftime('%Y-%m',created_at) = ? and is_published = ?`
	err = DB.Raw(querysql, condition, true).Row().Scan(&count)
	return
}

// Tag  插入标签
func (tag *Tag) Insert() error {
	return DB.FirstOrCreate(tag, "name = ?", tag.Name).Error
}


//列出标签和错误
func ListTag() ([]*Tag, error) {
	var tags []*Tag
	rows, err := DB.Raw("select t.*,count(*) total from tags t inner join post_tags pt on t.id = pt.tag_id inner join posts p on pt.post_id = p.id where p.is_published = ? group by pt.tag_id", true).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var tag Tag
		DB.ScanRows(rows, &tag)
		tags = append(tags, &tag)
	}
	return tags, nil
}

//列出标签
func MustListTag() []*Tag {
	tags, _ := ListTag()
	return tags
}


//通过发布的序号，列出所有标签
func ListTagByPostId(id string) ([]*Tag, error) {
	var tags []*Tag
	pid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, err
	}
	rows, err := DB.Raw("select t.* from tags t inner join post_tags pt on t.id = pt.tag_id where pt.post_id = ?", uint(pid)).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var tag Tag
		DB.ScanRows(rows, &tag)
		tags = append(tags, &tag)
	}
	return tags, nil
}

//标签数量
func CountTag() int {
	var count int
	DB.Model(&Tag{}).Count(&count)
	return count
}

//列出所有标签
func ListAllTag() ([]*Tag, error) {
	var tags []*Tag
	err := DB.Model(&Tag{}).Find(&tags).Error
	return tags, err
}

// post_tags
//发布上插入的标签
func (pt *PostTag) Insert() error {
	return DB.FirstOrCreate(pt, "post_id = ? and tag_id = ?", pt.PostId, pt.TagId).Error
}

//删除发布的标签
func DeletePostTagByPostId(postId uint) error {
	return DB.Delete(&PostTag{}, "post_id = ?", postId).Error
}

// user
// insert user 插入用户
func (user *User) Insert() error {
	return DB.Create(user).Error
}

// update user 更新用户信息 为什么更新所有用户信息，#存疑
func (user *User) Update() error {
	return DB.Save(user).Error
}

//查询用户的信息 内联条件查询 SELECT * FROM users WHERE email = 'username' LIMIT 1
func GetUserByUsername(username string) (*User, error) {
	var user User
	err := DB.First(&user, "email = ?", username).Error
	return &user, err
}
// 根据条件获取第一条记录，否则根据跟定条件创建一个新的记录
func (user *User) FirstOrCreate() (*User, error) {
	err := DB.FirstOrCreate(user, "github_login_id = ?", user.GithubLoginId).Error
	return user, err
}
//根据githubid获取一个用户信息
func IsGithubIdExists(githubId string, id uint) (*User, error) {
	var user User
	err := DB.First(&user, "github_login_id = ? and id != ?", githubId, id).Error
	return &user, err
}

//根据主键id获取一个用户信息
func GetUser(id interface{}) (*User, error) {
	var user User
	err := DB.First(&user, id).Error
	return &user, err
}

//更新用户信息
func (user *User) UpdateProfile(avatarUrl, nickName string) error {
	return DB.Model(user).Update(User{AvatarUrl: avatarUrl, NickName: nickName}).Error
}

//更新用户邮箱信息
func (user *User) UpdateEmail(email string) error {
	if len(email) > 0 {
		return DB.Model(user).Update("email", email).Error
	} else {
		return DB.Model(user).Update("email", gorm.Expr("NULL")).Error
	}
}

//更新用户信息 #存疑
func (user *User) UpdateGithubUserInfo() error {
	var githubLoginId interface{}
	if len(user.GithubLoginId) == 0 {
		githubLoginId = gorm.Expr("NULL")
	} else {
		githubLoginId = user.GithubLoginId
	}
	return DB.Model(user).Update(map[string]interface{}{
		"github_login_id": githubLoginId,
		"avatar_url":      user.AvatarUrl,
		"github_url":      user.GithubUrl,
	}).Error
}


//设置用户锁定状态
func (user *User) Lock() error {
	return DB.Model(user).Update(map[string]interface{}{
		"lock_state": user.LockState,
	}).Error
}

//列出所有用户
func ListUsers() ([]*User, error) {
	var users []*User
	err := DB.Find(&users, "is_admin = ?", false).Error
	return users, err
}

// Comment 插入评论
func (comment *Comment) Insert() error {
	return DB.Create(comment).Error
}

//更新评论
func (comment *Comment) Update() error {
	return DB.Model(comment).UpdateColumn("read_state", true).Error
}

//设置所有的评论状态可读
func SetAllCommentRead() error {
	return DB.Model(&Comment{}).Where("read_state = ?", false).Update("read_state", true).Error
}

//列入所有未读评论
func ListUnreadComment() ([]*Comment, error) {
	var comments []*Comment
	err := DB.Where("read_state = ?", false).Order("created_at desc").Find(&comments).Error
	return comments, err
}
//列出所有必须未读
func MustListUnreadComment() []*Comment {
	comments, _ := ListUnreadComment()
	return comments
}

//删除评论
func (comment *Comment) Delete() error {
	return DB.Delete(comment, "user_id = ?", comment.UserID).Error
}

//根据发布列出评论
func ListCommentByPostID(postId string) ([]*Comment, error) {
	pid, err := strconv.ParseUint(postId, 10, 64)
	if err != nil {
		return nil, err
	}
	var comments []*Comment
	rows, err := DB.Raw("select c.*,u.github_login_id nick_name,u.avatar_url,u.github_url from comments c inner join users u on c.user_id = u.id where c.post_id = ? order by created_at desc", uint(pid)).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var comment Comment
		DB.ScanRows(rows, &comment)
		comments = append(comments, &comment)
	}
	return comments, err
}

/*func GetComment(id interface{}) (*Comment, error) {
	var comment Comment
	err := DB.First(&comment, id).Error
	return &comment, err
}*/
//列出评论数量
func CountComment() int {
	var count int
	DB.Model(&Comment{}).Count(&count)
	return count
}

// Subscriber 插入订阅用户
func (s *Subscriber) Insert() error {
	return DB.FirstOrCreate(s, "email = ?", s.Email).Error
}

//更新订阅用户信息
func (s *Subscriber) Update() error {
	return DB.Model(s).Update(map[string]interface{}{
		"verify_state":    s.VerifyState,
		"subscribe_state": s.SubscribeState,
		"out_time":        s.OutTime,
		"signature":       s.Signature,
		"secret_key":      s.SecretKey,
	}).Error
}

//列出所有有效订阅者
func ListSubscriber(invalid bool) ([]*Subscriber, error) {
	var subscribers []*Subscriber
	db := DB.Model(&Subscriber{})
	if invalid {
		db.Where("verify_state = ? and subscribe_state = ?", true, true)
	}
	err := db.Find(&subscribers).Error
	return subscribers, err
}

//列出有效订阅者数量
func CountSubscriber() (int, error) {
	var count int
	err := DB.Model(&Subscriber{}).Where("verify_state = ? and subscribe_state = ?", true, true).Count(&count).Error
	return count, err
}

//根据邮箱返回订阅者
func GetSubscriberByEmail(mail string) (*Subscriber, error) {
	var subscriber Subscriber
	err := DB.Find(&subscriber, "email = ?", mail).Error
	return &subscriber, err
}

//根据签名返回订阅者
func GetSubscriberBySignature(key string) (*Subscriber, error) {
	var subscriber Subscriber
	err := DB.Find(&subscriber, "signature = ?", key).Error
	return &subscriber, err
}

//根据id返回订阅者
func GetSubscriberById(id uint) (*Subscriber, error) {
	var subscriber Subscriber
	err := DB.First(&subscriber, id).Error
	return &subscriber, err
}

// Link 插入友情链接
func (link *Link) Insert() error {
	return DB.FirstOrCreate(link, "url = ?", link.Url).Error
}

//更新信息 #所有信息
func (link *Link) Update() error {
	return DB.Save(link).Error
}

//删除友情链接
func (link *Link) Delete() error {
	return DB.Delete(link).Error
}

//列入所有链接
func ListLinks() ([]*Link, error) {
	var links []*Link
	err := DB.Order("sort asc").Find(&links).Error
	return links, err
}

//列入所有链接
func MustListLinks() []*Link {
	links, _ := ListLinks()
	return links
}

//根据id列出链接
func GetLinkById(id uint) (*Link, error) {
	var link Link
	err := DB.FirstOrCreate(&link, "id = ?", id).Error
	return &link, err
}

/*func GetLinkByUrl(url string) (*Link, error) {
	var link Link
	err := DB.Find(&link, "url = ?", url).Error
	return &link, err
}*/

//插入smmsfile
func (sf SmmsFile) Insert() (err error) {
	err = DB.Create(&sf).Error
	return
}

