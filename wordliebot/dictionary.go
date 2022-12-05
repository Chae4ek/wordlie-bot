package wordliebot

// Represents the dictionary of words
type MutableDictionary struct {
	firstLetterToWords map[rune][]string
	wordToID           map[string]int
	wordCount          int
}

// Used to hide words from the dictionary
type DictionaryView struct {
	dictionary                       *MutableDictionary
	firstLetterToCountOfSkippedWords map[rune]int // The maximum first words to skip in each mapped array of the dictionary
	hiddenWords                      *Set[string] // Contains words which cannot be in the first skipped words
}

func NewMutableDictionary() *MutableDictionary {
	return &MutableDictionary{firstLetterToWords: make(map[rune][]string), wordToID: make(map[string]int)}
}

func NewDictionaryView(dictionary *MutableDictionary) *DictionaryView {
	return &DictionaryView{dictionary: dictionary, firstLetterToCountOfSkippedWords: make(map[rune]int), hiddenWords: NewSet[string]()}
}

func (dictionary *MutableDictionary) AddWord(word string) {
	firstLetter := []rune(word)[0]
	words := dictionary.firstLetterToWords[firstLetter]

	if words == nil {
		words = make([]string, 0)
	}

	dictionary.wordToID[word] = dictionary.wordCount
	dictionary.wordCount++
	dictionary.firstLetterToWords[firstLetter] = append(words, word)
}

// Returns true if the word is successfully hidden, otherwise false
func (view *DictionaryView) HideWord(word string) bool {
	if view.isHidden(word) {
		return false
	}

	isNotHidden := !view.hiddenWords.Contains(word)
	if isNotHidden {
		view.hiddenWords.Add(word)
		view.tryToIncreaseSkipIndex(word)
	}
	return isNotHidden
}

func (view *DictionaryView) isHidden(word string) bool {
	firstLetter := []rune(word)[0]
	skipIndex := view.firstLetterToCountOfSkippedWords[firstLetter] - 1
	if skipIndex < 0 {
		return false
	}
	words := view.dictionary.firstLetterToWords[firstLetter]
	wordID := view.dictionary.wordToID[word]

	skippedWord := words[skipIndex]
	skippedWordID := view.dictionary.wordToID[skippedWord]

	return wordID <= skippedWordID
}

func (view *DictionaryView) tryToIncreaseSkipIndex(word string) {
	firstLetter := []rune(word)[0]
	skipIndex := view.firstLetterToCountOfSkippedWords[firstLetter] - 1
	words := view.dictionary.firstLetterToWords[firstLetter]

	wordsLength := len(words)
	for skipIndex++; skipIndex < wordsLength; skipIndex++ {
		wordToSkip := words[skipIndex]
		if !view.hiddenWords.Contains(wordToSkip) {
			break
		}
		view.hiddenWords.Delete(wordToSkip)
	}
	skipIndex--

	view.firstLetterToCountOfSkippedWords[firstLetter] = skipIndex + 1
}

// Returns (word, true) if the word was found, otherwise ("", false)
func (view *DictionaryView) GetMostFrequentWordBy(firstLetter rune) (string, bool) {
	wordCountToSkip := view.firstLetterToCountOfSkippedWords[firstLetter]
	words := view.dictionary.firstLetterToWords[firstLetter]

	if wordCountToSkip == len(words) {
		return "", false
	}

	return words[wordCountToSkip], true
}
