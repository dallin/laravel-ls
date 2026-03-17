package provider

import (
	"github.com/laravel-ls/laravel-ls/cache"
	"github.com/laravel-ls/laravel-ls/project"

	log "github.com/sirupsen/logrus"
)

type InitContext struct {
	Logger    *log.Entry
	RootPath  string
	FileCache *cache.FileCache
	Project   *project.Project
}

type Provider interface {
	Register(manager *Manager)
	Init(ctx InitContext)
}
