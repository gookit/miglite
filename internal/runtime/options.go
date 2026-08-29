package runtime

type InitOption struct{ Drop bool }
type StatusOption struct{}
type ShowOption struct {
	Tables bool
	Schema string
}
