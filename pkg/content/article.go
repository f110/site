package content

import (
	"bufio"
	"bytes"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/kjk/notionapi"
	"github.com/kjk/notionapi/tomarkdown"
	"golang.org/x/xerrors"
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

func (a *Article) GetBody(client *notionapi.Client) ([]byte, error) {
	p, err := client.DownloadPage(a.ID)
	if err != nil {
		return nil, xerrors.Errorf(": %w", err)
	}
	for _, v := range p.Root().Content {
		if v.Type == notionapi.BlockImage {
			u, err := url.Parse(v.Source)
			if err != nil {
				return nil, xerrors.Errorf(": %w", err)
			}
			filename := filepath.Base(u.Path)
			basename := strings.TrimSuffix(filename, filepath.Ext(filename))
			newFilename := basename + "-" + v.FileIDs[0] + filepath.Ext(filename)
			buf, err := fetchFile(client, v)
			if err != nil {
				return nil, xerrors.Errorf(": %w", err)
			}
			a.Files = append(a.Files, ArticleFile{
				Filename: newFilename,
				Data:     buf,
			})
		}
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

	q := page.CollectionViewByID(page.CollectionViewRecords[0].ID).Query2
	collection, err := client.QueryCollection(
		page.CollectionRecords[0].ID,
		page.CollectionViewRecords[0].ID,
		q,
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
		draft := checkboxPropertyValue(r.Block.Properties, db.Properties["Draft"].ID)
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
			Draft:            draft,
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

func fetchFile(client *notionapi.Client, block *notionapi.Block) ([]byte, error) {
	var u *url.URL
	switch block.Type {
	case notionapi.BlockImage:
		values := &url.Values{}
		values.Set("table", "block")
		values.Set("id", block.ID)
		values.Set("spaceId", block.RawJSON["space_id"].(string))
		values.Set("userId", block.CreatedByID)
		values.Set("cache", "v2")
		imageURL, err := url.Parse("https://www.notion.so/image/" + url.PathEscape(block.Source))
		if err != nil {
			return nil, xerrors.Errorf(": %w", err)
		}
		imageURL.RawQuery = values.Encode()
		u = imageURL
	}
	if u == nil {
		return nil, nil
	}

	res, err := client.DownloadFile(u.String(), block.ID)
	if err != nil {
		return nil, xerrors.Errorf(": %w", err)
	}

	return res.Data, nil
}
