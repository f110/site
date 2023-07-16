package content

import (
	"bytes"
	"context"
	"time"

	"go.f110.dev/notion-api/v3"
	"go.f110.dev/notion-api/v3/markdown"
	"go.f110.dev/xerrors"
	"gopkg.in/yaml.v2"
)

type ArticleFile struct {
	Filename string
	Data     []byte
}

type Article struct {
	ID               string
	Title            string
	EnglishTitle     string
	Tags             []string
	ToC              bool
	SectionNumbering bool
	Freeze           bool
	Draft            bool
	Date             *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time

	Files []ArticleFile
}

func (a *Article) GetBody(client *notion.Client) ([]byte, error) {
	blocks, err := client.GetBlocks(context.Background(), a.ID)
	if err != nil {
		return nil, xerrors.WithStack(err)
	}

	buf, err := markdown.Render(blocks)
	if err != nil {
		return nil, xerrors.WithStack(err)
	}
	return buf, nil
}

type ArticleMetadata struct {
	Title            string    `yaml:"title"`
	Date             time.Time `yaml:"date"`
	LastMod          time.Time `yaml:"lastmod,omitempty"`
	IsCJKLanguage    bool      `yaml:"isCJKLanguage"`
	ToC              bool      `yaml:"toc,omitempty"`
	SectionNumbering bool      `yaml:"section_numbering,omitempty"`
	Draft            bool      `yaml:"draft,omitempty"`
	Tags             []string  `yaml:"tags,flow"`
}

func (a *Article) Render(body []byte) ([]byte, error) {
	date := a.Date
	if date == nil {
		date = &a.CreatedAt
	}

	meta := &ArticleMetadata{
		Title:            a.Title,
		Date:             *date,
		IsCJKLanguage:    true,
		ToC:              a.ToC,
		SectionNumbering: a.SectionNumbering,
		Tags:             a.Tags,
	}
	if !a.Freeze && a.UpdatedAt.Sub(a.CreatedAt) > 24*time.Hour {
		meta.LastMod = a.UpdatedAt
	}
	metaBuf, err := yaml.Marshal(meta)
	if err != nil {
		return nil, xerrors.WithStack(err)
	}

	buf := new(bytes.Buffer)
	buf.WriteString("---\n")
	buf.Write(metaBuf)
	buf.WriteString("---\n")
	buf.WriteRune('\n')
	buf.Write(body)

	return buf.Bytes(), nil
}

func GetArticles(client *notion.Client, id string) ([]*Article, error) {
	pages, err := client.GetPages(context.Background(), id, nil, []*notion.Sort{{Property: "Updated", Direction: "descending"}})
	if err != nil {
		return nil, xerrors.WithStack(err)
	}

	articles := make([]*Article, 0)
	for _, v := range pages {
		tags := collect(v.Properties["Tags"].MultiSelect, func(v *notion.Option) string { return v.Name })
		var date *time.Time
		if v.Properties["Date"].Date != nil {
			date = &v.Properties["Date"].Date.Start.Time
		}

		articles = append(articles, &Article{
			ID:               v.ID,
			Title:            v.Properties["Name"].Title[0].PlainText,
			EnglishTitle:     v.Properties["English Title"].RichText[0].PlainText,
			Tags:             tags,
			SectionNumbering: v.Properties["Section Numbering"].Checkbox,
			ToC:              v.Properties["ToC"].Checkbox,
			Freeze:           v.Properties["Freeze"].Checkbox,
			Date:             date,
			Draft:            v.Properties["Draft"].Checkbox,
			CreatedAt:        v.Properties["Created"].CreatedTime.Time,
			UpdatedAt:        v.Properties["Updated"].LastEditedTime.Time,
		})
	}

	return articles, nil
}

func collect[T, K any](in []T, f func(T) K) []K {
	r := make([]K, 0)
	for _, v := range in {
		r = append(r, f(v))
	}
	return r
}
