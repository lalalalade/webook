package domain

type Article struct {
	Id      int64
	Title   string
	Content string
	// Author 要从用户来
	Author Author
	Status ArticlesStatus
}

type ArticlesStatus uint8

const (
	ArticleStatusUnknown ArticlesStatus = iota
	ArticleStatusUnPublished
	ArticleStatusPublished
	ArticleStatusPrivate
)

func (s ArticlesStatus) ToUint8() uint8 {
	return uint8(s)
}

func (s ArticlesStatus) Valid() bool {
	return s.ToUint8() > 0
}

func (s ArticlesStatus) NonPublished() bool {
	return s != ArticleStatusPublished
}

func (s ArticlesStatus) String() string {
	switch s {
	case ArticleStatusPrivate:
		return "private"
	case ArticleStatusUnPublished:
		return "unpublished"
	case ArticleStatusPublished:
		return "published"
	default:
		return "unknown"
	}
}

// ArticleStatusV1 适合状态复杂 业务操作多
type ArticleStatusV1 struct {
	Val  uint8
	Name string
}

var (
	ArticleStatusV1Unknown = ArticleStatusV1{Val: 0, Name: "unknown"}
)

// Author 在帖子领域内是一个值对象
type Author struct {
	Id   int64
	Name string
}
