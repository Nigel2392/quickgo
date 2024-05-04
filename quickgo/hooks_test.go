package quickgo_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/Nigel2392/goldcrest"
	"github.com/Nigel2392/quickgo/v2/quickgo"
	"github.com/Nigel2392/quickgo/v2/quickgo/config"
)

type HTTPBuffer struct {
	*bytes.Buffer
	StatusCode int
}

func (b *HTTPBuffer) Header() http.Header {
	return http.Header{}
}

func (b *HTTPBuffer) WriteHeader(i int) {
	b.StatusCode = i
}

func TestAppHook(t *testing.T) {
	var hookName = "quickgo.test.TestAppHook"
	goldcrest.Register(
		hookName, 0,
		func(a *quickgo.App) error { a.Config = &config.QuickGo{Host: hookName}; return nil },
	)
	var a = &quickgo.App{}
	for _, hook := range goldcrest.Get[quickgo.AppHook](hookName) {
		if err := hook(a); err != nil {
			t.Fatal(err)
		}
	}
	if a.Config.Host != hookName {
		t.Fatalf("expected %q, got %q", hookName, a.Config.Host)
	}
}

func TestAppServeHook(t *testing.T) {
	var hookName = "quickgo.test.TestAppServeHook"
	goldcrest.Register(
		hookName, 0,
		func(a *quickgo.App, w http.ResponseWriter, r *http.Request) (written bool, err error) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
			return true, nil
		},
	)
	var (
		a       = &quickgo.App{}
		w       = &HTTPBuffer{new(bytes.Buffer), 0}
		written bool
		err     error
	)
	for _, hook := range goldcrest.Get[quickgo.AppServeHook](hookName) {
		if written, err = hook(a, w, nil); err != nil {
			t.Fatal(err)
		}

		if !written {
			t.Fatal("expected written to be true")
		}
	}
	if w.StatusCode != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, w.StatusCode)
	}

}

func TestAppListProjectsHook(t *testing.T) {
	var hookName = "quickgo.test.TestAppListProjectsHook"
	var (
		err error
		a   = &quickgo.App{}
		p   = make([]string, 0)
	)
	p = append(p, "project1")
	p = append(p, "project2")

	goldcrest.Register(
		hookName, 0,
		func(a *quickgo.App, projects []string) ([]string, error) {
			projects = append(projects, "project3")
			return projects, nil
		},
	)
	for _, hook := range goldcrest.Get[quickgo.AppListProjectsHook](hookName) {
		if p, err = hook(a, p); err != nil {
			t.Fatal(err)
		}
	}

	if len(p) != 3 {
		t.Fatalf("expected 3, got %d", len(p))
	}

	if p[2] != "project3" {
		t.Fatalf("expected %q, got %q", "project3", p[2])
	}
}

func TestProjectHook(t *testing.T) {
	var hookName = "quickgo.test.TestProjectHook"

	goldcrest.Register(
		hookName, 0,
		func(a *quickgo.App, p *config.Project) error {
			p.Name = hookName
			return nil
		},
	)

	var (
		a = &quickgo.App{}
		p = &config.Project{}
	)

	for _, hook := range goldcrest.Get[quickgo.ProjectHook](hookName) {
		if err := hook(a, p); err != nil {
			t.Fatal(err)
		}
	}

	if p.Name != hookName {
		t.Fatalf("expected %q, got %q", hookName, p.Name)
	}
}

func TestProjectWithDirHook(t *testing.T) {
	var hookName = "quickgo.test.TestProjectWithDirHook"
	var (
		err error
		a   = &quickgo.App{}
		p   = &config.Project{}
		d   = "test"
	)

	goldcrest.Register(
		hookName, 0,
		func(a *quickgo.App, p *config.Project, d string) error {
			if d != "test" {
				t.Fatalf("expected %q, got %q", "test", d)
			}
			p.Name = hookName
			return nil
		},
	)

	for _, hook := range goldcrest.Get[quickgo.ProjectWithDirHook](hookName) {
		if err = hook(a, p, d); err != nil {
			t.Fatal(err)
		}
	}
}
