package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"go.f110.dev/notion-api/v3"
	"go.f110.dev/xerrors"
	"golang.org/x/oauth2"

	"github.com/f110/site/internal/content"
)

func Update(rootCmd *cobra.Command) {
	dir := ""
	pageId := ""

	updateCmd := &cobra.Command{
		Use:   "update-content",
		Short: "Update contents from notion",
		Long: `update-content will download contents from notion and make a markdown for hugo.
accessing notion will be required authentication. NOTION_TOKEN is the token of notion api.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return xerrors.WithStack(err)
			}

			token := os.Getenv("NOTION_TOKEN")
			ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
			tc := oauth2.NewClient(context.Background(), ts)
			client, err := notion.New(tc, notion.BaseURL)
			if err != nil {
				return xerrors.WithStack(err)
			}

			return content.UpdateContent(client, pageId, dir)
		},
	}
	fs := updateCmd.Flags()
	fs.StringVar(&dir, "dir", "", "Output directory")
	fs.StringVar(&pageId, "id", "", "Database id")

	rootCmd.AddCommand(updateCmd)
}
