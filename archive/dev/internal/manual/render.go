package manual

import (
	"fmt"
	"io/fs"
	"strings"
)

func RenderPlain(topic string) string {
	data, err := docs.ReadFile("docs/" + topic + ".md")
	if err != nil {
		return fmt.Sprintf("No manual entry for '%s'\n", topic)
	}
	return string(data)
}

func RenderMan(topic string) string {
	// Minimal roff-style wrapper
	data, err := docs.ReadFile("docs/" + topic + ".md")
	if err != nil {
		return fmt.Sprintf(".TH %s 1\n.SH NAME\nNo manual entry available\n", strings.ToUpper(topic))
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf(".TH %s 1\n", strings.ToUpper(topic)))
	b.WriteString(".SH NAME\n")
	b.WriteString(topic + "\n")
	b.WriteString(".SH DESCRIPTION\n")
	b.Write(data)
	return b.String()
}

// ListTopics returns available manual topics
func ListTopics() []string {
	var topics []string
	fs.WalkDir(docs, "docs", func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(path, ".md") {
			name := strings.TrimSuffix(strings.TrimPrefix(path, "docs/"), ".md")
			topics = append(topics, name)
		}
		return nil
	})
	return topics
}
