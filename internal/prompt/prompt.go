package prompt

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/PullRequestInc/go-gpt3"
	"github.com/mergestat/scribe/pkg/introspect"
)

// CodexToSQL generates SQL from a natural language prompt
func CodexToSQL(ctx context.Context, promptPrefix, prompt string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("missing OPENAI_API_KEY environment variable")
	}

	client := gpt3.NewClient(apiKey)
	var temp float32 = 0
	var topP float32 = 1
	var maxTokens = 512
	res, err := client.CompletionWithEngine(ctx, "code-davinci-002", gpt3.CompletionRequest{
		Prompt:           []string{promptPrefix + prompt + "SELECT"},
		Temperature:      &temp,
		TopP:             &topP,
		FrequencyPenalty: 1,
		Stop:             []string{";"},
		MaxTokens:        &maxTokens,
	})
	if err != nil {
		return "", err
	}

	for _, choice := range res.Choices {
		return "SELECT" + choice.Text, nil
	}

	return "", nil
}

// GeneratePromptPrefixFromDB generates a prompt prefix from an introspected database
func GeneratePromptPrefixFromDB(db *introspect.Database) (string, error) {
	var promptPrefix strings.Builder
	fmt.Fprintf(&promptPrefix, "-- Database: %s\n", db.Driver)
	for schemaName, schema := range db.Schemas {
		for tableName, table := range schema.Tables {
			fmt.Fprintf(&promptPrefix, "-- Table: %s.%s (%s)\n", schemaName, tableName, table.Type)
			for _, column := range table.Columns {
				fmt.Fprintf(&promptPrefix, "--  Column: %s (%s)\n", column.Name, column.Type)
			}
		}
	}
	return promptPrefix.String(), nil
}
