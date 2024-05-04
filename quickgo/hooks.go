package quickgo

import (
	"net/http"

	"github.com/Nigel2392/quickgo/v2/quickgo/config"
)

type (
	// quickgo.loaded
	AppHook func(*App) error

	// quickgo.funcs.listProjects
	AppListProjectsHook func(*App, []string) ([]string, error)

	// quickgo.server
	AppServeHook func(*App, http.ResponseWriter, *http.Request) (written bool, err error)

	// quickgo.project.loaded
	// quickgo.project.example
	// quickgo.project.beforeSave
	// quickgo.project.afterSave
	ProjectHook func(*App, *config.Project) error

	// quickgo.project.beforeWrite
	// quickgo.project.afterWrite
	// quickgo.project.beforeLoad
	ProjectWithDirHook func(a *App, proj *config.Project, directory string) error
)

const (
	HookQuickGoLoaded       = "quickgo.loaded"
	HookQuickGoListProjects = "quickgo.funcs.listProjects"
	HookQuickGoServer       = "quickgo.server"
	HookProjectLoaded       = "quickgo.project.loaded"
	HookProjectBeforeLoad   = "quickgo.project.beforeLoad"
	HookProjectExample      = "quickgo.project.example"
	HookProjectBeforeSave   = "quickgo.project.beforeSave"
	HookProjectAfterSave    = "quickgo.project.afterSave"
	HookProjectBeforeWrite  = "quickgo.project.beforeWrite"
	HookProjectAfterWrite   = "quickgo.project.afterWrite"
)
