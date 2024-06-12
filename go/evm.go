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

func jumpdest(pc int, code []byte, stack []*big.Int) (int, []*big.Int, bool) {
	stack = []*big.Int{}
loop:
	for i := pc; i < len(code); i++ {
		if code[i] == 0x5B {

			switch code[i-1] {
			case 0x60, 0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6A, 0x6B, 0x6C, 0x6D, 0x6E, 0x6F, 0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x76, 0x77, 0x78, 0x79, 0x7A, 0x7B, 0x7C, 0x7D, 0x7E, 0x7F:

			default:
				pc = i
				break loop
			}
		}
		if i == len(code)-1 && code[len(code)-1] != 0x5B {
			return pc, stack, false
		}

	}
	return pc, stack, true
}

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
		case 0x62:
			if pc+3 > len(code) {
				return nil, false // Error: not enough bytes left for PUSH3
			}
			value := new(big.Int).SetBytes(code[pc : pc+3])
			stack = append([]*big.Int{value}, stack...)
			pc += 3
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

		case 0x1D:
			if len(stack) < 2 {
				return nil, false
			}
			extended := stack[1]
			UINT256Max := new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(1))
			INT256MAX := new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(255), nil), big.NewInt(1))
			Check_val := new(big.Int).Sub(INT256MAX, extended)
			if stack[0].Cmp(big.NewInt(256)) != -1 {
				mask := big.NewInt(1)
				mask = mask.Lsh(mask, 255)
				// Create a mask for the first bit
				firstBitMask := stack[1]
				// Assuming mask is 256 bits

				// Extract the first bit of mask
				firstBit := new(big.Int).And(mask, firstBitMask)

				if firstBit.Cmp(big.NewInt(0)) == 0 {
					// If the first bit is 0
					extended = new(big.Int).Lsh(mask, 1)
				} else {
					// If the first bit is 1
					mask = new(big.Int).Lsh(mask, 1)
					mask = new(big.Int).Sub(mask, big.NewInt(1))
					extended = mask
				}

			} else {
				if Check_val.Cmp(big.NewInt(0)) == -1 {
					extended := stack[1]
					shift := uint(stack[0].Uint64())

					// Create a mask that has ones in the positions that should be filled with ones after the shift
					mask := new(big.Int).Lsh(big.NewInt(1), shift)
					mask.Sub(mask, big.NewInt(1))
					mask.Lsh(mask, 256-shift)

					// Perform the right shift and apply the mask
					extended.Rsh(extended, shift)
					extended.Or(extended, mask)
				} else {
					extended.Rsh(extended, uint(stack[0].Int64()))
				}
			}

			extended.And(extended, UINT256Max)
			stack = stack[2:]

			stack = append([]*big.Int{extended}, stack...)

		case 0x1A:
			extended := stack[1]
			UINT256Max := new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(1))
			extended.And(extended, UINT256Max)
			mask := big.NewInt(255)

			shift := (31 - stack[0].Int64()) * 8
			if shift >= 0 && shift <= 256 {
				mask = mask.Lsh(mask, uint(shift))
			}

			extended = extended.And(extended, mask)
			extended = extended.Rsh(extended, uint(shift))
			extended = extended.And(extended, big.NewInt(255))
			stack = stack[2:]
			stack = append([]*big.Int{extended}, stack...)
		case 0x80, 0x81, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89, 0x8A, 0x8B, 0x8C, 0x8D, 0x8E, 0x8F:
			op2 := op - 0x80

			dup := stack[op2]
			stack = append([]*big.Int{dup}, stack...)
		case 0x90, 0x91, 0x92, 0x93, 0x94, 0x95, 0x96, 0x97, 0x98, 0x99, 0x9A, 0x9B, 0x9C, 0x9D, 0x9E, 0x9F:
			op2 := op - 0x90
			toSwap := stack[op2+1]
			top := stack[0]
			stack[0] = toSwap
			stack[op2+1] = top
		case 0xFE:
			return nil, false

		case 0x58:

			counter := 0
			for i := pc - 2; i >= 0; i-- {
				if i < len(code) {
					if code[i] == 60 {
						counter = counter + 2
					} else {
						counter++
					}
				}

			}
			stack = append([]*big.Int{big.NewInt(int64(counter))}, stack...)
		case 0x5A:
			UINT256Max := new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(1))
			stack = append([]*big.Int{UINT256Max}, stack...)
		case 0x56:
			hello := true
			pc, stack, hello = jumpdest(pc, code, stack)
			if !hello {
				return stack, false
			}
		case 0x57:
			value := stack[1]
			if value.Cmp(big.NewInt(0)) != 0 {
				hello := true
				pc, stack, hello = jumpdest(pc, code, stack)
				if !hello {
					return stack, false
				}
			} else {
				stack = []*big.Int{}

			}
		}

	}
	return stack, true
}
