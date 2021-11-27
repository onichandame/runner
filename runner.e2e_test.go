package runner_test

import (
	"os/exec"
	"testing"

	goutils "github.com/onichandame/go-utils"
	"github.com/onichandame/runner"
	"github.com/stretchr/testify/assert"
)

func TestRunner(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		panic("test requires bash")
	}
	t.Run("lifecycle", func(t *testing.T) {
		t.Run("start", func(t *testing.T) {
			t.Run("can start when ready", func(t *testing.T) {
				r := runner.NewRunner(runner.RunnerConfig{
					Command: "bash",
				})
				defer r.Stop()
				assert.Equal(t, `READY`, r.GetStatus())
				assert.NotPanics(t, func() { r.Start() })
			})
			t.Run("cannot start when running", func(t *testing.T) {
				r := runner.NewRunner(runner.RunnerConfig{
					Command: "bash",
				})
				r.Start()
				defer r.Stop()
				assert.Equal(t, `RUNNING`, r.GetStatus())
				assert.Panics(t, func() { r.Start() })
			})
			t.Run("cannot start when completed", func(t *testing.T) {
				r := runner.NewRunner(runner.RunnerConfig{
					Command: "bash",
					Args:    []string{"--version"},
				})
				r.Start()
				r.Wait()
				assert.Equal(t, `COMPLETED`, r.GetStatus())
				assert.Panics(t, func() { r.Start() })
			})
			t.Run("cannot start when stopped", func(t *testing.T) {
				r := runner.NewRunner(runner.RunnerConfig{
					Command: "bash",
				})
				r.Start()
				r.Stop()
				assert.Equal(t, `STOPPED`, r.GetStatus())
				assert.Panics(t, func() { r.Start() })
			})
			t.Run("cannot start when failed", func(t *testing.T) {
				r := runner.NewRunner(runner.RunnerConfig{
					Command: "bash",
					Args:    []string{"--asdf"},
				})
				r.Start()
				r.Wait()
				assert.Equal(t, `FAILED`, r.GetStatus())
				assert.Panics(t, func() { r.Start() })
			})
		})
		t.Run("wait", func(t *testing.T) {
			t.Run("can not wait before start", func(t *testing.T) {
				r := runner.NewRunner(runner.RunnerConfig{
					Command: "bash",
					Args:    []string{"--version"},
				})
				assert.Equal(t, `READY`, r.GetStatus())
				assert.Panics(t, func() { r.Wait() })
			})
			t.Run("can wait when running", func(t *testing.T) {
				r := runner.NewRunner(runner.RunnerConfig{
					Command: "bash",
				})
				r.Start()
				assert.Equal(t, "RUNNING", r.GetStatus())
				done := make(chan int)
				var err error
				go func() {
					defer goutils.RecoverToErr(&err)
					defer close(done)
					r.Wait()
				}()
				r.Stop()
				<-done
				assert.Nil(t, err)
			})
			t.Run("can wait when completed", func(t *testing.T) {
				r := runner.NewRunner(runner.RunnerConfig{
					Command: "bash",
					Args:    []string{"--version"},
				})
				r.Start()
				r.Wait()
				assert.Equal(t, `COMPLETED`, r.GetStatus())
				assert.NotPanics(t, func() { r.Wait() })
			})
			t.Run("can wait when stopped", func(t *testing.T) {
				r := runner.NewRunner(runner.RunnerConfig{
					Command: "bash",
				})
				r.Start()
				r.Stop()
				r.Wait()
				assert.Equal(t, `STOPPED`, r.GetStatus())
				assert.NotPanics(t, func() { r.Wait() })
			})
			t.Run("can wait when failed", func(t *testing.T) {
				r := runner.NewRunner(runner.RunnerConfig{
					Command: "bash",
					Args:    []string{`--asdf`},
				})
				r.Start()
				r.Wait()
				assert.Equal(t, `FAILED`, r.GetStatus())
				assert.NotNil(t, r.Error)
				assert.Equal(t, r.Error, r.Wait())
			})
		})
	})
}
