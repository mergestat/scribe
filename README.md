# Scribe 📝

`scribe` is a command line interface for translating natural language prompts into SQL queries.
It makes use of [OpenAI codex models](https://beta.openai.com/docs/models/codex) via the OpenAI API to execute translations.

The user is expected to supply a connection string to the `scribe` command, so that the underlying SQL database may be introspected to supply tables, views, column names and column types to the prompt context.

![Demo GIF](docs/demo.gif)

## Usage

You will need to set the `OPENAI_API_KEY` environment variable.

```sh
scribe "all users with first name patrick" --connection "pg://postgres:password@localhost/?sslmode=disable"
SELECT * FROM users WHERE first_name = 'patrick'
```

The `--connection` flag is used to supply a connection string to the SQL database.
It can also be set with the `SCRIBE_CONNECTION` environment variable.

You may also supply **additional SQL examples** from a local file to improve the quality of translations.

## Installation

### Homebrew

```sh
brew install mergestat/scribe/scribe
```

## Roadmap
- Additional SQL database support (today only `postgres` is supported)
- Interactive "REPL" like terminal experience
- Option to execute queries on the underlying database and return results
