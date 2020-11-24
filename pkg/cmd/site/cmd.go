package site

import (
	"os"

	"github.com/kjk/notionapi"
	"github.com/spf13/cobra"
	"golang.org/x/xerrors"

	"github.com/f110/site/pkg/content"
)

func Update(rootCmd *cobra.Command) {
	dir := ""
	pageId := ""

	updateCmd := &cobra.Command{
		Use:   "update-content",
		Short: "Update contents from notion",
		Long: `update-content will download contents from notion and make a markdown for hugo.
accessing notion will be required authentication. NOTION_COOKIE_VALUE is the token of notion api.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return xerrors.Errorf(": %w", err)
			}

			client := &notionapi.Client{
				AuthToken: os.Getenv("NOTION_COOKIE_VALUE"),
			}

			return content.UpdateContent(client, pageId, dir)
		},
	}
	fs := updateCmd.Flags()
	fs.StringVar(&dir, "dir", "", "Output directory")
	fs.StringVar(&pageId, "id", "", "Page id")

	rootCmd.AddCommand(updateCmd)
}
