package controllers

import (
	"github.com/revel/revel"
)

type App struct {
	*revel.Controller
	GormController
}

func (c App) Index() revel.Result {
	return c.Redirect("/docs/index.html")
}

