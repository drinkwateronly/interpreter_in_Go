package evaluator

import (
	"Monkey_1/ast"
	"Monkey_1/object"
	"fmt"
)

var (
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
	NULL  = &object.Null{}
)

// Eval 输入ast.Node，内部求值，返回一个值的表达 object.Object
func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	// 语句
	case *ast.Program:
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
		right := Eval(node.Right, env)
		if isError(left) {
			return left // 阻断返回值，否则返回的是，返回值为错误的obj
		}
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
		return &object.ReturnValue{Value: val}
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)

	case *ast.Identifier:
		return evalIdentifier(node, env)

	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Env: env, Body: body}
	}

	return nil
}

// 淘汰 P369 因为可能出现嵌套if都有return时，遇到第一个return就会返回
/*
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
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
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
		// 对 于 *object.Integer ， 总 是 有 新 分 配 的 object.Integer实例，也就是使⽤新的指针。⽽整数不
		//能通过⽐较不同的实例之间的指针来判断相等性，否则5 1+== 5将为false。这不是我们期望的⾏为。
		return evalIntegerInfixExpression(operator, left, right)
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

func nativeBoolToBooleanObject(b bool) object.Object {
	if b {
		return TRUE
	}
	return FALSE
}

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

func isTrue(condition object.Object) bool {
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

// evalProgram P369-P370 较为重要 降低通用性？
func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object
	for _, statement := range program.Statements {
		result = Eval(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}
	return result
}

// ?
func evalBlockStatement(bs *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object
	for _, statement := range bs.Statements {
		result = Eval(statement, env)
		if result != nil {
			resultType := result.Type()
			if resultType == object.RETURN_VALUE_OBJ || resultType == object.ERROR_OBJ {
				return result
			}
		}
	}
	return result

}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	val, ok := env.Get(node.Value)
	if !ok {
		return newError("identifier not found: " + node.Value)
	}
	return val

}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}
