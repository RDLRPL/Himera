package core

import (
	"github.com/RDLxxx/Himera/HGD/browser"
	"github.com/RDLxxx/Himera/HGD/utils"
)

var Monitor, _ = utils.GetPrimaryMonitor()
var Browse = browser.NewBrowser(
	Monitor.Width,
	Monitor.Height,
	"https://polytech.alabuga.ru",
	"browser from scratch btw",
	40.0,
)
