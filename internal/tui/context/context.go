package context

type Repository struct {
	Name string
	Path string
}

type ProgramContext struct {
	Width      int
	Height     int
	Repos      []Repository
	ActiveRepo int
	Message    string
}
