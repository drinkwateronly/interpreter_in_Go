package object

// Environment 为什么不直接使⽤map，⽽是使⽤封装
type Environment struct {
	store map[string]Object // 根据变量名，获取Object，注意，函数也是Object，所以可以赋值给变量
	outer *Environment
}

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil}
}

// NewEnclosedEnvironment 扩展已有的环境
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name) // 当前作用域没有，找外层的作用域
	}
	return obj, ok
}

func (e *Environment) Set(name string, obj Object) Object {
	e.store[name] = obj
	return obj
}
