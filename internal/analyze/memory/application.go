package memory

import "sync"

type Application struct {
}

var ApplicationMap sync.Map

func (a *Application) Update(i interface{}) {
	app := i.(string)
	value, ok := ApplicationMap.Load(app)
	if ok {
		count := value.(int)
		count++
		ApplicationMap.Store(app, count)
	} else {
		ApplicationMap.Store(app, 1)
	}
}
