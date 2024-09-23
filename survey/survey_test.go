package survey

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSurvey_MarshalUnmarshal(t *testing.T) {
	// Create sample questions
	q1 := &multipleChoiceQuestion{
		selectQuestion{
			Options: []answerChoice{{Selected: false}, {Selected: true}, {Selected: false}, {Selected: false}},
		},
	}

	q2 := &selectAllThatApplyQuestion{
		selectQuestion{
			Options: []answerChoice{{Selected: true}, {Selected: false}, {Selected: true}, {Selected: true}},
		},
	}

	q3 := &textEntryQuestion{
		Text: "Test answer",
	}

	svy := &survey{
		version: 1,
		questions: map[uint8]question{
			1: q1,
			2: q2,
			3: q3,
		},
	}

	// Marshal te survey
	data, err := svy.marshal()
	assert.NoError(t, err)

	newSurvey := &survey{}
	err = newSurvey.unmarshal(data)
	assert.NoError(t, err)

	assert.True(t, surveysEqual(svy, newSurvey))
}

func TestSurvey_Unmarshal_InvalidData(t *testing.T) {
	// Test: Data too short
	data := []byte{1}
	svy := &survey{}
	err := svy.unmarshal(data)
	assert.Error(t, err)
	assert.Equal(t, "no questions", err.Error())

	// Test: Invalid question data (not enough bytes for question header)
	data = []byte{1, 1, 1, 2}
	svy = &survey{}
	err = svy.unmarshal(data)
	assert.Error(t, err)
	assert.Equal(t, "invalid question data", err.Error())

	// Test: Unknown question type
	data = []byte{1, 1, 1, 255, 0, 0}
	svy = &survey{}
	err = svy.unmarshal(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown question type")

	// Test: Question payload is too short
	data = []byte{1, 1, 1, 0, 5}
	svy = &survey{}
	err = svy.unmarshal(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid question data")

	// Test: Failure in question unmarshal
	data = []byte{
		1,    // Version
		1,    // Number of questions
		1,    // Question ID
		0,    // Question type (MultipleChoice)
		2,    // Length of question data
		2,    // Number of options (1)
		0xff, // Multiple options selected
	}
	svy = &survey{}
	err = svy.unmarshal(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal question 1")
}

func TestSurvey_Marshal_ErrorInQuestionMarshal(t *testing.T) {
	// Create a TextEntryQuestion with text too long to marshal
	q1 := &textEntryQuestion{
		Text: string(make([]byte, 256)), // Text length is 256, which is too long
	}

	svy := &survey{
		version: 1,
		questions: map[uint8]question{
			1: q1,
		},
	}

	// Marshal the survey and expect an error
	_, err := svy.marshal()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal question 1")
}

func surveysEqual(s1, s2 *survey) bool {
	if s1.version != s2.version {
		return false
	}
	if len(s1.questions) != len(s2.questions) {
		return false
	}
	for id, q1 := range s1.questions {
		q2, ok := s2.questions[id]
		if !ok {
			return false
		}
		if !questionsEqual(q1, q2) {
			return false
		}
	}
	return true
}

func questionsEqual(q1, q2 question) bool {
	if q1.getType() != q2.getType() {
		return false
	}
	switch q1Typed := q1.(type) {
	case *multipleChoiceQuestion:
		q2Typed, ok := q2.(*multipleChoiceQuestion)
		if !ok {
			return false
		}
		return equalAnswerChoices(q1Typed.Options, q2Typed.Options)
	case *selectAllThatApplyQuestion:
		q2Typed, ok := q2.(*selectAllThatApplyQuestion)
		if !ok {
			return false
		}
		return equalAnswerChoices(q1Typed.Options, q2Typed.Options)
	case *textEntryQuestion:
		q2Typed, ok := q2.(*textEntryQuestion)
		if !ok {
			return false
		}
		return q1Typed.Text == q2Typed.Text
	default:
		// Unknown question type
		return false
	}
}
