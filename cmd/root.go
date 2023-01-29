package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/briandowns/spinner"
	"github.com/mergestat/scribe/internal/prompt"
	"github.com/mergestat/scribe/pkg/introspect"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
)

var (
	connection       string
	examplesFilePath string
)

func init() {
	rootCmd.PersistentFlags().StringVar(&connection, "connection", os.Getenv("SCRIBE_CONNECTION"), "connection string to the SQL database")
	rootCmd.PersistentFlags().StringVar(&examplesFilePath, "examples", "", "path to the file containing SQL examples")
}

var rootCmd = &cobra.Command{
	Use:   "scribe",
	Short: "Scribe is a CLI for translating natural language prompts into SQL queries",
	Long:  `Scribe is a CLI for translating natural language prompts into SQL queries.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("missing prompt")
			if err := cmd.Root().Help(); err != nil {
				log.Fatal(err)
			}
			os.Exit(1)
		}

		if connection == "" {
			fmt.Println("missing connection string")
			if err := cmd.Root().Help(); err != nil {
				log.Fatal(err)
			}
			os.Exit(1)
		}

		s := spinner.New(spinner.CharSets[40], 100*time.Millisecond)
		s.Suffix = " introspecting SQL database..."
		s.Start()
		defer s.Stop()

		db, err := introspect.Introspect(connection, &introspect.Options{})
		if err != nil {
			log.Fatal(err)
		}

		promptPrefix, err := prompt.GeneratePromptPrefixFromDB(db)
		if err != nil {
			log.Fatal(err)
		}

		// if the user supplied a file with examples, append them to the prompt
		if examplesFilePath != "" {
			cwd, err := os.Getwd()
			if err != nil {
				log.Fatal(err)
			}

			p := filepath.Join(cwd, examplesFilePath)

			examples, err := ioutil.ReadFile(p)
			if err != nil {
				log.Fatal(err)
			}

			promptPrefix += "\n" + string(examples)
		}

		s.Suffix = " generating SQL from prompt..."

		// generate the SQL query from the prompt by calling the OpenAI API
		res, err := prompt.CodexToSQL(cmd.Context(), promptPrefix, args[0])
		if err != nil {
			log.Fatal(err)
		}

		s.Stop()

		// determine the lexer to use for syntax highlighting based on the database driver
		lexer := "sql"
		switch db.Driver {
		case "pg", "postgres", "pgsql":
			lexer = "postgres"
		}

		format := "terminal256"
		// detect if we're running in a terminal to determine the output format
		// if we're not in a terminal, don't colorize the output for easier piping
		// https://stackoverflow.com/questions/68889637/is-it-possible-to-detect-if-a-writer-is-tty-or-not
		_, err = unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
		if err != nil {
			format = ""
		}

		// colorize the output
		err = quick.Highlight(os.Stdout, res, lexer, format, "monokai")
		if err != nil {
			log.Fatal(err)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
