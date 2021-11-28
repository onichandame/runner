package runner_test

import (
	"os/exec"
	"runtime"
	"testing"
	"time"

	goutils "github.com/onichandame/go-utils"
	"github.com/onichandame/runner"
	"github.com/stretchr/testify/assert"
)

func TestRunner(t *testing.T) {
	var term string
	switch runtime.GOOS {
	case `windows`:
		term = `powershell`
	default:
		term = `sh`
	}
	if _, err := exec.LookPath("sleep"); err != nil {
		panic("test requires sleep")
	}
	if _, err := exec.LookPath("echo"); err != nil {
		panic("test requires echo")
	}
	t.Run("lifecycle", func(t *testing.T) {
		t.Run("start", func(t *testing.T) {
			t.Run("can start when ready", func(t *testing.T) {
				r := runner.NewRunner(runner.RunnerConfig{
					Command: "echo",
					Args:    []string{`hi`},
				})
				defer r.Stop()
				assert.Equal(t, `READY`, r.GetStatus())
				assert.NotPanics(t, func() { r.Start() })
			})
			t.Run("cannot start when running", func(t *testing.T) {
				r := runner.NewRunner(runner.RunnerConfig{
					Command: "sleep",
					Args:    []string{`3`},
				})
				r.Start()
				defer r.Stop()
				assert.Equal(t, `RUNNING`, r.GetStatus())
				assert.Panics(t, func() { r.Start() })
			})
			t.Run("cannot start when completed", func(t *testing.T) {
				r := runner.NewRunner(runner.RunnerConfig{
					Command: "echo",
					Args:    []string{"hi"},
				})
				r.Start()
				r.Wait()
				assert.Equal(t, `COMPLETED`, r.GetStatus())
				assert.Panics(t, func() { r.Start() })
			})
			t.Run("cannot start when stopped", func(t *testing.T) {
				r := runner.NewRunner(runner.RunnerConfig{
					Command: "sleep",
					Args:    []string{`3`},
				})
				r.Start()
				r.Stop()
				assert.Equal(t, `STOPPED`, r.GetStatus())
				assert.Panics(t, func() { r.Start() })
			})
			t.Run("cannot start when failed", func(t *testing.T) {
				r := runner.NewRunner(runner.RunnerConfig{
					Command: "sleep",
					Args:    []string{`a`},
				})
				r.Start()
				r.Wait()
				assert.Equal(t, `FAILED`, r.GetStatus())
				assert.Panics(t, func() { r.Start() })
			})
		})
		t.Run("wait", func(t *testing.T) {
			t.Run("can wait before start", func(t *testing.T) {
				r := runner.NewRunner(runner.RunnerConfig{
					Command: "echo",
					Args:    []string{`hi`},
				})
				go func() { time.Sleep(time.Microsecond * 500); r.Start() }()
				assert.Equal(t, `READY`, r.GetStatus())
				assert.NotPanics(t, func() { r.Wait() })
			})
			t.Run("can wait when running", func(t *testing.T) {
				r := runner.NewRunner(runner.RunnerConfig{
					Command: "sleep",
					Args:    []string{`.2`},
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
				<-done
				assert.Nil(t, err)
			})
			t.Run("can wait when completed", func(t *testing.T) {
				r := runner.NewRunner(runner.RunnerConfig{
					Command: "echo",
					Args:    []string{"hi"},
				})
				r.Start()
				r.Wait()
				assert.Equal(t, `COMPLETED`, r.GetStatus())
				assert.NotPanics(t, func() { r.Wait() })
			})
			t.Run("can wait when stopped", func(t *testing.T) {
				r := runner.NewRunner(runner.RunnerConfig{
					Command: "sleep",
					Args:    []string{`3`},
				})
				r.Start()
				r.Stop()
				r.Wait()
				assert.Equal(t, `STOPPED`, r.GetStatus())
				assert.NotPanics(t, func() { r.Wait() })
			})
			t.Run("can wait when failed", func(t *testing.T) {
				r := runner.NewRunner(runner.RunnerConfig{
					Command: "sleep",
					Args:    []string{`a`},
				})
				r.Start()
				r.Wait()
				assert.Equal(t, `FAILED`, r.GetStatus())
				assert.NotNil(t, r.Error)
				assert.Equal(t, r.Error, r.Wait())
			})
		})
		t.Run("stop", func(t *testing.T) {
			t.Run("can not stop when ready", func(t *testing.T) {
				r := runner.NewRunner(runner.RunnerConfig{})
				assert.Equal(t, `READY`, r.GetStatus())
				assert.Panics(t, func() { r.Stop() })
			})
			t.Run("can stop when running", func(t *testing.T) {
				r := runner.NewRunner(runner.RunnerConfig{
					Command: "sleep",
					Args:    []string{`3`},
				})
				r.Start()
				assert.Equal(t, `RUNNING`, r.GetStatus())
				assert.NotPanics(t, func() { r.Stop() })
			})
			t.Run("can not stop when stopped", func(t *testing.T) {
				r := runner.NewRunner(runner.RunnerConfig{
					Command: "sleep",
					Args:    []string{`3`},
				})
				r.Start()
				r.Stop()
				r.Wait()
				assert.Equal(t, `STOPPED`, r.GetStatus())
				assert.Panics(t, func() { r.Stop() })
			})
			t.Run("can not stop when completed", func(t *testing.T) {
				r := runner.NewRunner(runner.RunnerConfig{
					Command: "echo",
					Args:    []string{`hi`},
				})
				r.Start()
				r.Wait()
				assert.Equal(t, `COMPLETED`, r.GetStatus())
				assert.Panics(t, func() { r.Stop() })
			})
			t.Run("can not stop when failed", func(t *testing.T) {
				r := runner.NewRunner(runner.RunnerConfig{
					Command: "sleep",
					Args:    []string{`a`},
				})
				r.Start()
				r.Wait()
				assert.Equal(t, `FAILED`, r.GetStatus())
				assert.Panics(t, func() { r.Stop() })
			})
		})
	})
	t.Run("interaction", func(t *testing.T) {
		t.Run("output", func(t *testing.T) {
			r := runner.NewRunner(runner.RunnerConfig{
				Command: "echo",
				Args:    []string{"hi"},
			})
			outchan := r.ReadOutput()
			r.Start()
			out, ok := <-outchan
			assert.True(t, ok)
			assert.Equal(t, "hi", out)
			r.Wait()
		})
		t.Run("log", func(t *testing.T) {
			r := runner.NewRunner(runner.RunnerConfig{
				Command: `echo`,
				Args:    []string{`hi`},
			})
			r.Start()
			r.Wait()
			log := r.ReadLog()
			assert.Equal(t, `hi`, log)
		})
		t.Run("input", func(t *testing.T) {
			r := runner.NewRunner(runner.RunnerConfig{
				Command: term,
			})
			r.Start()
			defer r.Stop()
			outchan := r.ReadOutput()
			r.WriteInput("echo hi\n")
			assert.Equal(t, "echo hi\n", <-outchan)
			assert.Equal(t, `hi`, <-outchan)
		})
	})
}
