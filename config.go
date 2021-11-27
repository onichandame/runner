package runner

type RunnerConfig struct {
	CWD     string
	Command string
	Args    []string
	Env     []string
}
