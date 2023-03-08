package template

import (
	"testing"
)

func TestExecuteTemplateInStringField(t *testing.T) {
	type input struct {
		A, B string
		C    int
	}
	tpl := newGenericTemplate(&input{"{{.Payload}}", "text", 6}, &Environment{"Name", "Replaced"})
	if err := tpl.execute(); err != nil {
		t.Error(err)
		return
	}
	if tpl.target.(*input).A != "Replaced" {
		t.Errorf("Expected to replace Payload, got %q", tpl.target.(*input).A)
	}
	if tpl.target.(*input).B != "text" {
		t.Errorf("Exptected to not replace static text, got %q", tpl.target.(*input).B)
	}
}

func TestExecuteTemplateInEmbeddedField(t *testing.T) {
	type Embedded struct{ C string }
	type input struct {
		E1 *Embedded
		E2 Embedded
	}
	tpl := newGenericTemplate(&input{&Embedded{"{{.Name | upper}}"}, Embedded{"{{.Payload}}"}}, &Environment{"Name", "Replaced"})
	if err := tpl.execute(); err != nil {
		t.Error(err)
		return
	}
	if tpl.target.(*input).E1.C != "NAME" {
		t.Errorf("Expected to replace Name, got %q", tpl.target.(*input).E1.C)
	}
	if tpl.target.(*input).E2.C != "Replaced" {
		t.Errorf("Expected to replace Payload, got %q", tpl.target.(*input).E2.C)
	}
}

func TestExecuteTemplateInSlice(t *testing.T) {
	type input struct{ A []string }
	tpl := newGenericTemplate(&input{[]string{"{{.Name}}", "1", "1{{.Payload}}3"}}, &Environment{"Name", "2"})
	if err := tpl.execute(); err != nil {
		t.Error(err)
		return
	}
	for i, exp := range []string{"Name", "1", "123"} {
		if tpl.target.(*input).A[i] != exp {
			t.Errorf("Mismatch value in index %d, got %q expecting %q", i, tpl.target.(*input).A[i], exp)
		}
	}
}

func TestExecuteTemplateInEmbbededSlice(t *testing.T) {
	type Embedded struct{ A string }
	type input struct {
		E1 *[]Embedded
		E2 []Embedded
	}
	in := &input{
		&([]Embedded{Embedded{"{{.Payload}}"}}),
		[]Embedded{Embedded{"{{.Name}}"}},
	}
	tpl := newGenericTemplate(in, &Environment{"2", "1"})
	if err := tpl.execute(); err != nil {
		t.Error(err)
		return
	}
	if (*tpl.target.(*input).E1)[0].A != "1" {
		t.Errorf("Mismatch value got %q expecting %q", (*tpl.target.(*input).E1)[0].A, "1")
	}
	if tpl.target.(*input).E2[0].A != "2" {
		t.Errorf("Mismatch value got %q expecting %q", tpl.target.(*input).E2[0].A, "2")
	}
}

func TestExecuteTemplateWithAnError(t *testing.T) {
	type input struct{ A string }
	tpl := newGenericTemplate(&input{"{{.NotFound}}"}, &Environment{})
	if err := tpl.execute(); err == nil {
		t.Error("Expecting error, got nothing")
	}
}

func TestExecuteTemplateWithComplexSprigFunction(t *testing.T) {
	type input struct{ A string }
	in := &input{`{{.Payload | replace "\n" ""}}`}
	tpl := newGenericTemplate(in, &Environment{"", `{
"hello": "world"
}`})
	if err := tpl.execute(); err != nil {
		t.Error(err)
		return
	}
	if in.A != "{&#34;hello&#34;: &#34;world&#34;}" {
		t.Error("Mismatch in generated output", in.A)
	}
}
