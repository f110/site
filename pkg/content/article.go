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
	Freeze           bool
	Date             *time.Time
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
	LastMod          time.Time `yaml:"lastmod,omitempty"`
	IsCJKLanguage    bool      `yaml:"isCJKLanguage"`
	ToC              bool      `yaml:"toc,omitempty"`
	SectionNumbering bool      `yaml:"section_numbering,omitempty"`
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
		date := datePropertyValue(r.Block.Properties, db.Properties["Date"].ID)
		toc := checkboxPropertyValue(r.Block.Properties, db.Properties["ToC"].ID)
		sectionNumbering := checkboxPropertyValue(r.Block.Properties, db.Properties["Section Numbering"].ID)
		freeze := checkboxPropertyValue(r.Block.Properties, db.Properties["Freeze"].ID)
		title := textPropertyValue(r.Block.Properties, "title")
		engTitle := textPropertyValue(r.Block.Properties, db.Properties["English Title"].ID)
		tags := multiSelectPropertyValue(r.Block.Properties, db.Properties["Tags"].ID)

		articles = append(articles, &Article{
			ID:               v,
			Title:            title,
			EnglishTitle:     engTitle,
			Tags:             tags,
			SectionNumbering: sectionNumbering,
			ToC:              toc,
			Freeze:           freeze,
			Date:             date,
			CreatedAt:        r.Block.CreatedOn(),
			UpdatedAt:        r.Block.LastEditedOn(),
		})
	}

	return articles, nil
}

func datePropertyValue(properties map[string]interface{}, key string) *time.Time {
	value := properties[key]
	if value == nil {
		return nil
	}
	v, ok := value.([]interface{})
	if !ok {
		return nil
	}
	v, ok = v[0].([]interface{})
	if !ok {
		return nil
	}
	v, ok = v[1].([]interface{})
	if !ok {
		return nil
	}
	v, ok = v[0].([]interface{})
	if !ok {
		return nil
	}
	d := v[1].(map[string]interface{})
	if v, ok := d["type"]; !ok {
		return nil
	} else {
		t, ok := v.(string)
		if !ok {
			return nil
		}
		if t != "date" {
			return nil
		}
	}
	sd, ok := d["start_date"]
	if !ok {
		return nil
	}
	startDateValue, ok := sd.(string)
	if !ok {
		return nil
	}
	startDate, err := time.Parse("2006-01-02", startDateValue)
	if err != nil {
		return nil
	}

	return &startDate
}

func checkboxPropertyValue(properties map[string]interface{}, key string) bool {
	value := properties[key]
	if value == nil {
		return false
	}
	v, ok := value.([]interface{})
	if !ok {
		return false
	}
	v, ok = v[0].([]interface{})
	if !ok {
		return false
	}
	checkbox, ok := v[0].(string)
	if !ok {
		return false
	}

	if checkbox == "No" {
		return false
	}

	return true
}

func textPropertyValue(properties map[string]interface{}, key string) string {
	value := properties[key]
	if value == nil {
		return ""
	}
	v, ok := value.([]interface{})
	if !ok {
		return ""
	}
	v, ok = v[0].([]interface{})
	if !ok {
		return ""
	}
	text, ok := v[0].(string)
	if !ok {
		return ""
	}

	return text
}

func multiSelectPropertyValue(properties map[string]interface{}, key string) []string {
	value := textPropertyValue(properties, key)
	if value == "" {
		return nil
	}
	return strings.Split(value, ",")
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
