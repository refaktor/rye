package ryeco

import (
	"../env"
)

func Loop(n env.Integer, c func() env.Object) env.Object {
	var r env.Object
	for i := 0; int64(i) < n.Value; i++ {
		r = c()
	}
	return r
}

func Add(a env.Integer, b env.Integer) env.Object {
	return env.Integer{a.Value + b.Value}
}
