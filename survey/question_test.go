package survey

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectAllThatApplyQuestion(t *testing.T) {
	q1 := &selectAllThatApplyQuestion{
		selectQuestion{
			Options: []answerChoice{{Selected: true}, {Selected: false}, {Selected: false}, {Selected: true}, {Selected: false}, {Selected: false}},
		},
	}
	expectedData := []byte{6, 0b10010000}
	data, err := q1.marshal()
	assert.NoError(t, err)
	assert.Equal(t, expectedData, data)

	q1unmarshaled := &selectAllThatApplyQuestion{}
	err = q1unmarshaled.unmarshal(expectedData)
	assert.NoError(t, err)
	assert.True(t, equalAnswerChoices(q1.Options, q1unmarshaled.Options))

	q2 := &selectAllThatApplyQuestion{
		selectQuestion{
			Options: []answerChoice{{Selected: true}, {Selected: true}, {Selected: false}, {Selected: true}, {Selected: true}, {Selected: true}, {Selected: true}, {Selected: true}, {Selected: false}, {Selected: true}, {Selected: true}, {Selected: true}},
		},
	}
	expectedData = []byte{12, 0b11011111, 0b01110000}
	data, err = q2.marshal()
	assert.NoError(t, err)
	assert.Equal(t, expectedData, data)

	q2unmarshaled := &selectAllThatApplyQuestion{}
	err = q2unmarshaled.unmarshal(expectedData)
	assert.NoError(t, err)
	assert.True(t, equalAnswerChoices(q2.Options, q2unmarshaled.Options))

	// marshal error case, more than 255 options
	q3 := &selectAllThatApplyQuestion{}
	for i := 0; i < 256; i++ {
		q3.Options = append(q3.Options, answerChoice{Selected: true})
	}
	_, err = q3.marshal()
	assert.Error(t, err)

	// unmarshal error case, length does not match expected
	q4 := &selectAllThatApplyQuestion{}
	err = q4.unmarshal([]byte{10, 0b10000000})
	assert.Error(t, err)
	q5 := &selectAllThatApplyQuestion{}
	err = q5.unmarshal([]byte{10})
	assert.Error(t, err)
}

func TestMultipleChoiceQuestions(t *testing.T) {
	// same backend as SelectAllThatApplyQuestion,
	// just verifies that only one option is selected
	q1 := &multipleChoiceQuestion{
		selectQuestion{
			Options: []answerChoice{{Selected: true}, {Selected: false}, {Selected: false}, {Selected: false}, {Selected: false}, {Selected: false}},
		},
	}
	expectedData := []byte{6, 0b10000000}
	data, err := q1.marshal()
	assert.NoError(t, err)
	assert.Equal(t, expectedData, data)

	q1unmarshaled := &multipleChoiceQuestion{}
	err = q1unmarshaled.unmarshal(expectedData)
	assert.NoError(t, err)
	assert.True(t, equalAnswerChoices(q1.Options, q1unmarshaled.Options))

	q2 := &multipleChoiceQuestion{
		selectQuestion{
			Options: []answerChoice{{Selected: true}, {Selected: false}, {Selected: true}, {Selected: false}, {Selected: false}, {Selected: false}},
		},
	}
	_, err = q2.marshal()
	assert.Error(t, err)
	err = q2.unmarshal([]byte{6, 0b10000100})
	assert.Error(t, err)
}

func TestTextEntryQuestion(t *testing.T) {
	// Test: Normal Case
	q1 := &textEntryQuestion{Text: "Hello, world!"}
	expectedData := []byte{byte(len(q1.Text)), 'H', 'e', 'l', 'l', 'o', ',', ' ', 'w', 'o', 'r', 'l', 'd', '!'}
	data, err := q1.marshal()
	assert.NoError(t, err)
	assert.Equal(t, expectedData, data)

	q1unmarshaled := &textEntryQuestion{}
	err = q1unmarshaled.unmarshal(expectedData)
	assert.NoError(t, err)
	assert.Equal(t, q1.Text, q1unmarshaled.Text)

	// Test: Empty Text
	q2 := &textEntryQuestion{Text: ""}
	expectedDataEmpty := []byte{0}
	dataEmpty, err := q2.marshal()
	assert.NoError(t, err)
	assert.Equal(t, expectedDataEmpty, dataEmpty)

	q2unmarshaled := &textEntryQuestion{}
	err = q2unmarshaled.unmarshal(expectedDataEmpty)
	assert.NoError(t, err)
	assert.Equal(t, q2.Text, q2unmarshaled.Text)

	// Test: Maximum Length Text
	maxText := string(make([]byte, 255))
	q3 := &textEntryQuestion{Text: maxText}
	expectedDataMax := append([]byte{255}, []byte(maxText)...)
	dataMax, err := q3.marshal()
	assert.NoError(t, err)
	assert.Equal(t, expectedDataMax, dataMax)

	q3unmarshaled := &textEntryQuestion{}
	err = q3unmarshaled.unmarshal(expectedDataMax)
	assert.NoError(t, err)
	assert.Equal(t, q3.Text, q3unmarshaled.Text)

	// Test: Text Too Long
	q4 := &textEntryQuestion{Text: string(make([]byte, 256))}
	_, err = q4.marshal()
	assert.Error(t, err)
	assert.Equal(t, "text too long", err.Error())

	// Test: Invalid Payload Length
	invalidPayload := []byte{5, 't', 'e', 's', 't'}
	q5unmarshaled := &textEntryQuestion{}
	err = q5unmarshaled.unmarshal(invalidPayload)
	assert.Error(t, err)
	assert.Equal(t, "invalid payload length", err.Error())

	// Test: No Data in Payload
	noDataPayload := []byte{}
	q6unmarshaled := &textEntryQuestion{}
	err = q6unmarshaled.unmarshal(noDataPayload)
	assert.Error(t, err)
	assert.Equal(t, "no data", err.Error())

	// Test: censor text
	profanePayload := []byte{4, 'f', 'u', 'c', 'k'}
	q7 := &textEntryQuestion{}
	err = q7.unmarshal(profanePayload)
	assert.NoError(t, err)
	assert.Equal(t, "****", q7.Text)
}

func equalAnswerChoices(a, b []answerChoice) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
