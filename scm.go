/*
 * A minimal Scheme interpreter, as seen in lis.py and SICP
 * Pieter Kelchtermans 2013
 * LICENSE: WTFPL 0.1
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

func eval(e expr, en *env) (res expr) {
	switch e := e.(type) {
	case number:
		res = e
	case symbol:
		res = en.Find(e).vars[e]
	case []expr:
		switch e[0] {
		case symbol("quote"):
			res = e[1]
		case symbol("if"):
			if eval(e[1], en).(bool) {
				res = eval(e[2], en)
			} else {
				res = eval(e[3], en)
			}
		case symbol("set!"):
			v := e[1].(symbol)
			en.Find(v).vars[v] = eval(e[2], en)
			res = "ok"
		case symbol("define"):
			en.vars[e[1].(symbol)] = eval(e[2], en)
			res = "ok"
		case symbol("lambda"):
			params := make([]symbol, 0)
			for _, p := range e[1].([]expr) {
				params = append(params,
					p.(symbol))
			}
			res = proc{params, e[2], en}
		case symbol("begin"):
			for _, i := range e[1:] {
				res = eval(i, en)
			}
		default:
			values := make([]number, 0)
			for _, i := range e[1:] {
				values = append(values,
					eval(i, en).(number))
			}
			res = apply(eval(e[0], en), values)
		}
	default:
		log.Println("Unknown expression type - EVAL", e)
	}
	return
}

func apply(p expr, args expr) (res expr) {
	switch p := p.(type) {
	case pNumeric:
		res = p(args.([]number)...)
	case pBoolean:
		res = p(args.([]number)...)
	case proc:
		en := new(env)
		en.vars = make(map[symbol]expr)
		en.outer = p.en
		for i := range p.parameters {
			en.vars[p.parameters[i]] = args.([]number)[i]
		}
		res = eval(p.body, en)
	default:
		log.Println("Unknown procedure type - APPLY", p)
	}
	return
}

type proc struct {
	parameters []symbol
	body       expr
	en         *env
}

/*
 Environments
*/

type vars map[symbol]expr
type env struct {
	vars
	outer *env
}

func (e env) Find(s symbol) env {
	if _, ok := e.vars[s]; ok {
		return e
	} else {
		return e.outer.Find(s)
	}
}

/*
 Primitives
*/

type pNumeric func(...number) number
type pBoolean func(...number) bool

var globalenv = env{
	vars{
		symbol("#t"): true,
		symbol("#f"): false,
		symbol("+"): pNumeric(func(a ...number) (v number) {
			v = a[0]
			for _, i := range a[1:] {
				v += i
			}
			return
		}),
		symbol("-"): pNumeric(func(a ...number) (v number) {
			v = a[0]
			for _, i := range a[1:] {
				v -= i
			}
			return
		}),
		symbol("*"): pNumeric(func(a ...number) (v number) {
			v = a[0]
			for _, i := range a[1:] {
				v *= i
			}
			return
		}),
		symbol("/"): pNumeric(func(a ...number) (v number) {
			v = a[0]
			for _, i := range a[1:] {
				v /= i
			}
			return
		}),
		symbol("<="): pBoolean(func(a ...number) (v bool) {
			return a[0] <= a[1]
		})},

	nil}

/*
 Parsing
*/

type expr interface{} //expressions can be anything
type symbol string    //symbols are golang strings
type number float64   //Constant numbers float64

func read(s string) expr {
	tokens := tokenize(s)
	return readFrom(&tokens)
}

//Syntactic Analysis
func readFrom(tokens *[]string) expr {
	if len(*tokens) == 0 {
		log.Print("unexpected EOF while reading")
	}
	//pop first element from tokens
	token := (*tokens)[0]
	*tokens = (*tokens)[1:]
	switch token {
	case "(": //a list begins
		L := make([]expr, 0)
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

func Repl() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		if input, err := reader.ReadString('\n'); err == nil {
			fmt.Println("==>", eval(read(input[:len(input)-1]), &globalenv))
		} else {
			fmt.Println("Bye.")
			os.Exit(0)
		}
	}
}
