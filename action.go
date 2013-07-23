package cli

import (
  "fmt"
  "regexp"
)

type Action interface {
  Execute(*Context)
}

type ContextAction struct {
  Function func(context *Context)
}

type PlainAction struct {
  Function func()
}

func (action ContextAction) Execute(context *Context) {
  action.Function(context)
}

func (action PlainAction) Execute(context *Context) {
  action.Function()
}

func ParseAction(function interface{}) Action {
  regex, err := regexp.Compile(`func\(\S+\)=`)

  if err != nil {
    return nil
  }

  match := regex.MatchString(fmt.Sprintf("%s", function))

  if match {
    return ContextAction {function.(func(c *Context))}
  } else {
    return PlainAction {function.(func())}
  }
}
