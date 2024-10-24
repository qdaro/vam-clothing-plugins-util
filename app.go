package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"app/lib"

	"github.com/adrg/xdg"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx          context.Context
	windowStates *lib.WindowStateStore
	configStore  *lib.ConfigStore
	config       *lib.AppConfig
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Create & load config files
	a.config = &lib.AppConfig{}
	a.configStore = lib.NewConfigStore(a.getConfigPath("config"))
	a.windowStates = lib.NewWindowStateStore(a.getConfigPath("windows"))
	lib.Check(a.configStore.Load(a.config))
	lib.Check(a.windowStates.Load())
	a.applyConfig()

	// Restore main window position/size
	mainWindow, ok := a.windowStates.Get("main")
	if !ok {
		mainWindow = a.windowStates.Set("main", a.getWindowState())
	}
	runtime.WindowSetPosition(a.ctx, mainWindow.X, mainWindow.Y)
	runtime.WindowSetSize(a.ctx, mainWindow.Width, mainWindow.Height)
}

// Greet returns a greeting for the given name
func (a *App) onSecondInstanceLaunch(secondInstanceData options.SecondInstanceData) {
	runtime.WindowShow(a.ctx)
}

func (a *App) getWindowState() *lib.WindowState {
	x, y := runtime.WindowGetPosition(a.ctx)
	w, h := runtime.WindowGetSize(a.ctx)
	return &lib.WindowState{X: x, Y: y, Width: w, Height: h}
}

func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	a.configStore.Save(a.config)
	// Save window state when window is not fullscreen/minimized/maximized.
	// Ideally this should happen after window move/resize events, but there
	// is no way to tap into position in wails atm :(
	if runtime.WindowIsNormal(a.ctx) {
		a.windowStates.Set("main", a.getWindowState())
		lib.Check(a.windowStates.Save())
	}
	return false
}

func (a *App) getConfigPath(name string) string {
	return lib.Must(xdg.ConfigFile(filepath.Join("Clothing Plugins Util", name+".json")))
}

func (a *App) message(message *lib.Message) {
	runtime.EventsEmit(a.ctx, "message", message)
}

func (a *App) GetConfig() *lib.AppConfig {
	return a.config
}

func (a *App) SetConfig(config *lib.AppConfig) error {
	oldConfig := a.config
	a.config = config

	err := a.configStore.Save(a.config)
	if err != nil {
		a.config = oldConfig
		return err
	}

	a.applyConfig()
	runtime.EventsEmit(a.ctx, "config", a.config)
	return err
}

func (a *App) applyConfig() {
	runtime.WindowSetAlwaysOnTop(a.ctx, a.config.OnTop)
}

var clothingVajExp = regexp.MustCompile(`(?i).*/custom/clothing/(?:female|male)/[^/]+/[^/]+/.*\.vaj$`)
var clothingVapExps = []*regexp.Regexp{
	regexp.MustCompile(`(?i).*/custom/clothing/(?:female|male)/[^/]+/[^/]+/.*\.vap$`),
	regexp.MustCompile(`(?i).*/custom/atom/person/appearance/.*\.vap$`),
	regexp.MustCompile(`(?i).*/custom/atom/person/clothing/.*\.vap$`),
}
var clothingCplExp = regexp.MustCompile(`(?i).*/custom/clothing/(?:female|male)/[^/]+/[^/]+/.*\.clothingplugins$`)

// Initializes Clothing Plugin Manager in .vaj files
func (a *App) InitPaths(paths []string) {
	for _, path := range paths {
		err := filepath.Walk(path, func(walkedPath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			normalizedPath := filepath.ToSlash(walkedPath)

			if clothingVajExp.MatchString(normalizedPath) {
				a.message(lib.FixVaj(normalizedPath, false))
			}

			return nil
		})
		if err != nil {
			a.message(&lib.Message{Icon: lib.Ptr("file"), Title: path, Notes: []lib.Note{{
				Variant: "danger",
				Text:    fmt.Sprintf("Failed to walk path \"%s\".", path),
				Details: lib.Ptr(err.Error()),
			}}})
		}
	}
}

// Fixes .vaj. clothingplugins, and .vap files for release
func (a *App) FixPaths(paths []string) {
	for _, path := range paths {
		err := filepath.Walk(path, func(walkedPath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			normalizedPath := filepath.ToSlash(walkedPath)

			if clothingVajExp.MatchString(normalizedPath) {
				a.message(lib.FixVaj(normalizedPath, true))
			} else if clothingCplExp.MatchString(normalizedPath) {
				a.message(lib.FixCpl(normalizedPath))
			} else {
				for _, exp := range clothingVapExps {
					if exp.MatchString(normalizedPath) {
						a.message(lib.FixVap(normalizedPath))
						break
					}
				}
			}

			return nil
		})
		if err != nil {
			a.message(&lib.Message{Icon: lib.Ptr("file"), Title: path, Notes: []lib.Note{{
				Variant: "danger",
				Text:    fmt.Sprintf("Failed to walk path \"%s\".", path),
				Details: lib.Ptr(err.Error()),
			}}})
		}
	}
}

// Dummy method to force wails to generate bindings for message types.
func (a *App) Dummy() lib.Message {
	return lib.Message{}
}
