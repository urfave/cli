package cli

type Options map[string] interface{}

func (o Options) Int(key string) int {
  val := o[key]
  if val != nil {
    return val.(int)
  } else {
    return 0
  }
}

func (o Options) String(key string) string {
  val := o[key]
  if val != nil {
    return val.(string)
  } else {
    return ""
  }
}

func (o Options) Bool(key string) bool {
  val := o[key]
  if val != nil {
    return val.(bool)
  } else {
    return false
  }
}
