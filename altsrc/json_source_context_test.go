package altsrc

import "testing"

func TestJsonSourceContext_Json(t *testing.T) {
	ctx, err := NewJSONSource([]byte(`{"test":{"key":"value"}}`))
	expect(t, nil, err)

	export, err := ctx.(*jsonSource).Json("test")
	expect(t, []byte(`{"key":"value"}`), export)
	expect(t, nil, err)
}
