package survey

import (
	"fmt"

	goaway "github.com/TwiN/go-away"
)

type questionType byte

const (
	multipleChoice questionType = iota
	selectAllThatApply
	textEntry
)

type question interface {
	getType() questionType
	marshal() ([]byte, error)
	unmarshal(payload []byte) error
	getTitle() string
}

type answerChoice struct {
	Title    string
	Selected bool
}

type baseQuestion struct {
	Title string
}

func (q *baseQuestion) getTitle() string {
	return q.Title
}

type selectQuestion struct {
	baseQuestion
	Options []answerChoice
}

type textEntryQuestion struct {
	baseQuestion
	Text string
}

type selectAllThatApplyQuestion struct {
	selectQuestion
}

type multipleChoiceQuestion struct {
	selectQuestion
}

func (q *multipleChoiceQuestion) getType() questionType {
	return multipleChoice
}

func (q *selectAllThatApplyQuestion) getType() questionType {
	return selectAllThatApply
}

func (q *textEntryQuestion) getType() questionType {
	return textEntry
}

var _ question = (*multipleChoiceQuestion)(nil)
var _ question = (*selectAllThatApplyQuestion)(nil)
var _ question = (*textEntryQuestion)(nil)

// Marshal encodes the SelectQuestion into a byte slice.
// The encoding format is:
// - Byte 0: Number of options (n)
// - Bytes 1 to ceil(n/8): Bit vector representing selected options
//
// For example, if there are 6 options and the first and fourth are selected:
// - Byte 0: 6 (0b00000110)
// - Byte 1: 0b10010000
//
// The bit vector works as follows:
// - Each bit represents an option.
// - Bit 7 of Byte 1 represents Option 0 (most significant bit).
// - Bit 6 of Byte 1 represents Option 1.
// - ...
// - Bit 0 of Byte 1 represents Option 7 (least significant bit).
func (q *selectQuestion) marshal() ([]byte, error) {
	if len(q.Options) > 255 {
		return nil, fmt.Errorf("too many options")
	}
	numOptions := uint8(len(q.Options))
	bitVectorSize := (numOptions + 7) / 8
	bitVector := make([]byte, bitVectorSize)
	for i, choice := range q.Options {
		if choice.Selected {
			byteIndex := i / 8
			bitPosition := 7 - (i % 8)
			bitVector[byteIndex] |= 1 << bitPosition
		}
	}
	payload := make([]byte, 1+len(bitVector))
	payload[0] = numOptions
	copy(payload[1:], bitVector)
	return payload, nil
}

func (q *selectQuestion) unmarshal(payload []byte) error {
	if len(payload) <= 1 {
		return fmt.Errorf("no data")
	}
	numOptions := payload[0]
	bitVectorSize := (numOptions + 7) / 8
	if len(payload) != 1+int(bitVectorSize) {
		return fmt.Errorf("invalid payload length")
	}
	bitVector := payload[1:]
	q.Options = make([]answerChoice, numOptions)
	for i := uint8(0); i < numOptions; i++ {
		byteIndex := i / 8
		bitPosition := 7 - (i % 8)
		if (bitVector[byteIndex] & (1 << bitPosition)) != 0 {
			q.Options[i].Selected = true
		}
	}
	return nil
}

// Marshal for MultipleChoiceQuestion is the same as SelectQuestion,
// but with the constraint that only one option can be selected.
func (q *multipleChoiceQuestion) marshal() ([]byte, error) {
	numSelected := 0
	for _, choice := range q.Options {
		if choice.Selected {
			numSelected++
		}
	}
	if numSelected > 1 {
		return nil, fmt.Errorf("multiple options selected")
	}
	return q.selectQuestion.marshal()
}

// Unmarshal for MultipleChoiceQuestion is the same as SelectQuestion,
// but with the constraint that only one option can be selected.
func (q *multipleChoiceQuestion) unmarshal(payload []byte) error {
	if err := q.selectQuestion.unmarshal(payload); err != nil {
		return err
	}
	numSelected := 0
	for _, choice := range q.Options {
		if choice.Selected {
			numSelected++
		}
	}
	if numSelected > 1 {
		return fmt.Errorf("multiple options selected")
	}
	return nil
}

// Marshal encodes the TextEntryQuestion into a byte slice.
// The encoding format is:
// - Bytes 0: Length of text (n)
// - Bytes 1 to n: UTF-8 encoded text
func (q *textEntryQuestion) marshal() ([]byte, error) {
	textBytes := []byte(goaway.Censor(q.Text))
	if len(textBytes) > 255 {
		return nil, fmt.Errorf("text too long")
	}
	payload := make([]byte, 1+len(textBytes))
	payload[0] = byte(len(textBytes))
	copy(payload[1:], textBytes)
	return payload, nil
}

// Unmarshal decodes the TextEntryQuestion from a byte slice,
// censoring the text in the process.
func (q *textEntryQuestion) unmarshal(payload []byte) error {
	if len(payload) == 0 {
		return fmt.Errorf("no data")
	}
	textLength := payload[0]
	if len(payload) != 1+int(textLength) {
		return fmt.Errorf("invalid payload length")
	}
	textBytes := payload[1:]
	text := string(textBytes)
	// looking at the library, it seems that Censor() *should* retain the length
	// of the original text, but we still should normalize the length just in case
	censoredText := goaway.Censor(text)
	if len(censoredText) > len(text) {
		censoredText = censoredText[:len(text)]
	} else if len(censoredText) < len(text) {
		censoredText += string(make([]byte, len(text)-len(censoredText)))
	}
	copy(payload[1:], []byte(censoredText))
	q.Text = censoredText
	return nil
}
