package bootstrap

import (
	"furryjan/internal/config"
	"furryjan/internal/db"
)

type Data struct {
	Config   *config.Config
	Database *db.DB
}

type CleanupStack struct {
	actions []func()
}

func NewCleanupStack() *CleanupStack {
	return &CleanupStack{actions: make([]func(), 0, 2)}
}

func (c *CleanupStack) Add(action func()) {
	if action == nil {
		return
	}
	c.actions = append(c.actions, action)
}

func (c *CleanupStack) Run() {
	for i := len(c.actions) - 1; i >= 0; i-- {
		c.actions[i]()
	}
}
