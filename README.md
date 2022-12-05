# Wordlie Bot

This is a telegram bot to play a word game.<br/>
The rules are simple:

- You or the bot write the first word.
- Then, one by one, you write the words that begin with the last letter of the previous word.
- If you or the bot don't know a word, you'll have to come up with another one.

Dictionary source: http://norvig.com/ngrams/

## Compilation

First you need to create a `.env` file like this:

```properties
TELEGRAM_APITOKEN = "TOKEN"
```

Replace `TOKEN` to a private telegram bot token.

Then just run:

```
go run main.go
```

Or if you want to build it into an execution file:

```
go build
```
