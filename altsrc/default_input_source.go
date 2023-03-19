package altsrc

// defaultInputSource creates a default InputSourceContext.
func defaultInputSource() (InputSourceContext, error) {
	return &MapInputSource{file: "", valueMap: map[interface{}]interface{}{}}, nil
}
