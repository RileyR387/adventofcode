package main

import (
    "os"
    "bufio"
    "io"
    "fmt"
    "strings"
    "strconv"
)

// An editable operation instruction
type InstructionOp struct {
    operation string
    arg int
    oldOp string
}
func(x InstructionOp) String() string {
    if x.arg < 0 {
        return fmt.Sprintf("%v %v", x.operation, x.arg )
    }
    return fmt.Sprintf("%v +%v", x.operation, x.arg )
}

// An instruction in memory
type Instruction struct {
    op InstructionOp
    evalCount int
}
func(x Instruction) String() string { return fmt.Sprintf("%v evaled: %v", x.op.String(), x.evalCount ) }

type Parser struct {
    code []Instruction
    i int
    errOpIndex int
    accumulation int
    opStack []int
}
func NewParser() *Parser { return &Parser{ []Instruction{}, 0, -1, 0, []int{} } }

func (self *Parser) popStack() {
    // backtrack accumulations and operation edits
    if self.code[self.i].op.operation == "acc" && self.code[self.i].evalCount == 1 {
        self.accumulation = self.accumulation - self.code[self.i].op.arg
    } else if self.code[self.i].op.oldOp != "" {
        self.code[self.i].op.operation = self.code[self.i].op.oldOp
    }
    self.code[self.i].evalCount--
    // pop
    self.i, self.opStack = self.opStack[len(self.opStack)-1], self.opStack[:len(self.opStack)-1]
}

// process the next instruction
func (self *Parser) next() {
    fmt.Printf("%+v accumulation: %v\n",self.code[self.i], self.accumulation)
    fmt.Printf("%v\n",self.opStack)
    switch op := self.code[self.i].op.operation; op {
        case "nop":
            self.opStack = append(self.opStack, self.i)
            self.i++
            break
        case "jmp":
            self.opStack = append(self.opStack, self.i)
            self.i = self.i + self.code[self.i].op.arg
            break
        case "acc":
            self.accumulation = self.accumulation + self.code[self.i].op.arg
            self.opStack = append(self.opStack, self.i)
            self.i++
            break
    }
}

// if we can edit the current instruction edit and return true, else false
func (self *Parser) edit() bool {
    switch op := self.code[self.i].op.operation; op {
        case "nop":
            if self.code[self.i].op.arg != 0 {
                self.code[self.i].op.oldOp = "nop"
                self.code[self.i].op.operation = "jmp"
                self.errOpIndex = self.i
                return true
            } else {
                return false
            }
        case "jmp":
            self.code[self.i].op.oldOp = "jmp"
            self.code[self.i].op.operation = "nop"
            self.errOpIndex = self.i
            return true
        case "acc":
            return false
    }
    return false
}

func (self *Parser) Parse() ( bool, int, int) {
    for {
        if self.i >= len( self.code ) { break }

        self.code[self.i].evalCount++

        if self.code[self.i].evalCount > 1 {
            // errored.. go up the stack
            if self.errOpIndex >= 0 {
                // had a prior instruction edit, go back further up the stack than the last one
                for {
                    if self.i == self.errOpIndex {
                        self.popStack()
                        break
                    }
                    self.popStack()
                }
                self.errOpIndex = -1
            }
            // pop the stack more until we can edit and re-eval
            for {
                if self.edit() {
                    break
                }
                self.popStack()
            }
        }
        // eval next instruction
        self.next()
    }
    return true, self.errOpIndex, self.accumulation
}

func (self *Parser) AddStringOp(operation string) {
    opts := strings.Split(operation, " ")
    instr := opts[0]
    arg, err := strconv.Atoi(opts[1])
    if err != nil { panic("Invalid input") }
    self.code = append(self.code, Instruction{ InstructionOp{ instr, arg, "" }, 0 } )
}

func main() {
    reader := bufio.NewReader(os.Stdin)
    parser := NewParser()
    for {
        str, _, err := reader.ReadLine()
        if err == io.EOF { break }
        line := strings.TrimSpace(string(str))
        if line == "" { break }
        parser.AddStringOp(line)
    }

    if ok, errLine, accum := parser.Parse(); ok {
        fmt.Printf("Found error on line %v with accumulation of: %v\n", errLine, accum )
    } else {
        fmt.Printf("Unknown error\n")
    }
}
