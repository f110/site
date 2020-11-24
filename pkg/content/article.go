package content

import (
	"bufio"
	"bytes"
	"strings"
	"time"

	"github.com/kjk/notionapi"
	"github.com/kjk/notionapi/tomarkdown"
	"golang.org/x/xerrors"
	"gopkg.in/yaml.v2"
)

type Article struct {
	ID               string
	Title            string
	EnglishTitle     string
	Tags             []string
	ToC              bool
	SectionNumbering bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (a *Article) GetBody(client *notionapi.Client) ([]byte, error) {
	p, err := client.DownloadPage(a.ID)
	if err != nil {
		return nil, xerrors.Errorf(": %w", err)
	}
	buf := tomarkdown.ToMarkdown(p)
	s := bufio.NewScanner(bytes.NewReader(buf))
	body := new(bytes.Buffer)
	for i := 1; s.Scan(); i++ {
		if i > 2 {
			body.WriteString(s.Text())
			body.WriteRune('\n')
		}
	}

	return body.Bytes(), nil
}

type ArticleMetadata struct {
	Title            string    `yaml:"title"`
	Date             time.Time `yaml:"date"`
	IsCJKLanguage    bool      `yaml:"isCJKLanguage"`
	ToC              bool      `yaml:"toc,omitempty"`
	SectionNumbering bool      `yaml:"section_numbering,omitempty"`
	Tags             []string  `yaml:"tags,flow"`
}

func (a *Article) Render(body []byte) ([]byte, error) {
	meta := &ArticleMetadata{
		Title:            a.Title,
		Date:             a.CreatedAt,
		IsCJKLanguage:    true,
		ToC:              a.ToC,
		SectionNumbering: a.SectionNumbering,
		Tags:             a.Tags,
	}
	metaBuf, err := yaml.Marshal(meta)
	if err != nil {
		return nil, xerrors.Errorf(": %w", err)
	}

	buf := new(bytes.Buffer)
	buf.WriteString("---\n")
	buf.Write(metaBuf)
	buf.WriteString("---\n")
	buf.WriteRune('\n')
	buf.Write(body)

	return buf.Bytes(), nil
}

type Database struct {
	Properties   map[string]*Property
	IdToProperty map[string]*Property
}

type Property struct {
	ID   string
	Name string
	Type string
}

func GetArticles(client *notionapi.Client, id string) ([]*Article, error) {
	page, err := client.DownloadPage(id)
	if err != nil {
		return nil, xerrors.Errorf(": %w", err)
	}
	db := newDatabase(page)

	collection, err := client.QueryCollection(
		page.CollectionRecords[0].ID,
		page.CollectionViewRecords[0].ID,
		&notionapi.Query{
			Sort: []*notionapi.QuerySort{
				{Property: db.Properties["Created"].ID, Direction: "descending"},
			},
		},
		nil,
	)
	if err != nil {
		return nil, xerrors.Errorf(": %w", err)
	}

	articles := make([]*Article, 0)
	for _, v := range collection.Result.BlockIDS {
		r := collection.RecordMap.Blocks[v]
		_, toc := r.Block.Properties[db.Properties["ToC"].ID]
		_, sectionNumbering := r.Block.Properties[db.Properties["Section Numbering"].ID]
		title := r.Block.Properties["title"].([]interface{})[0].([]interface{})[0].(string)
		engTitle := r.Block.Properties[db.Properties["English Title"].ID].([]interface{})[0].([]interface{})[0].(string)
		tag := r.Block.Properties[db.Properties["Tags"].ID].([]interface{})[0].([]interface{})[0].(string)
		tags := strings.Split(tag, ",")

		articles = append(articles, &Article{
			ID:               v,
			Title:            title,
			EnglishTitle:     engTitle,
			Tags:             tags,
			SectionNumbering: sectionNumbering,
			ToC:              toc,
			CreatedAt:        r.Block.CreatedOn(),
			UpdatedAt:        r.Block.LastEditedOn(),
		})
	}

	return articles, nil
}

func newDatabase(page *notionapi.Page) *Database {
	col := page.CollectionByID(page.CollectionRecords[0].ID)
	prop := make(map[string]*Property)
	idToProp := make(map[string]*Property)
	for id, v := range col.Schema {
		prop[v.Name] = &Property{ID: id, Name: v.Name, Type: v.Type}
		idToProp[id] = prop[v.Name]
	}

	return &Database{Properties: prop, IdToProperty: idToProp}
}
