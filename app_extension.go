package cli

// AppExtension defines an interface for extensions to implement so they can be added to an App
type AppExtension interface {
	// MyName returns the name this extension should be looked up by when using App.GetExtension()
	MyName() string
}
