package runner

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

	goutils "github.com/onichandame/go-utils"
)

type status int

const (
	ready     status = iota
	running   status = iota
	completed status = iota
	stopped   status = iota
	failed    status = iota
)

type Runner struct {
	Error error

	cmd         *exec.Cmd
	status      status
	statuslock  sync.Mutex
	cancel      func()
	stdin       io.WriteCloser
	log         *strings.Builder
	clients     []chan string
	clientslock sync.Mutex
	done        chan int
}

func NewRunner(cfg RunnerConfig) *Runner {
	var r Runner
	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel
	r.cmd = exec.CommandContext(ctx, cfg.Command, cfg.Args...)
	r.cmd.Env = cfg.Env
	r.cmd.Dir = cfg.CWD
	stdin, err := r.cmd.StdinPipe()
	goutils.Assert(err)
	r.stdin = stdin
	r.log = new(strings.Builder)
	r.clients = make([]chan string, 0)
	r.status = ready
	r.done = make(chan int)
	return &r
}

func (r *Runner) Start() {
	r.statuslock.Lock()
	defer r.statuslock.Unlock()
	if r.status != ready {
		panic(fmt.Errorf("cannot restart a runner"))
	}
	stdout, err := r.cmd.StdoutPipe()
	goutils.Assert(err)
	stderr, err := r.cmd.StderrPipe()
	goutils.Assert(err)
	goutils.Assert(r.cmd.Start())
	outscanner := bufio.NewScanner(stdout)
	errscanner := bufio.NewScanner(stderr)
	go func() {
		for outscanner.Scan() {
			line := outscanner.Text()
			r.log.WriteString(line)
			r.broadcast(line)
		}
	}()
	go func() {
		for errscanner.Scan() {
			line := errscanner.Text()
			r.log.WriteString(line)
			r.broadcast(line)
		}
	}()
	go func() {
		err := r.cmd.Wait()
		r.Error = err
		r.statuslock.Lock()
		defer r.statuslock.Unlock()
		if r.status == running {
			if err == nil {
				r.status = completed
			} else {
				fmt.Println(err)
				r.status = failed
			}
		}
		close(r.done)
	}()
	r.status = running
}
func (r *Runner) Stop() {
	r.statuslock.Lock()
	defer r.statuslock.Unlock()
	if r.status != running {
		panic(fmt.Errorf("cannot stop a non-running runner"))
	}
	r.status = stopped
	r.cancel()
	r.clientslock.Lock()
	defer r.clientslock.Unlock()
	for _, client := range r.clients {
		close(client)
	}
	r.clients = make([]chan string, 0)
}
func (r *Runner) Wait() error {
	if r.status < running {
		panic(fmt.Errorf("cannot wait before runner starts"))
	}
	<-r.done
	return r.Error
}
func (r *Runner) WriteInput(inp string) {
	if r.status != running {
		panic(fmt.Errorf("cannot interact with a non-running runner"))
	}
	_, err := io.WriteString(r.stdin, inp)
	r.broadcast(inp)
	goutils.Assert(err)
}
func (r *Runner) ReadOutput() <-chan string {
	out := make(chan string)
	if r.status != stopped {
		r.clientslock.Lock()
		defer r.clientslock.Unlock()
		r.clients = append(r.clients, out)
	}
	go func() {
		log := r.log.String()
		if log != "" {
			out <- r.log.String()
		}
		if r.status == stopped {
			close(out)
		}
	}()
	return out
}

func (r *Runner) GetStatus() string {
	switch r.status {
	case running:
		return `RUNNING`
	case completed:
		return `COMPLETED`
	case failed:
		return `FAILED`
	case ready:
		return `READY`
	case stopped:
		return `STOPPED`
	default:
		return `UNKNOWN`
	}
}

func (r *Runner) broadcast(msg string) {
	for _, client := range r.clients {
		client <- msg
	}
}
