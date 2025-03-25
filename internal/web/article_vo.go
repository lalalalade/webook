package web

import "github.com/lalalalade/webook/internal/domain"

// ArticleVO 给前端看的
type ArticleVO struct {
	Id       int64
	Title    string
	Abstract string
	Content  string
	Author   string
	Status   uint8
	Ctime    string
	Utime    string
}

type ArticleReq struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type ListReq struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

func (req ArticleReq) toDomain(uid int64) domain.Article {
	return domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uid,
		},
	}
}
