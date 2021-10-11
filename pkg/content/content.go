package content

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kjk/notionapi"
	"golang.org/x/xerrors"
	"gopkg.in/yaml.v2"
)

func UpdateContent(client *notionapi.Client, pageId, dir string) error {
	articles, err := GetArticles(client, pageId)
	if err != nil {
		return xerrors.Errorf(": %w", err)
	}

	for _, v := range articles {
		if v.Draft {
			continue
		}

		dirName := strings.ReplaceAll(strings.ToLower(v.EnglishTitle), " ", "-")
		dirName = strings.ReplaceAll(dirName, "'", "")
		if _, err := os.Stat(filepath.Join(dir, dirName)); os.IsNotExist(err) {
			log.Printf("Mkdir %s", filepath.Join(dir, dirName))
			if err := os.MkdirAll(filepath.Join(dir, dirName), 0755); err != nil {
				return xerrors.Errorf(": %w", err)
			}
		}

		if _, err := os.Stat(filepath.Join(dir, dirName, "index.md")); os.IsNotExist(err) {
			body, err := v.GetBody(client)
			if err != nil {
				return xerrors.Errorf(": %w", err)
			}

			c, err := v.Render(body)
			if err != nil {
				return xerrors.Errorf(": %w", err)
			}
			log.Printf("Create %s", filepath.Join(dir, dirName, "index.md"))
			err = ioutil.WriteFile(filepath.Join(dir, dirName, "index.md"), c, 0644)
			if err != nil {
				return xerrors.Errorf(": %w", err)
			}
			for _, file := range v.Files {
				log.Printf("Create %s", filepath.Join(dir, dirName, file.Filename))
				err = ioutil.WriteFile(filepath.Join(dir, dirName, file.Filename), file.Data, 0644)
				if err != nil {
					return xerrors.Errorf(": %w", err)
				}
			}
		} else {
			if v.Freeze {
				continue
			}

			lastmod, err := LastMod(filepath.Join(dir, dirName, "index.md"))
			if err != nil {
				return xerrors.Errorf(": %w", err)
			}
			if v.UpdatedAt.After(lastmod) && v.UpdatedAt.Sub(v.CreatedAt) > 24*time.Hour {
				body, err := v.GetBody(client)
				if err != nil {
					return xerrors.Errorf(": %w", err)
				}
				c, err := v.Render(body)
				if err != nil {
					return xerrors.Errorf(": %w", err)
				}
				log.Printf("Update %s", filepath.Join(dir, dirName, "index.md"))
				err = ioutil.WriteFile(filepath.Join(dir, dirName, "index.md"), c, 0644)
				if err != nil {
					return xerrors.Errorf(": %w", err)
				}
				for _, file := range v.Files {
					log.Printf("Create %s", filepath.Join(dir, dirName, file.Filename))
					err = ioutil.WriteFile(filepath.Join(dir, dirName, file.Filename), file.Data, 0644)
					if err != nil {
						return xerrors.Errorf(": %w", err)
					}
				}
			}
		}
	}

	return nil
}

func ReadDateFromFile(file string) (time.Time, error) {
	f, err := os.Open(file)
	if err != nil {
		return time.Time{}, xerrors.Errorf(": %w", err)
	}
	m := &ArticleMetadata{}
	err = yaml.NewDecoder(f).Decode(m)
	if err != nil {
		return time.Time{}, xerrors.Errorf(": %w", err)
	}

	return m.Date, nil
}

func LastMod(file string) (time.Time, error) {
	f, err := os.Open(file)
	if err != nil {
		return time.Time{}, xerrors.Errorf(": %w", err)
	}
	m := &ArticleMetadata{}
	err = yaml.NewDecoder(f).Decode(m)
	if err != nil {
		return time.Time{}, xerrors.Errorf(": %w", err)
	}
	if !m.LastMod.IsZero() {
		return m.LastMod, nil
	}

	return m.Date, nil
}
