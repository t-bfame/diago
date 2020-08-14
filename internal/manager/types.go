package manager

type TestInstance struct {
	Id       string
	Name     string
	Image    string
	Priority int
	Env      map[string]string
}
