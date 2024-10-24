package lib

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
)

type ConfigStore struct {
	Path string
}

func NewConfigStore(path string) *ConfigStore {
	return &ConfigStore{Path: path}
}

func (c *ConfigStore) Load(v interface{}) error {
	data, err := os.ReadFile(c.Path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}

	err = json.Unmarshal(data, v)
	if err != nil {
		return err
	}

	return nil
}

func (c *ConfigStore) Save(v interface{}) error {
	data, err := JSONMarshalPretty(v)
	if err != nil {
		return err
	}

	err = os.WriteFile(c.Path, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// Storing window states
type WindowState struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type WindowStateStore struct {
	store  *ConfigStore
	states map[string]*WindowState
}

func NewWindowStateStore(path string) *WindowStateStore {
	return &WindowStateStore{
		store:  NewConfigStore(path),
		states: map[string]*WindowState{},
	}
}

func (ws *WindowStateStore) Get(name string) (state *WindowState, ok bool) {
	state, ok = ws.states[name]
	return state, ok
}

func (ws *WindowStateStore) Set(name string, state *WindowState) *WindowState {
	ws.states[name] = state
	return state
}

func (ws *WindowStateStore) Load() error {
	err := ws.store.Load(&ws.states)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

func (ws *WindowStateStore) Save() error {
	return ws.store.Save(ws.states)
}
