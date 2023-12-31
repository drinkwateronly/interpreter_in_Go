package evaluator

import (
	"Monkey_1/ast"
	"Monkey_1/object"
	"fmt"
)

// 实例，此后的这些值都是指向这些实例的，无需额外新建实例。
var (
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
	NULL  = &object.Null{}
)

// Eval 输入ast.Node，内部求值，返回一个值的表达 object.Object
func Eval(node ast.Node, env *object.Environment) object.Object {
	// 出现了Eval()的地方都需要判断是否出错
	// node的类型断言
	switch node := node.(type) {
	// AST的根节点
	case *ast.Program:
		// evalStatements 会逐行执行代码，并没有考虑嵌套，导致嵌套遇到return时，会立即返回第一个return
		//return evalStatements(node.Statements)
		return evalProgram(node, env)

	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)

	// 表达式
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}

	case *ast.Boolean:
		// 以下语句每次都创建object.Boolean，实际上只需要true和false的引用即可
		//return &object.Boolean{Value: node.Value}
		if node.Value {
			return TRUE
		}
		return FALSE

	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right // 阻断返回值，否则返回的是，返回值为错误的obj
		}
		return evalPrefixExpression(node.Operator, right)

	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left // 阻断返回值，否则返回的是，返回值为错误的obj
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right // 阻断返回值，否则返回的是，返回值为错误的obj
		}
		return evalInfixExpression(node.Operator, left, right)

	case *ast.BlockStatement: // ？
		//return evalStatements(node.Statements)
		return evalBlockStatement(node, env)

	case *ast.IfExpression:
		return evalIfExpression(node, env)

	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val // 阻断返回值，否则返回的是，返回值为错误的obj
		}
		return &object.ReturnValue{Value: val} // return 终止了Eval的执行

	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)

	case *ast.Identifier:
		return evalIdentifier(node, env)

	case *ast.FunctionLiteral:
		// 简单地将参数列表和函数体赋值
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Env: env, Body: body}

	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			// 只有一个参数 且 该参数是error
			return args[0]
		}
		return applyFunction(function, args)
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}

	case *ast.HashLiteral:
		return evalHashLiteral(node, env)

	case *ast.IndexExpression:
		arrayIdentifier := Eval(node.ArrayIdentifier, env)
		if isError(arrayIdentifier) {
			return arrayIdentifier
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(arrayIdentifier, index)
	}

	return nil
}

// 淘汰 P369 因为可能出现嵌套if都有return时，遇到第一个return就会返回
/* > ⾸先要注意的就是不能通过复⽤ evalStatements函数对块语句求值。需要将它重命名为 evalProgram，降低其通⽤性
func evalStatements(stmts []ast.Statement) object.Object {
	var result object.Object
	for _, statement := range stmts { // ？
		result = Eval(statement)

		returnValue, ok := result.(*object.ReturnValue)
		if ok {
			return returnValue.Value // 是object
		}
	}
	return result
}
*/

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixExpression(right)
	default:
		//return NULL
		return newError("unknown operator: %s%s", operator, right.Type())

	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	// 接受所有的object，无需判断
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default: // 其他值一律为TRUE，因此返回FALSE
		return FALSE
	}
}

func evalMinusPrefixExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ { // 此处应该是要报错？
		//return NULL // 更新：是的
		return newError("unknown operator: -%s", right.Type())
	}
	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	/*
		// 此处与原作者的实现稍有不同，但最终决定与作者保持一致
			if left.Type() != right.Type() {
				return NULL
			}
			switch left.Type() {
			case object.INTEGER_OBJ:
				return evalIntegerInfixExpression(operator, left, right)
			case object.BOOLEAN_OBJ:
				return evalBooleanInfixExpression(operator, left, right)
			default:
				return NULL
			}
	*/
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		// 对 于 *object.Integer，总是有新分配的object.Integer实例，也就是使⽤新的指针。⽽整数不
		// 能通过⽐较不同的实例之间的指针来判断相等性，否则5 == 5将为false。这不是我们期望的⾏为。
		return evalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case operator == "==":
		// 直接对比object本身，其中包括了对比值与类型，这之所以可⾏，是因为程序中⼀直都在使⽤
		// 指向对象的指针，⽽布尔值只有TRUE和FALSE两个对象。 这也适⽤于NULL，但不适用于整数或其他。
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())

	default:
		//return NULL
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())

	}

}

// evalIntegerInfixExpression int的中缀表达式求值，由 evalInfixExpression 调用
func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value
	switch operator {
	case "+":
		return &object.Integer{Value: leftValue + rightValue}
	case "-":
		return &object.Integer{Value: leftValue - rightValue}
	case "*":
		return &object.Integer{Value: leftValue * rightValue}
	case "/":
		return &object.Integer{Value: leftValue / rightValue}
	case "==":
		return nativeBoolToBooleanObject(leftValue == rightValue)
	case "!=":
		return nativeBoolToBooleanObject(leftValue != rightValue)
	case ">":
		return nativeBoolToBooleanObject(leftValue > rightValue)
	case "<":
		return nativeBoolToBooleanObject(leftValue < rightValue)
	default:
		//return NULL
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(operator string, left, right object.Object) object.Object {
	if operator != "+" {
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
	leftValue := left.(*object.String).Value
	rightValue := right.(*object.String).Value
	return &object.String{Value: leftValue + rightValue}
}

/*
在evalInfixExpression 已经实现了

	func evalBooleanInfixExpression(operator string, left, right object.Object) object.Object {
		leftValue := left.(*object.Boolean).Value
		rightValue := right.(*object.Boolean).Value
		switch operator {
		case "==":
			return &object.Boolean{Value: leftValue == rightValue}
		case "!=":
			return &object.Boolean{Value: leftValue != rightValue}
		default:
			return NULL
		}
	}
*/
func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}
	if isTrue(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	} else {
		return NULL
	}
}

// evalProgram P369-P370 较为重要 降低通用性
func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object
	for _, statement := range program.Statements {
		// 对每个statement eval，
		result = Eval(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}
	return result // 作为解释器，通常只返回最后一个语句的求值
}

// ?
func evalBlockStatement(bs *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object
	for _, statement := range bs.Statements {
		result = Eval(statement, env)
		if result != nil {
			resultType := result.Type()
			// 是返回值，或有错误时，立刻返回
			if resultType == object.RETURN_VALUE_OBJ || resultType == object.ERROR_OBJ {
				return result
			}
		}
	}
	return result

}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	// 通过env查找标识符对应值
	if val, ok := env.Get(node.Value); ok {
		return val
	}
	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}
	return newError("identifier not found: " + node.Value)
}

// evalExpressions 对多个表达式求值，返回object列表，用于对函数调用的参数列表求值
func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object
	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

func evalIndexExpression(identifier object.Object, index object.Object) object.Object {
	switch {
	case identifier.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(identifier, index)
	case identifier.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(identifier, index)
	default:
		return newError("index operator not supported: %s", identifier.Type())
	}
}

func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObj := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObj.Elements) - 1)
	if idx < 0 || idx > max {
		return NULL
	}
	return arrayObj.Elements[idx]
}

func evalHashIndexExpression(hash, index object.Object) object.Object {
	hashObject := hash.(*object.Hash)
	key, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}
	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return NULL
	}
	return pair.Value
}

func evalHashLiteral(node *ast.HashLiteral, env *object.Environment) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)
	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}
		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", key.Type())
		}
		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}
		hashed := hashKey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: value}
	}

	return &object.Hash{Pairs: pairs}
}

// applyFunction 根据参数列表args，对函数fn调用求值
func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		// 新建环境，即作用域
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv) // 为什么扩展的是定义函数时的环境，⽽不是当前环境？闭包
		return unwrapReturnValue(evaluated)
	case *object.Builtin:
		return fn.Fn(args...)
	default:
		return newError("not a function: %s", fn.Type())
	}

}

// extendFunctionEnv 新建环境，将参数列表args存入
func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)
	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}
	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	// 如果obj是ReturnValue，则不返回obj，因为外层调用的函数收到ReturnValue会直接返回
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)} // a...
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func nativeBoolToBooleanObject(b bool) object.Object {
	if b {
		return TRUE
	}
	return FALSE
}

func isTrue(condition object.Object) bool {
	// Monkey中，真值为既不是空 又不是false的值，即不一定是true
	switch condition {
	case NULL:
		return false
	case FALSE:
		return false
	//case TRUE:
	//	return true
	default:
		return true
	}
}
