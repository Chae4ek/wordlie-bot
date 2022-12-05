package wordliebot

import (
	"bufio"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type WordlieBot struct {
	*tgbotapi.BotAPI
	mutex        sync.Mutex
	sessions     map[int64]*GameSession
	dictionary   *MutableDictionary
	firstLetters []rune
}

type GameSession struct {
	userID   int64
	view     *DictionaryView
	lastWord []rune
	started  bool
}

func NewWordlieBot(tgbot *tgbotapi.BotAPI, dictionaryFileName string) *WordlieBot {
	dictionary := NewMutableDictionary()

	file, err := os.Open(dictionaryFileName)
	if err != nil {
		log.Panic(err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Panic(err)
		}
	}()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		word := scanner.Text()
		dictionary.AddWord(word)
	}

	firstLetters := make([]rune, 0, len(dictionary.firstLetterToWords))
	for letter := range dictionary.firstLetterToWords {
		firstLetters = append(firstLetters, letter)
	}

	return &WordlieBot{BotAPI: tgbot, sessions: make(map[int64]*GameSession), dictionary: dictionary, firstLetters: firstLetters}
}

func (bot *WordlieBot) ProcessIncomingMessage(incomingMsg *tgbotapi.Message) {
	userID := incomingMsg.Chat.ID

	if incomingMsg.IsCommand() {
		bot.processCommand(userID, incomingMsg.Command())
		return
	}

	bot.processTextMsg(userID, incomingMsg.Text)
}

func (bot *WordlieBot) ProcessIncomingCallbackQuery(callbackQuery *tgbotapi.CallbackQuery) {
	userID := callbackQuery.Message.Chat.ID
	bot.processCommand(userID, "idk")
}

func (bot *WordlieBot) processCommand(userID int64, command string) {
	switch command {
	case "idk":
		session := bot.sessions[userID]
		if session == nil || !session.started {
			bot.sendText(userID, "You haven't started the game yet")
		} else {
			bot.tryToSendNextWord(userID, session, session.lastWord[0])
		}
	case "start":
		bot.help(userID)
	case "help":
		bot.help(userID)
	case "game":
		session := bot.getGameSession(userID)
		session.endGameSession()
		rand.Seed(time.Now().UnixNano())
		word, _ := session.view.GetMostFrequentWordBy(bot.firstLetters[rand.Intn(len(bot.firstLetters))])
		session.view.HideWord(word)
		session.started = true
		session.lastWord = []rune(word)
		bot.sendTextWithButton(userID, "First word is: "+word)
	case "end":
		session := bot.sessions[userID]
		if session == nil || !session.started {
			bot.sendText(userID, "You haven't started the game yet")
			return
		}
		session.endGameSession()
		bot.sendText(userID, "You lose)")
	default:
		bot.sendText(userID, "Unknown command. Type /help")
	}
}

func (bot *WordlieBot) help(userID int64) {
	msg := tgbotapi.NewMessage(userID, "This is a telegram bot to play a word game.\n"+
		"The rules are simple:\n\n"+
		"- You or the bot write the first word (type /game to make bot start).\n"+
		"- Then, one by one, you write the words that begin with the last letter of the previous word.\n"+
		"- If you or the bot don't know a word, you'll have to come up with another one (type /idk or press the button to make bot come up another word).\n"+
		"- If you or the bot cannot come up any word, the game will be end (type /end to give u[p](https://www.youtube.com/watch?v=dQw4w9WgXcQ)).")
	msg.ParseMode = "markdown"
	msg.DisableWebPagePreview = true
	bot.send(msg)
}

func (bot *WordlieBot) processTextMsg(userID int64, msg string) {
	session := bot.getGameSession(userID)
	msgUTF8 := []rune(msg)

	if session.started {
		lastSessionLetter := session.lastWord[len(session.lastWord)-1]
		if msgUTF8[0] != lastSessionLetter {
			bot.sendText(userID, "Your word should start with '"+string(lastSessionLetter)+"'. Try again or type /end to give up!")
			return
		}
	}

	if _, exists := bot.dictionary.wordToID[msg]; !exists {
		replyMsg := "I don't know this word. Try again"
		if session.started {
			replyMsg += " or type /end to give up!"
		}
		bot.sendText(userID, replyMsg)
		return
	}

	if hidden := session.view.HideWord(msg); !hidden {
		bot.sendText(userID, "This word is already used. Try again or type /end to give up!")
		return
	}
	session.started = true
	session.lastWord = msgUTF8

	bot.tryToSendNextWord(userID, session, msgUTF8[len(msgUTF8)-1])
}

func (bot *WordlieBot) tryToSendNextWord(userID int64, session *GameSession, firstLetter rune) {
	if word, found := session.view.GetMostFrequentWordBy(firstLetter); found {
		session.view.HideWord(word)
		session.lastWord = []rune(word)
		bot.sendTextWithButton(userID, word)
		return
	}
	session.endGameSession()
	bot.sendText(userID, "I give up. You won")
}

func (bot *WordlieBot) getGameSession(userID int64) *GameSession {
	session := bot.sessions[userID]
	if session == nil {
		bot.mutex.Lock()
		session = bot.sessions[userID]
		if session == nil {
			session = &GameSession{userID: userID, view: NewDictionaryView(bot.dictionary)}
			bot.sessions[userID] = session
		}
		bot.mutex.Unlock()
	}
	return session
}

func (session *GameSession) endGameSession() {
	session.started = false
	session.view = NewDictionaryView(session.view.dictionary)
	session.lastWord = nil
}

func (bot *WordlieBot) sendTextWithButton(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = numericKeyboard
	bot.send(msg)
}

func (bot *WordlieBot) sendText(chatID int64, text string) {
	bot.send(tgbotapi.NewMessage(chatID, text))
}

func (bot *WordlieBot) send(msg tgbotapi.Chattable) {
	if _, err := bot.Send(msg); err != nil {
		log.Print(err.Error())
	}
}

var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("I don't know this word", "_")))
