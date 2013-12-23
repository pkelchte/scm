/*
 * A minimal Scheme interpreter, as seen in lis.py and SICP
 * http://norvig.com/lispy.html
 * http://mitpress.mit.edu/sicp/full-text/sicp/book/node77.html
 *
 * Pieter Kelchtermans 2013
 * LICENSE: WTFPL 2.0
 */
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	Repl()
}

/*
 Eval / Apply
*/

func eval(expression scmo, en *env) (value scmo) {
	switch e := expression.(type) {
	case number:
		value = e
	case symbol:
		value = en.Find(e).vars[e]
	case []scmo:
		switch e[0] {
		case symbol("quote"):
			value = e[1]
		case symbol("if"):
			if eval(e[1], en).(bool) {
				value = eval(e[2], en)
			} else {
				value = eval(e[3], en)
			}
		case symbol("set!"):
			v := e[1].(symbol)
			en.Find(v).vars[v] = eval(e[2], en)
			value = "ok"
		case symbol("define"):
			en.vars[e[1].(symbol)] = eval(e[2], en)
			value = "ok"
		case symbol("lambda"):
			value = proc{e[1], e[2], en}
		case symbol("begin"):
			for _, i := range e[1:] {
				value = eval(i, en)
			}
		default:
			operands := e[1:]
			values := make([]scmo, len(operands))
			for i, x := range operands {
				values[i] = eval(x, en)
			}
			value = apply(eval(e[0], en), values)
		}
	default:
		log.Println("Unknown expression type - EVAL", e)
	}
	return
}

func apply(procedure scmo, args []scmo) (value scmo) {
	switch p := procedure.(type) {
	case func(...scmo) scmo:
		value = p(args...)
	case proc:
		en := &env{make(vars), p.en}
		switch params := p.params.(type) {
		case []scmo:
			for i, param := range params {
				en.vars[param.(symbol)] = args[i]
			}
		default:
			en.vars[params.(symbol)] = args
		}
		value = eval(p.body, en)
	default:
		log.Println("Unknown procedure type - APPLY", p)
	}
	return
}

type proc struct {
	params, body scmo
	en           *env
}

/*
 Environments
*/

type vars map[symbol]scmo
type env struct {
	vars
	outer *env
}

func (e *env) Find(s symbol) *env {
	if _, ok := e.vars[s]; ok {
		return e
	} else {
		return e.outer.Find(s)
	}
}

/*
 Primitives
*/

var globalenv env

func init() {
	globalenv = env{
		vars{ //aka an incomplete set of compiled-in functions
			symbol("#t"): true,
			symbol("#f"): false,
			symbol("+"): func(a ...scmo) scmo {
				v := a[0].(number)
				for _, i := range a[1:] {
					v += i.(number)
				}
				return v
			},
			symbol("-"): func(a ...scmo) scmo {
				v := a[0].(number)
				for _, i := range a[1:] {
					v -= i.(number)
				}
				return v
			},
			symbol("*"): func(a ...scmo) scmo {
				v := a[0].(number)
				for _, i := range a[1:] {
					v *= i.(number)
				}
				return v
			},
			symbol("/"): func(a ...scmo) scmo {
				v := a[0].(number)
				for _, i := range a[1:] {
					v /= i.(number)
				}
				return v
			},
			symbol("<="): func(a ...scmo) scmo {
				return a[0].(number) <= a[1].(number)
			},
			symbol("equal?"): func(a ...scmo) scmo {
				return a[0] == a[1]
			},
			symbol("cons"): func(a ...scmo) scmo {
				return []scmo{a[0], a[1]}
			},
			symbol("car"): func(a ...scmo) scmo {
				return a[0].([]scmo)[0]
			},
			symbol("cdr"): func(a ...scmo) scmo {
				return a[0].([]scmo)[1:]
			},
			symbol("list"): eval(read(
				"(lambda z z)"),
				&globalenv),
		},
		nil}
}

/*
 Parsing
*/
type scmo interface{} //scheme objects are e.g. symbols, numbers, expressions, procedures, lists, ...
type symbol string     //symbols are golang strings
type number float64    //constant numbers float64

func read(s string) (expression scmo) {
	tokens := tokenize(s)
	return readFrom(&tokens)
}

//Syntactic Analysis
func readFrom(tokens *[]string) (expression scmo) {
	if len(*tokens) == 0 {
		log.Print("unexpected EOF while reading")
	}
	token := (*tokens)[0]
	//pop first element from tokens
	*tokens = (*tokens)[1:]
	switch token {
	case "(": //a list begins
		L := make([]scmo, 0)
		for (*tokens)[0] != ")" {
			L = append(L, readFrom(tokens))
		}
		*tokens = (*tokens)[1:]
		return L
	case ")":
		log.Print("unexpected )")
		return nil
	default: //an atom occurs
		if f, err := strconv.ParseFloat(token, 64); err == nil {
			return number(f) //numbers become float64
		} else {
			return symbol(token) //others stay string
		}
	}
}

//Lexical Analysis
func tokenize(s string) []string {
	return strings.Split(
		strings.Replace(strings.Replace(s, "(", "( ",
			-1), ")", " )",
			-1), " ")
}

/*
 Interactivity
*/

func String(v scmo) string {
	switch v := v.(type) {
	case []scmo:
		l := make([]string, len(v))
		for i, x := range v {
			l[i] = String(x)
		}
		return "(" + strings.Join(l, " ") + ")"
	default:
		return fmt.Sprint(v)
	}
}

func Repl() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		if input, err := reader.ReadString('\n'); err == nil {
			ans := eval(read(input[:len(input)-1]), &globalenv)
			globalenv.vars[symbol("ans")] = ans
			fmt.Println("==>", String(ans))
		} else {
			fmt.Println("Bye.")
			os.Exit(0)
		}
	}
}
