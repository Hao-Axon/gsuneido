// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

// Package opcodes defines the bytecode instructions
// generated by compiler and executed by runtime
package opcodes

//go:generate stringer -type=Opcode

// to make stringer: go generate

// Where applicable there are matching methods in thread.go or ops.go
// e.g. t.Pop or ops.Add

type Opcode byte

const (
	Nop Opcode = iota

	// stack --------------------------------------------------------

	// Pop pops the top of the stack
	Pop
	// Dup duplicates the top of the stack
	Dup
	// Dup2 duplicates the top two values on the stack i.e. a,b => a,b,a,b
	Dup2
	// Dupx2 duplicates the top under the next two, used for post inc/dec
	Dupx2

	// push values --------------------------------------------------

	// Int <int16> pushes an integer
	Int
	// Value <uint8> pushes a literal Value
	Value
	// True pushes True
	True
	// False pushes False
	False
	// Zero pushes Zero (0)
	Zero
	// One pushes One (1)
	One
	// MaxInt pushes int32 max, used by RangeTo and RangeLen
	MaxInt
	// EmptyStr pushes EmptyStr ("")
	EmptyStr

	// load and store -----------------------------------------------

	// Load <uint8> pushes a local variable onto the stack
	Load
	// Store <uint8> pops the top value off the stack into a local variable
	Store
	// Dyload <uint8> pushes a dynamic variable onto the stack
	// It looks up the frame stack to find it, and copies it locally
	Dyload
	// Global <uint16> pushes the value of a global name
	Global
	// Get replaces the top two values (ob & mem) with ob.Get(mem)
	Get
	// Put pops the top three values (ob, mem, val) and does ob.Put(mem, val)
	Put
	// RangeTo replaces the top three values (x,i,j) with x.RangeTo(i,j)
	RangeTo
	// RangeLen replaces the top three values (x,i,n) with x.RangeLen(i,n)
	RangeLen
	// This pushes frame.this
	This

	// operations ---------------------------------------------------

	// Is replaces the top two values with x is y
	Is
	// Isnt replaces the top two values with x isnt y
	Isnt
	// Match replaces the top two values with x =~ y
	Match
	// MatchNot replaces the top two values with x !~ y
	MatchNot
	// Lt replaces the top two values with x < y
	Lt
	// Lte replaces the top two values with x <= y
	Lte
	// Gt replaces the top two values with x > y
	Gt
	// Gte replaces the top two values with x >= y
	Gte
	// Add replaces the top two values with x + y
	Add
	// Sub replaces the top two values with x - y
	Sub
	// Cat replaces the top two values with x $ y (strings)
	Cat
	// Mul replaces the top two values with x * y
	Mul
	// Div replaces the top two values with x / y
	Div
	// Mod replaces the top two values with x % y
	Mod
	// LeftShift replaces the top two values with x << y (integers)
	LeftShift
	// RightShift replaces the top two values with x >> y (unsigned)
	RightShift
	// BitOr replaces the top two values with x | y (integers)
	BitOr
	// BitAnd replaces the top two values with x | y (integers)
	BitAnd
	// BitXor replaces the top two values with x & y (integers)
	BitXor
	// BitNot replaces the top value with ^y (integer)
	BitNot
	// Not replaces the top value with not x (logical)
	Not
	// UnaryPlus converts the top value to a number
	UnaryPlus
	// UnaryMinus replaces the top value with -x
	UnaryMinus

	// control flow -------------------------------------------------

	// Or <int16> jumps if top is true, else it pops and continues
	// panics if top is not True or false
	Or
	// And <int16> jumps if top is false, else it pops and continues
	// panics if top is not True or false
	And
	// Bool checks that top is True or False, else it panics
	Bool
	// QMark pops and if false jumps, else it continues
	// panics if top is not True or false
	QMark
	// In <int16> pops the top value and compares it to the next value
	// if equal it pops the second value and pushes True,
	// else it leaves the second value on the stack
	In
	// Jump <int16> jumps to a relative location in the code
	Jump
	// JumpTrue <int16> pops and if true jumps, else it continues
	// panics if top is not True or False
	JumpTrue
	// JumpFalse <int16> pops and if false jumps, else it continues
	// panics if top is not True or False
	JumpFalse
	// JumpIs <int16> pops the top value and compares it to the next value
	// if equal it pops the second value and jumps
	// else it leaves the second value on the stack and continues
	// panics if top is not True or False
	JumpIs
	// JumpIsnt <int16> pops the top value and compares it to the next value
	// if not equal it leaves the second value on the stack
	// else it pops the second value on the stack and continues
	// panics if top is not True or False
	JumpIsnt
	// Iter replaces the top with top.Iter()
	Iter
	// ForIn <int16> <uint8> calls top.Next()
	// if the result is equal to top, it jumps
	// else it continues
	ForIn

	// exceptions ---------------------------------------------------

	// Throw pops and panics with that value
	Throw
	// Try <int16> <uint8> registers the catch jump and the catch pattern
	// so we will start catching
	Try
	// Catch <int16> clears the catch information to stop catching
	// and jumps past the catch code
	Catch

	// call and return ----------------------------------------------

	// CallFuncDiscard <uint8> calls the function popped from the stack
	// with the specified StdArgSpecs or frame.fn.ArgsSpecs
	// and discards the result
	CallFuncDiscard

	// CallFuncNoNil <uint8> calls the function popped from the stack
	// with the specified StdArgSpecs or frame.fn.ArgsSpecs
	// and pushes the result which must not be nil
	CallFuncNoNil

	// CallFuncNilOk<uint8> calls the function popped from the stack
	// with the specified StdArgSpecs or frame.fn.ArgsSpecs
	// and pushes the result which may be nil (return special case)
	CallFuncNilOk

	// CallMethDiscard <uint8> calls the method popped from the stack
	// with the specified StdArgSpecs or frame.fn.ArgsSpecs
	// and discards the result
	CallMethDiscard

	// CallMethNoNil <uint8> calls the method popped from the stack
	// with the specified StdArgSpecs or frame.fn.ArgsSpecs
	// and pushes the result which must not be nil
	CallMethNoNil

	// CallMethNilOk <uint8> calls the method popped from the stack
	// with the specified StdArgSpecs or frame.fn.ArgsSpecs
	// and pushes the result which may be nil (return special case)
	CallMethNilOk

	// Super <uint16> specifies where to start the method lookup
	// for the following CallMeth
	Super
	// Return returns the top of the stack
	Return
	// ReturnNil returns nil i.e. no return value
	ReturnNil

	// blocks -------------------------------------------------------

	// Block <uint8> pushes a new block instance
	Block
	// BlockBreak panics "block:break" (handled by application code)
	BlockBreak
	// BlockContinue panics "block:continue" (handled by application code)
	BlockContinue
	// BlockReturn panics "block return" (handled by runtime)
	BlockReturn
	// BlockReturnNil pushes nil and then does BlockReturn
	BlockReturnNil
)
