// Package evm is an **incomplete** implementation of the Ethereum Virtual
// Machine for the "EVM From Scratch" course:
// https://github.com/w1nt3r-eth/evm-from-scratch
//
// To work on EVM From Scratch In Go:
//
// - Install Golang: https://golang.org/doc/install
// - Go to the `go` directory: `cd go`
// - Edit `evm.go` (this file!), see TODO below
// - Run `go test ./...` to run the tests
package evm

import (
	"math/big"
)

// Run runs the EVM code and returns the stack and a success indicator.
func Evm(code []byte) ([]*big.Int, bool) {
	var stack []*big.Int
	pc := 0

	for pc < len(code) {
		op := code[pc]
		pc++

		// TODO: Implement the EVM here!
		switch op {
		case 0x00:
			return stack, true

		case 0x5F:
			value := big.NewInt(0)
			stack = append([]*big.Int{value}, stack...)

		case 0x60:
			if pc+1 > len(code) {
				return nil, false // Error: not enough bytes left for PUSH1
			}
			value := new(big.Int).SetBytes(code[pc : pc+1])
			stack = append([]*big.Int{value}, stack...)
			pc++

		case 0x61:
			if pc+2 > len(code) {
				return nil, false // Error: not enough bytes left for PUSH2
			}
			value := new(big.Int).SetBytes(code[pc : pc+2])
			stack = append([]*big.Int{value}, stack...)
			pc += 2
		case 0x63:
			if pc+4 > len(code) {
				return nil, false // Error: not enough bytes left for PUSH4
			}
			value := new(big.Int).SetBytes(code[pc : pc+4])
			stack = append([]*big.Int{value}, stack...)
			pc += 4
		case 0x65:
			if pc+6 > len(code) {
				return nil, false // Error: not enough bytes left for PUSH6
			}
			value := new(big.Int).SetBytes(code[pc : 6+pc])
			stack = append([]*big.Int{value}, stack...)
			pc += 6

		case 0x69:
			if pc+10 > len(code) {
				return nil, false
			}
			value := new(big.Int).SetBytes(code[pc : 10+pc])
			stack = append([]*big.Int{value}, stack...)
			pc += 10
		case 0x6A:
			if pc+11 > len(code) {
				return nil, false
			}
			value := new(big.Int).SetBytes(code[pc : 11+pc])
			stack = append([]*big.Int{value}, stack...)
			pc += 11
		case 0x7F:
			if pc+32 > len(code) {
				return nil, false
			}
			value := new(big.Int).SetBytes(code[pc : 32+pc])
			stack = append([]*big.Int{value}, stack...)
			pc += 32
		case 0x50:
			if len(stack) < 1 {
				return nil, false
			}

			stack = stack[1:]

		case 0x01:
			if len(stack) < 2 {
				return nil, false
			}

			value := new(big.Int).Add(stack[0], stack[1])
			value.Mod(value, new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil)) // Modulo operation to handle overflow
			stack = stack[2:]
			stack = append([]*big.Int{value}, stack...)

		case 0x02:
			if len(stack) < 2 {
				return nil, false
			}

			value := new(big.Int).Mul(stack[0], stack[1])
			value.Mod(value, new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil)) // Modulo operation to handle overflow
			stack = stack[2:]
			stack = append([]*big.Int{value}, stack...)
		case 0x03:
			if len(stack) < 2 {
				return nil, false
			}

			value := new(big.Int).Sub(stack[0], stack[1])
			value.Mod(value, new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil)) // Modulo operation to handle overflow`
			stack = stack[2:]
			stack = append([]*big.Int{value}, stack...)

		case 0x04:
			if len(stack) < 2 {
				return nil, false
			}
			if stack[1].Cmp(big.NewInt(0)) == 0 {
				stack = stack[2:]
				stack = append([]*big.Int{big.NewInt(0)}, stack...)
			} else {
				value := new(big.Int).Div(stack[0], stack[1])
				value.Mod(value, new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil)) // Modulo operation to handle overflow
				stack = stack[2:]
				stack = append([]*big.Int{value}, stack...)
			}
		case 0x06:
			if len(stack) < 2 {
				return nil, false
			}
			if stack[1].Cmp(big.NewInt(0)) == 0 {
				stack = stack[2:]
				stack = append([]*big.Int{big.NewInt(0)}, stack...)
			} else {
				value := new(big.Int).Mod(stack[0], stack[1])
				value.Mod(value, new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil)) // Modulo operation to handle overflow
				stack = stack[2:]
				stack = append([]*big.Int{value}, stack...)
			}
		case 0x08:
			if len(stack) < 3 {
				return nil, false
			}

			valueMod := new(big.Int).Add(stack[0], stack[1])
			stack = stack[2:]
			value := new(big.Int).Mod(valueMod, stack[0])
			stack = stack[1:]
			value.Mod(value, new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil))
			stack = append([]*big.Int{value}, stack...)
		case 0x09:
			if len(stack) < 3 {
				return nil, false
			}

			valueMod := new(big.Int).Mul(stack[0], stack[1])
			stack = stack[2:]
			value := new(big.Int).Mod(valueMod, stack[0])
			stack = stack[1:]
			value.Mod(value, new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil))
			stack = append([]*big.Int{value}, stack...)
		case 0x0A:
			if len(stack) < 2 {
				return nil, false
			}

			value := new(big.Int).Exp(stack[0], stack[1], nil)
			stack = stack[2:]

			value.Mod(value, new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil))
			stack = append([]*big.Int{value}, stack...)
		case 0x0B:
			if len(stack) < 2 {
				return nil, false
			}

			// Pop b from the stack

			// Pop x from the stack

			// Ensure b is within the bounds of our bit width (0-31 for a 256-bit number)

			// Pop b and x from the stack
			b := stack[0]
			x := stack[1]
			stack = stack[2:]

			// Calculate the sign extension mask
			bInt := int(b.Int64())
			if bInt >= 32 {
				return stack, false
			}
			bits := (bInt + 1) * 8
			signBit := new(big.Int).Lsh(big.NewInt(1), uint(bits-1))

			// Check if the sign bit is set
			if x.Cmp(signBit) >= 0 {
				// If the sign bit is set, extend with 1s
				extended := new(big.Int).Lsh(big.NewInt(1), uint(256-bits))
				extended.Sub(extended, big.NewInt(1))
				extended.Lsh(extended, uint(bits))
				x.Or(x, extended)
			} else {
				// Ensure higher bits are zero
				mask := new(big.Int).Lsh(big.NewInt(1), uint(bits))
				mask.Sub(mask, big.NewInt(1))
				x.And(x, mask)
			}

			// Push the result back onto the stack
			stack = append([]*big.Int{x}, stack...)

		case 0x05:
			if len(stack) < 2 {
				return nil, false
			}

			if stack[1].Cmp(big.NewInt(0)) == 0 {
				stack = stack[2:]
				stack = append([]*big.Int{big.NewInt(0)}, stack...)
			} else {
				value1 := stack[0].Int64()
				int8Value1 := int8(value1)
				value2 := stack[1].Int64()
				int8Value2 := int8(value2)

				value := int8Value1 / int8Value2

				bits := 8

				// Check if the sign bit is set
				if value < 0 {
					value8 := new(big.Int).Add(big.NewInt(int64(256)), big.NewInt(int64(value)))
					// If the sign bit s set, extend with 1s
					extended := new(big.Int).Lsh(big.NewInt(1), uint(256-bits))
					extended.Sub(extended, big.NewInt(1))
					extended.Lsh(extended, uint(bits))
					value8.Or(value8, extended)
					stack = stack[2:]
					stack = append([]*big.Int{value8}, stack...)
				} else {
					stack = stack[2:]
					stack = append([]*big.Int{big.NewInt(int64(value))}, stack...)
				}

			}
		case 0x07:
			if len(stack) < 2 {
				return nil, false
			}

			if stack[1].Cmp(big.NewInt(0)) == 0 {
				stack = stack[2:]
				stack = append([]*big.Int{big.NewInt(0)}, stack...)
			} else {
				value1 := stack[0].Int64()
				int8Value1 := int8(value1)
				value2 := stack[1].Int64()
				int8Value2 := int8(value2)

				value := int8Value1 % int8Value2

				bits := 8

				// Check if the sign bit is set
				if value < 0 {
					value8 := new(big.Int).Add(big.NewInt(int64(256)), big.NewInt(int64(value)))
					// If the sign bit s set, extend with 1s
					extended := new(big.Int).Lsh(big.NewInt(1), uint(256-bits))
					extended.Sub(extended, big.NewInt(1))
					extended.Lsh(extended, uint(bits))
					value8.Or(value8, extended)
					stack = stack[2:]
					stack = append([]*big.Int{value8}, stack...)
				} else {
					stack = stack[2:]
					stack = append([]*big.Int{big.NewInt(int64(value))}, stack...)
				}

			}

		case 0x10:
			if len(stack) < 2 {
				return nil, false
			}

			if stack[0].Cmp(stack[1]) == 0 || stack[0].Cmp(stack[1]) == 1 {
				stack = stack[2:]
				stack = append([]*big.Int{big.NewInt(0)}, stack...)
			} else {
				stack = stack[2:]
				stack = append([]*big.Int{big.NewInt(1)}, stack...)
			}
		case 0x11:
			if len(stack) < 2 {
				return nil, false
			}

			if stack[0].Cmp(stack[1]) == 0 || stack[0].Cmp(stack[1]) == -1 {
				stack = stack[2:]
				stack = append([]*big.Int{big.NewInt(0)}, stack...)
			} else {
				stack = stack[2:]
				stack = append([]*big.Int{big.NewInt(1)}, stack...)
			}
		case 0x12:
			if len(stack) < 2 {
				return nil, false
			}

			value1 := stack[0].Int64()
			int8Value1 := int8(value1)
			value2 := stack[1].Int64()
			int8Value2 := int8(value2)

			if int8Value1 < int8Value2 {
				value := 1
				stack = stack[2:]
				stack = append([]*big.Int{big.NewInt(int64(value))}, stack...)
			} else {
				value := 0
				stack = stack[2:]
				stack = append([]*big.Int{big.NewInt(int64(value))}, stack...)
			}
		case 0x13:
			if len(stack) < 2 {
				return nil, false
			}

			value1 := stack[0].Int64()
			int8Value1 := int8(value1)
			value2 := stack[1].Int64()
			int8Value2 := int8(value2)

			if int8Value1 > int8Value2 {
				value := 1
				stack = stack[2:]
				stack = append([]*big.Int{big.NewInt(int64(value))}, stack...)
			} else {
				value := 0
				stack = stack[2:]
				stack = append([]*big.Int{big.NewInt(int64(value))}, stack...)
			}
		case 0x14:
			if len(stack) < 2 {
				return nil, false
			}

			if stack[0].Cmp(stack[1]) == 0 {
				stack = stack[2:]
				stack = append([]*big.Int{big.NewInt(1)}, stack...)
			} else {
				stack = stack[2:]
				stack = append([]*big.Int{big.NewInt(0)}, stack...)
			}

		case 0x15:

			if stack[0].Cmp(big.NewInt(0)) == 0 {
				stack = stack[1:]
				stack = append([]*big.Int{big.NewInt(1)}, stack...)
			} else {
				stack = stack[1:]
				stack = append([]*big.Int{big.NewInt(0)}, stack...)
			}

		case 0x19:
			UINT256Max := new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(1))
			value := new(big.Int).Xor(UINT256Max, stack[0])

			stack = stack[1:]
			stack = append([]*big.Int{value}, stack...)

			// Check if the sign bit is

		case 0x16:
			value := new(big.Int).And(stack[0], stack[1])

			stack = stack[2:]
			stack = append([]*big.Int{value}, stack...)
		case 0x17:
			value := new(big.Int).Or(stack[0], stack[1])

			stack = stack[2:]
			stack = append([]*big.Int{value}, stack...)
		case 0x18:
			value := new(big.Int).Xor(stack[0], stack[1])

			stack = stack[2:]
			stack = append([]*big.Int{value}, stack...)

		case 0x1B:
			if len(stack) < 2 {
				return nil, false
			}

			extended := stack[1]
			UINT256Max := new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(1))

			if stack[0].Cmp(big.NewInt(255)) > 1 {
				extended = big.NewInt(0)
			} else {
				extended = new(big.Int).Lsh(stack[1], uint(stack[0].Int64()))
				extended.And(extended, UINT256Max)
			}

			stack = stack[2:]
			stack = append([]*big.Int{extended}, stack...)

		case 0x1C:
			if len(stack) < 2 {
				return nil, false
			}

			extended := stack[1]
			UINT256Max := new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(1))

			if stack[0].Cmp(big.NewInt(255)) > 1 {
				extended = big.NewInt(0)
			} else {
				extended = new(big.Int).Rsh(stack[1], uint(stack[0].Int64()))
				extended.And(extended, UINT256Max)
			}

			stack = stack[2:]
			stack = append([]*big.Int{extended}, stack...)

		}

	}
	return stack, true
}
