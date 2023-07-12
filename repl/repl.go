package repl

import (
	"Monkey_1/lexer"
	"Monkey_1/parser"
	"bufio"
	"fmt"
	"io"
)

const PROMPT = ">> "

const BRAND = ` __  __  ___  _   _ _  _________   __ 
|  \/  |/ _ \| \ | | |/ / ____\ \ / /
| |\/| | | | |  \| | ' /|  _|  \ V /
| |  | | |_| | |\  | . \| |___  | |
|_|  |_|\___/|_| \_|_|\_\_____| |_|  
`

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Fprintf(out, PROMPT)
		scanned := scanner.Scan() // 从 in 读入下一行 ，并移除行末的换行符
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			printParseErrors(out, p.Errors())
			continue
		}
		io.WriteString(out, program.String())
		io.WriteString(out, "\n")
		//for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
		//	fmt.Fprintf(out, "%+v\n", tok)
		//}
	}
}

func printParseErrors(out io.Writer, errors []string) {
	io.WriteString(out, BRAND)
	io.WriteString(out, "Woops! We ran into some monkey business here!\n parser errors:")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
