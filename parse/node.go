// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Parse nodes.
package parse

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

var textFormat = "%s" // Changed to "%q" in tests for better error messages.
// A Node is an element in the parse tree. The interface is trivial.
// The interface contains an unexported method so that only
// types local to this package can satisfy it.
type Node interface {
	Type() NodeType
	String() string
	// Copy does a deep copy of the Node and all its components.
	// To avoid type assertions, some XxxNodes also have specialized
	// CopyXxx methods that return *XxxNode.
	Copy() Node
	Position() Pos // byte position of start of node in full original input string
	// tree returns the containing *Tree.
	// It is unexported so all implementations of Node are in this package.
	tree() *Tree
}

// NodeType identifies the type of a parse tree node.
type NodeType int

// Pos represents a byte position in the original input text from which
// this template was parsed.
type Pos int

func (p Pos) Position() Pos {
	return p
}

// Type returns itself and provides an easy default implementation
// for embedding in a Node. Embedded in all non-trivial Nodes.
func (t NodeType) Type() NodeType {
	return t
}

const (
	NodeText       NodeType = iota // Plain text.
	NodeAction                     // A non-control action such as a field evaluation.
	NodeBool                       // A boolean constant.
	NodeChain                      // A sequence of field accesses.
	NodeCommand                    // An element of a pipeline.
	NodeDot                        // The cursor, dot.
	nodeElse                       // An else action. Not added to tree.
	nodeEnd                        // An end action. Not added to tree.
	NodeField                      // A field or method name.
	NodeIdentifier                 // An identifier; always a function name.
	NodeIf                         // An if action.
	NodeList                       // A list of Nodes.
	NodeNil                        // An untyped nil constant.
	NodeNumber                     // A numerical constant.
	NodePipe                       // A pipeline of commands.
	NodeRange                      // A range action.
	NodeString                     // A string constant.
	NodeTemplate                   // A template invocation action.
	NodeVariable                   // A $ variable.
	NodeWith                       // A with action.
)

// Nodes.
// ListNode holds a sequence of nodes.
type ListNode struct {
	NodeType
	Pos
	tr    *Tree
	Nodes []Node // The element nodes in lexical order.
}

func (t *Tree) newList(pos Pos) *ListNode {
	return &ListNode{tr: t, NodeType: NodeList, Pos: pos}
}
func (l *ListNode) append(n Node) {
	l.Nodes = append(l.Nodes, n)
}
func (l *ListNode) tree() *Tree {
	return l.tr
}
func (l *ListNode) String() string {
	b := new(bytes.Buffer)
	for _, n := range l.Nodes {
		fmt.Fprint(b, n)
	}
	return b.String()
}
func (l *ListNode) CopyList() *ListNode {
	if l == nil {
		return l
	}
	n := l.tr.newList(l.Pos)
	for _, elem := range l.Nodes {
		n.append(elem.Copy())
	}
	return n
}
func (l *ListNode) Copy() Node {
	return l.CopyList()
}

// TextNode holds plain text.
type TextNode struct {
	NodeType
	Pos
	tr   *Tree
	Text []byte // The text; may span newlines.
}

func (t *Tree) newText(pos Pos, text string) *TextNode {
	return &TextNode{tr: t, NodeType: NodeText, Pos: pos, Text: []byte(text)}
}
func (t *TextNode) String() string {
	return fmt.Sprintf(textFormat, t.Text)
}
func (t *TextNode) tree() *Tree {
	return t.tr
}
func (t *TextNode) Copy() Node {
	return &TextNode{tr: t.tr, NodeType: NodeText, Pos: t.Pos, Text: append([]byte{}, t.Text...)}
}

// CommandNode holds a command (a pipeline inside an evaluating action).
type CommandNode struct {
	NodeType
	Pos
	tr   *Tree
	Args []Node // Arguments in lexical order: Identifier, field, or constant.
}

func (t *Tree) newCommand(pos Pos) *CommandNode {
	return &CommandNode{tr: t, NodeType: NodeCommand, Pos: pos}
}
func (c *CommandNode) append(arg Node) {
	c.Args = append(c.Args, arg)
}
func (c *CommandNode) String() string {
	s := ""
	for i, arg := range c.Args {
		if i > 0 {
			s += " "
		}
		s += arg.String()
	}
	return s
}
func (c *CommandNode) tree() *Tree {
	return c.tr
}
func (c *CommandNode) Copy() Node {
	if c == nil {
		return c
	}
	n := c.tr.newCommand(c.Pos)
	for _, c := range c.Args {
		n.append(c.Copy())
	}
	return n
}

// IdentifierNode holds an identifier.
type IdentifierNode struct {
	NodeType
	Pos
	tr    *Tree
	Ident string // The identifier's name.
}

// NewIdentifier returns a new IdentifierNode with the given identifier name.
func NewIdentifier(ident string) *IdentifierNode {
	return &IdentifierNode{NodeType: NodeIdentifier, Ident: ident}
}

// SetPos sets the position. NewIdentifier is a public method so we can't modify its signature.
// Chained for convenience.
// TODO: fix one day?
func (i *IdentifierNode) SetPos(pos Pos) *IdentifierNode {
	i.Pos = pos
	return i
}

// SetTree sets the parent tree for the node. NewIdentifier is a public method so we can't modify its signature.
// Chained for convenience.
// TODO: fix one day?
func (i *IdentifierNode) SetTree(t *Tree) *IdentifierNode {
	i.tr = t
	return i
}
func (i *IdentifierNode) String() string {
	return i.Ident
}
func (i *IdentifierNode) tree() *Tree {
	return i.tr
}
func (i *IdentifierNode) Copy() Node {
	return NewIdentifier(i.Ident).SetTree(i.tr).SetPos(i.Pos)
}

// VariableNode holds a list of variable names, possibly with chained field
// accesses. The dollar sign is part of the (first) name.
type VariableNode struct {
	NodeType
	Pos
	tr    *Tree
	Ident []string // Variable name and fields in lexical order.
}

func (t *Tree) newVariable(pos Pos, ident string) *VariableNode {
	return &VariableNode{tr: t, NodeType: NodeVariable, Pos: pos, Ident: strings.Split(ident, ".")}
}
func (v *VariableNode) String() string {
	s := ""
	for i, id := range v.Ident {
		if i > 0 {
			s += "."
		}
		s += id
	}
	return s
}
func (v *VariableNode) tree() *Tree {
	return v.tr
}
func (v *VariableNode) Copy() Node {
	return &VariableNode{tr: v.tr, NodeType: NodeVariable, Pos: v.Pos, Ident: append([]string{}, v.Ident...)}
}

// DotNode holds the special identifier '.'.
type DotNode struct {
	NodeType
	Pos
	tr *Tree
}

func (t *Tree) newDot(pos Pos) *DotNode {
	return &DotNode{tr: t, NodeType: NodeDot, Pos: pos}
}
func (d *DotNode) Type() NodeType {
	// Override method on embedded NodeType for API compatibility.
	// TODO: Not really a problem; could change API without effect but
	// api tool complains.
	return NodeDot
}
func (d *DotNode) String() string {
	return "."
}
func (d *DotNode) tree() *Tree {
	return d.tr
}
func (d *DotNode) Copy() Node {
	return d.tr.newDot(d.Pos)
}

// NilNode holds the special identifier 'nil' representing an untyped nil constant.
type NilNode struct {
	NodeType
	Pos
	tr *Tree
}

func (t *Tree) newNil(pos Pos) *NilNode {
	return &NilNode{tr: t, NodeType: NodeNil, Pos: pos}
}
func (n *NilNode) Type() NodeType {
	// Override method on embedded NodeType for API compatibility.
	// TODO: Not really a problem; could change API without effect but
	// api tool complains.
	return NodeNil
}
func (n *NilNode) String() string {
	return "nil"
}
func (n *NilNode) tree() *Tree {
	return n.tr
}
func (n *NilNode) Copy() Node {
	return n.tr.newNil(n.Pos)
}

// FieldNode holds a field (identifier starting with '.').
// The names may be chained ('.x.y').
// The period is dropped from each ident.
type FieldNode struct {
	NodeType
	Pos
	tr    *Tree
	Ident []string // The identifiers in lexical order.
}

func (t *Tree) newField(pos Pos, ident string) *FieldNode {
	return &FieldNode{tr: t, NodeType: NodeField, Pos: pos, Ident: strings.Split(ident[1:], ".")} // [1:] to drop leading period
}
func (f *FieldNode) String() string {
	s := ""
	for _, id := range f.Ident {
		s += "." + id
	}
	return s
}
func (f *FieldNode) tree() *Tree {
	return f.tr
}
func (f *FieldNode) Copy() Node {
	return &FieldNode{tr: f.tr, NodeType: NodeField, Pos: f.Pos, Ident: append([]string{}, f.Ident...)}
}

// ChainNode holds a term followed by a chain of field accesses (identifier starting with '.').
// The names may be chained ('.x.y').
// The periods are dropped from each ident.
type ChainNode struct {
	NodeType
	Pos
	tr    *Tree
	Node  Node
	Field []string // The identifiers in lexical order.
}

func (t *Tree) newChain(pos Pos, node Node) *ChainNode {
	return &ChainNode{tr: t, NodeType: NodeChain, Pos: pos, Node: node}
}

// Add adds the named field (which should start with a period) to the end of the chain.
func (c *ChainNode) Add(field string) {
	if len(field) == 0 || field[0] != '.' {
		panic("no dot in field")
	}
	field = field[1:] // Remove leading dot.
	if field == "" {
		panic("empty field")
	}
	c.Field = append(c.Field, field)
}
func (c *ChainNode) String() string {
	s := c.Node.String()
	for _, field := range c.Field {
		s += "." + field
	}
	return s
}
func (c *ChainNode) tree() *Tree {
	return c.tr
}
func (c *ChainNode) Copy() Node {
	return &ChainNode{tr: c.tr, NodeType: NodeChain, Pos: c.Pos, Node: c.Node, Field: append([]string{}, c.Field...)}
}

// BoolNode holds a boolean constant.
type BoolNode struct {
	NodeType
	Pos
	tr   *Tree
	True bool // The value of the boolean constant.
}

func (t *Tree) newBool(pos Pos, true bool) *BoolNode {
	return &BoolNode{tr: t, NodeType: NodeBool, Pos: pos, True: true}
}
func (b *BoolNode) String() string {
	if b.True {
		return "true"
	}
	return "false"
}
func (b *BoolNode) tree() *Tree {
	return b.tr
}
func (b *BoolNode) Copy() Node {
	return b.tr.newBool(b.Pos, b.True)
}

// NumberNode holds a number: signed or unsigned integer, float, or complex.
// The value is parsed and stored under all the types that can represent the value.
// This simulates in a small amount of code the behavior of Go's ideal constants.
type NumberNode struct {
	NodeType
	Pos
	tr      *Tree
	IsInt   bool    // Number has an integral value.
	IsFloat bool    // Number has a floating-point value.
	Int32   int64   // The signed integer value.
	Float64 float64 // The floating-point value.
	Text    string  // The original textual representation from the input.
}

func (t *Tree) newNumber(pos Pos, text string, typ itemType) (*NumberNode, error) {
	n := &NumberNode{tr: t, NodeType: NodeNumber, Pos: pos, Text: text}
	switch typ {
	case itemCharConstant:
		rune, _, tail, err := strconv.UnquoteChar(text[1:], text[0])
		if err != nil {
			return nil, err
		}
		if tail != "'" {
			return nil, fmt.Errorf("malformed character constant: %s", text)
		}
		n.Int32 = int64(rune)
		n.IsInt = true
		n.Float64 = float64(rune) // odd but those are the rules.
		n.IsFloat = true
		return n, nil
	}
	// Do integer test first so we get 0x123 etc.
	i, err := strconv.ParseInt(text, 0, 32)
	if err == nil {
		n.IsInt = true
		n.Int32 = i
	}
	// If an integer extraction succeeded, promote the float.
	if n.IsInt {
		n.IsFloat = true
		n.Float64 = float64(n.Int32)
	} else {
		f, err := strconv.ParseFloat(text, 64)
		if err == nil {
			n.IsFloat = true
			n.Float64 = f
			// If a floating-point extraction succeeded, extract the int if needed.
			if !n.IsInt && float64(int64(f)) == f {
				n.IsInt = true
				n.Int32 = int64(f)
			}
		}
	}
	if !n.IsInt && !n.IsFloat {
		return nil, fmt.Errorf("illegal number syntax: %q", text)
	}
	return n, nil
}

func (n *NumberNode) String() string {
	return n.Text
}
func (n *NumberNode) tree() *Tree {
	return n.tr
}
func (n *NumberNode) Copy() Node {
	nn := new(NumberNode)
	*nn = *n // Easy, fast, correct.
	return nn
}

// StringNode holds a string constant. The value has been "unquoted".
type StringNode struct {
	NodeType
	Pos
	tr     *Tree
	Quoted string // The original text of the string, with quotes.
	Text   string // The string, after quote processing.
}

func (t *Tree) newString(pos Pos, orig, text string) *StringNode {
	return &StringNode{tr: t, NodeType: NodeString, Pos: pos, Quoted: orig, Text: text}
}
func (s *StringNode) String() string {
	return s.Quoted
}
func (s *StringNode) tree() *Tree {
	return s.tr
}
func (s *StringNode) Copy() Node {
	return s.tr.newString(s.Pos, s.Quoted, s.Text)
}

// endNode represents an {{end}} action.
// It does not appear in the final parse tree.
type endNode struct {
	NodeType
	Pos
	tr *Tree
}

func (t *Tree) newEnd(pos Pos) *endNode {
	return &endNode{tr: t, NodeType: nodeEnd, Pos: pos}
}
func (e *endNode) String() string {
	return "{{end}}"
}
func (e *endNode) tree() *Tree {
	return e.tr
}
func (e *endNode) Copy() Node {
	return e.tr.newEnd(e.Pos)
}

// elseNode represents an {{else}} action. Does not appear in the final tree.
type elseNode struct {
	NodeType
	Pos
	tr   *Tree
	Line int // The line number in the input (deprecated; kept for compatibility)
}

func (t *Tree) newElse(pos Pos, line int) *elseNode {
	return &elseNode{tr: t, NodeType: nodeElse, Pos: pos, Line: line}
}
func (e *elseNode) Type() NodeType {
	return nodeElse
}
func (e *elseNode) String() string {
	return "{{else}}"
}
func (e *elseNode) tree() *Tree {
	return e.tr
}
func (e *elseNode) Copy() Node {
	return e.tr.newElse(e.Pos, e.Line)
}

// BranchNode is the common representation of if, range, and with.
type BranchNode struct {
	NodeType
	Pos
	tr       *Tree
	Line     int       // The line number in the input (deprecated; kept for compatibility)
	List     *ListNode // What to execute if the value is non-empty.
	ElseList *ListNode // What to execute if the value is empty (nil if absent).
}

func (b *BranchNode) String() string {
	name := ""
	switch b.NodeType {
	case NodeIf:
		name = "if"
	case NodeRange:
		name = "range"
	case NodeWith:
		name = "with"
	default:
		panic("unknown branch type")
	}
	if b.ElseList != nil {
		return fmt.Sprintf("{{%s}}%s{{else}}%s{{end}}", name, b.List, b.ElseList)
	}
	return fmt.Sprintf("{{%s}}%s{{end}}", name, b.List)
}
func (b *BranchNode) tree() *Tree {
	return b.tr
}
func (b *BranchNode) Copy() Node {
	switch b.NodeType {
	case NodeIf:
		return b.tr.newIf(b.Pos, b.Line, b.List, b.ElseList)
	case NodeRange:
		return b.tr.newRange(b.Pos, b.Line, b.List, b.ElseList)
	case NodeWith:
		return b.tr.newWith(b.Pos, b.Line, b.List, b.ElseList)
	default:
		panic("unknown branch type")
	}
}

// IfNode represents an {{if}} action and its commands.
type IfNode struct {
	BranchNode
}

func (t *Tree) newIf(pos Pos, line int, list, elseList *ListNode) *IfNode {
	return &IfNode{BranchNode{tr: t, NodeType: NodeIf, Pos: pos, Line: line, List: list, ElseList: elseList}}
}
func (i *IfNode) Copy() Node {
	return i.tr.newIf(i.Pos, i.Line, i.List.CopyList(), i.ElseList.CopyList())
}

// RangeNode represents a {{range}} action and its commands.
type RangeNode struct {
	BranchNode
}

func (t *Tree) newRange(pos Pos, line int, list, elseList *ListNode) *RangeNode {
	return &RangeNode{BranchNode{tr: t, NodeType: NodeRange, Pos: pos, Line: line, List: list, ElseList: elseList}}
}
func (r *RangeNode) Copy() Node {
	return r.tr.newRange(r.Pos, r.Line, r.List.CopyList(), r.ElseList.CopyList())
}

// WithNode represents a {{with}} action and its commands.
type WithNode struct {
	BranchNode
}

func (t *Tree) newWith(pos Pos, line int, list, elseList *ListNode) *WithNode {
	return &WithNode{BranchNode{tr: t, NodeType: NodeWith, Pos: pos, Line: line, List: list, ElseList: elseList}}
}
func (w *WithNode) Copy() Node {
	return w.tr.newWith(w.Pos, w.Line, w.List.CopyList(), w.ElseList.CopyList())
}

// TemplateNode represents a {{template}} action.
type TemplateNode struct {
	NodeType
	Pos
	tr   *Tree
	Line int    // The line number in the input (deprecated; kept for compatibility)
	Name string // The name of the template (unquoted).
}

func (t *Tree) newTemplate(pos Pos, line int, name string) *TemplateNode {
	return &TemplateNode{tr: t, NodeType: NodeTemplate, Pos: pos, Line: line, Name: name}
}
func (t *TemplateNode) String() string {
	return fmt.Sprintf("{{template %q}}", t.Name)
}
func (t *TemplateNode) tree() *Tree {
	return t.tr
}
func (t *TemplateNode) Copy() Node {
	return t.tr.newTemplate(t.Pos, t.Line, t.Name)
}
