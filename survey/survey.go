package survey

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v2"
)

type survey struct {
	version   byte
	questions map[uint8]question
}

// marshal encodes the Survey into a byte slice.
// The encoding format is:
// - Byte 0: Survey version
// - Byte 1: Number of questions
// - Bytes 2-: Concatenated question data
//
// Question data is encoded as follows:
// - Byte 0: Question ID
// - Byte 1: Question type
// - Byte 2: Length of question data
// - Bytes 3-: Question payload
func (s *survey) marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.WriteByte(s.version)
	buf.WriteByte(byte(len(s.questions)))
	for id, q := range s.questions {
		buf.WriteByte(id)
		buf.WriteByte(byte(q.getType()))
		data, err := q.marshal()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal question %d: %w", id, err)
		}
		buf.WriteByte(byte(len(data)))
		buf.Write(data)
	}
	return buf.Bytes(), nil
}

func (s *survey) unmarshal(data []byte) error {
	if len(data) <= 2 {
		return fmt.Errorf("no questions")
	}
	s.version = data[0]
	numQuestions := int(data[1])
	s.questions = make(map[uint8]question)

	questionsData := data[2:]
	offset := 0

	for i := 0; i < numQuestions; i++ {
		if len(questionsData[offset:]) <= 3 {
			return fmt.Errorf("invalid question data")
		}

		id := questionsData[offset]
		qType := questionType(questionsData[offset+1])
		qLen := int(questionsData[offset+2])

		if len(questionsData[offset+3:]) < qLen {
			return fmt.Errorf("question payload is too short for id %d", id)
		}

		qData := questionsData[offset+3 : offset+3+qLen]
		offset += 3 + qLen

		var q question
		switch qType {
		case multipleChoice:
			q = &multipleChoiceQuestion{}
		case selectAllThatApply:
			q = &selectAllThatApplyQuestion{}
		case textEntry:
			q = &textEntryQuestion{}
		default:
			return fmt.Errorf("unknown question type %d", qType)
		}

		if err := q.unmarshal(qData); err != nil {
			return fmt.Errorf("failed to unmarshal question %d: %w", id, err)
		}

		s.questions[id] = q
	}

	return nil
}

// yamlQuestion is a helper struct for parsing
type yamlQuestion struct {
	Type    string         `yaml:"type"`
	Title   string         `yaml:"title"`
	Options []answerChoice `yaml:"options,omitempty"`
	Text    string         `yaml:"text,omitempty"`
}

type yamlSurvey struct {
	Version   byte           `yaml:"version"`
	Questions []yamlQuestion `yaml:"questions"`
}

// parseSurveyFromYAML parses a YAML input into a Survey struct
func parseSurveyFromYAML(yamlData []byte) (*survey, error) {
	var yamlSurvey yamlSurvey
	err := yaml.Unmarshal(yamlData, &yamlSurvey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}

	parsedSurvey := &survey{
		version:   yamlSurvey.Version,
		questions: make(map[uint8]question),
	}

	for i, q := range yamlSurvey.Questions {
		var question question

		switch q.Type {
		case "MultipleChoice":
			question = &multipleChoiceQuestion{
				selectQuestion: selectQuestion{
					baseQuestion: baseQuestion{Title: q.Title},
					Options:      q.Options,
				},
			}
		case "SelectAllThatApply":
			question = &selectAllThatApplyQuestion{
				selectQuestion: selectQuestion{
					baseQuestion: baseQuestion{Title: q.Title},
					Options:      q.Options,
				},
			}
		case "TextEntry":
			question = &textEntryQuestion{
				baseQuestion: baseQuestion{Title: q.Title},
				Text:         q.Text,
			}
		default:
			return nil, fmt.Errorf("unknown question type: %s", q.Type)
		}

		parsedSurvey.questions[uint8(i+1)] = question
	}

	return parsedSurvey, nil
}

// func logSurvey(logger *zap.SugaredLogger, s *survey) {
// 	logger.Infow("Survey", "version", s.Version)
// 	for id, question := range s.Questions {
// 		switch q := question.(type) {
// 		case *multipleChoiceQuestion:
// 			logger.Infow("MultipleChoiceQuestion", "id", id, "title", q.getTitle())
// 			for _, opt := range q.Options {
// 				logger.Infow("Option", "title", opt.Title, "selected", opt.Selected)
// 			}
// 		case *selectAllThatApplyQuestion:
// 			logger.Infow("SelectAllThatApplyQuestion", "id", id, "title", q.getTitle())
// 			for _, opt := range q.Options {
// 				logger.Infow("Option", "title", opt.Title, "selected", opt.Selected)
// 			}
// 		case *textEntryQuestion:
// 			logger.Infow("TextEntryQuestion", "id", id, "title", q.getTitle(), "text", q.Text)
// 		}
// 	}
// }
