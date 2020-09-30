package action

// Action is an interface that allows for the definition of
// pre-defined shell commands.
type Action interface {
	Name() string
	Run()
}
