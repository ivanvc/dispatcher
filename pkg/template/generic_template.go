package template

import (
	"reflect"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

type genericTemplate struct {
	target any
	env    *Environment
}

func newGenericTemplate(target any, env *Environment) *genericTemplate {
	return &genericTemplate{target, env}
}

func (t *genericTemplate) execute() error {
	return t.recursivelyExecuteTemplate(reflect.ValueOf(t.target))
}

func (t *genericTemplate) recursivelyExecuteTemplate(value reflect.Value) error {
	if value.Kind() == reflect.Pointer {
		value = value.Elem()
	}

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)

		switch field.Kind() {
		case reflect.String:
			if err := t.replaceValueWithRenderedTemplate(field); err != nil {
				return err
			}
		case reflect.Slice, reflect.Array:
			if field.IsNil() {
				continue
			}
			if err := t.walkSlice(field); err != nil {
				return err
			}
		case reflect.Pointer:
			if field.IsNil() {
				continue
			}
			if field.Elem().Kind() == reflect.Slice || field.Elem().Kind() == reflect.Array {
				if err := t.walkSlice(field.Elem()); err != nil {
					return err
				}
			} else if err := t.recursivelyExecuteTemplate(field); err != nil {
				return err
			}
		case reflect.Struct:
			if err := t.recursivelyExecuteTemplate(field); err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *genericTemplate) walkSlice(field reflect.Value) error {
	for j := 0; j < field.Len(); j++ {
		innerField := field.Index(j)

		if innerField.Kind() == reflect.String {
			if err := t.replaceValueWithRenderedTemplate(innerField); err != nil {
				return err
			}
		}

		if innerField.Kind() == reflect.Struct || innerField.Kind() == reflect.Pointer && innerField.Elem().Kind() == reflect.Struct {
			if err := t.recursivelyExecuteTemplate(innerField); err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *genericTemplate) replaceValueWithRenderedTemplate(value reflect.Value) error {
	tpl, err := template.New(value.Type().Name()).Funcs(sprig.TxtFuncMap()).Parse(value.String())
	if err != nil {
		return err
	}

	var b strings.Builder
	if err := tpl.Execute(&b, t.env); err != nil {
		return err
	}

	value.SetString(b.String())

	return nil
}
