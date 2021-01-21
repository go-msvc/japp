package config_test

import (
	"testing"

	"github.com/go-msvc/japp/config"
)

type InOut struct {
	In  string
	Out string
}

func TestContent1(t *testing.T) {
	data := map[string]interface{}{
		"name": "Jan",
		"class": map[string]interface{}{
			"grade":    8,
			"size":     21,
			"students": []string{"John", "Susan"},
		},
	}
	c := []InOut{
		{"123", "123"},
		{"My name is {{.name}}.", "My name is Jan."},
		{"I am in grade {{.class.grade}} with {{.class.size}} students.", "I am in grade 8 with 21 students."},
		{"Friends {{.class.students}}", "Friends [John Susan]"},
		{"Test {{range .class.students}} {{.}} {{end}} END", "Test  John  Susan  END"},
		{"Class is{{range .class.students}} {{ . -}} {{end}}.", "Class is John Susan."},
	}
	for idx, test := range c {
		out, err := config.Text(test.In).Render(data)
		if err != nil || out != test.Out {
			t.Errorf("[%d]: %s -> %s,%v != %s", idx, test.In, out, err, test.Out)
		} else {
			t.Logf("[%d] OK %s -> %s", idx, test.In, out)
		}
	}
}
