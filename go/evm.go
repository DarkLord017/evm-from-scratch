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
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"golang.org/x/crypto/sha3"
)

type Transaction struct {
	To       string `json:"to"`
	From     string `json:"from"`
	Origin   string `json:"origin"`
	Gasprice string `json:"gasprice"`
	Value    string `json:"value"`
	Data     string `json:"data"`
}

type Log struct {
	Address string   `json:"address"`
	Data    string   `json:"data"`
	Topics  []string `json:"topics"`
}

type block struct {
	Basefee    string `json:"basefee"`
	Coinbase   string `json:"coinbase"`
	Timestamp  string `json:"timestamp"`
	Number     string `json:"number"`
	Difficulty string `json:"difficulty"`
	Gaslimit   string `json:"gaslimit"`
	ChainId    string `json:"chainId"`
}

type Account struct {
	Balance  string   `json:"balance"`
	UserCode usercode `json:"code"`
}

type usercode struct {
	Asm string `json:"asm"`
	Bin string `json:"bin"`
}

type Store map[string]Storage

type Accounts map[string]Account

func generateContractAddress(sender string, nonce uint64) string {
	// Convert sender address to bytes
	senderBytes, _ := hex.DecodeString(strings.TrimPrefix(sender, "0x"))

	// Convert nonce to bytes
	nonceBytes := new(big.Int).SetUint64(nonce).Bytes()

	// Concatenate sender address and nonce
	data := append(senderBytes, nonceBytes...)

	// Hash the concatenated data to generate a unique address
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(data)
	return "0x" + hex.EncodeToString(hasher.Sum(nil)[12:]) // Ethereum addresses are the last 20 bytes of the hash
}

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

type Storage struct {
	data      []byte
	offsetMax int
}

func NewStorage(size int) *Storage {
	return &Storage{
		data:      make([]byte, size),
		offsetMax: 0,
	}
}

// Store stores 32 bytes in memory at the specified offset.
func (m *Storage) Store(value []byte) {

	m.data = append(m.data, value...)
}

// Load loads 32 bytes from memory at the specified offset.
func (m *Storage) Load(offset int) []byte {

	return m.data[offset:]
}

type Memory struct {
	data      []byte
	offsetMax int
}

func NewMemory(size int) *Memory {
	return &Memory{
		data:      make([]byte, size),
		offsetMax: 0,
	}
}

// Store stores 32 bytes in memory at the specified offset.
func (m *Memory) Store(offset int, value []byte) {
	m.MSIZE(offset)
	copy(m.data[offset:], value)
}

// Load loads 32 bytes from memory at the specified offset.
func (m *Memory) Load(offset int) []byte {
	m.MSIZE(offset)
	return m.data[offset : offset+32]
}

func (m *Memory) LoadforSHA3(offset int, size int) []byte {
	m.MSIZE(offset)
	return m.data[offset : offset+size]
}

func (m *Memory) Store8(offset int, value byte) {
	m.MSIZE(offset - 32)
	m.data[offset] = value
}

func (m *Memory) MSIZE(offset int) int {
	if offset+32 > m.offsetMax {
		m.offsetMax = offset + 32
	}
	return m.offsetMax
}

func (m *Memory) GetOffsetMax() int {
	return m.offsetMax
}

// Run runs the EVM code and returns the stack and a success indicator.
func Evm(code []byte, transaction Transaction, Block block, state Accounts, sstore Store) ([]*big.Int, bool, []Log, string, Accounts) {
	var logs []Log // var account Account
	var stack []*big.Int
	var returnDataSize string
	memory := NewMemory(1024)
	if state == nil {
		state = make(Accounts)
	}
	// 1024 bytes of memory
	pc := 0
	ret := ""

	for pc < len(code) {
		op := code[pc]
		pc++

		// TODO: Implement the EVM here!
		switch op {
		default:
			if 0x60 <= op && op <= 0x7f {
				increment := int(op-0x60) + 1

				if pc+increment > len(code) {
					return nil, false, nil, ret, state

				}
				value := new(big.Int).SetBytes(code[pc : pc+increment])
				stack = append([]*big.Int{value}, stack...)
				pc += increment
			}
		case 0x00:
			return stack, true, nil, ret, state

		case 0x5F:
			value := big.NewInt(0)
			stack = append([]*big.Int{value}, stack...)

		case 0x50:
			if len(stack) < 1 {
				return nil, false, nil, ret, state
			}

			stack = stack[1:]

		case 0x01:
			if len(stack) < 2 {
				return nil, false, nil, ret, state
			}

			value := new(big.Int).Add(stack[0], stack[1])
			value.Mod(value, new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil)) // Modulo operation to handle overflow
			stack = stack[2:]
			stack = append([]*big.Int{value}, stack...)

		case 0x02:
			if len(stack) < 2 {
				return nil, false, nil, ret, state
			}

			value := new(big.Int).Mul(stack[0], stack[1])
			value.Mod(value, new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil)) // Modulo operation to handle overflow
			stack = stack[2:]
			stack = append([]*big.Int{value}, stack...)
		case 0x03:
			if len(stack) < 2 {
				return nil, false, nil, ret, state
			}

			value := new(big.Int).Sub(stack[0], stack[1])
			value.Mod(value, new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil)) // Modulo operation to handle overflow`
			stack = stack[2:]
			stack = append([]*big.Int{value}, stack...)

		case 0x04:
			if len(stack) < 2 {
				return nil, false, nil, ret, state
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
				return nil, false, nil, ret, state
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
				return nil, false, nil, ret, state
			}

			valueMod := new(big.Int).Add(stack[0], stack[1])
			stack = stack[2:]
			value := new(big.Int).Mod(valueMod, stack[0])
			stack = stack[1:]
			value.Mod(value, new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil))
			stack = append([]*big.Int{value}, stack...)
		case 0x09:
			if len(stack) < 3 {
				return nil, false, nil, ret, state
			}

			valueMod := new(big.Int).Mul(stack[0], stack[1])
			stack = stack[2:]
			value := new(big.Int).Mod(valueMod, stack[0])
			stack = stack[1:]
			value.Mod(value, new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil))
			stack = append([]*big.Int{value}, stack...)
		case 0x0A:
			if len(stack) < 2 {
				return nil, false, nil, ret, state
			}

			value := new(big.Int).Exp(stack[0], stack[1], nil)
			stack = stack[2:]

			value.Mod(value, new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil))
			stack = append([]*big.Int{value}, stack...)
		case 0x0B:
			if len(stack) < 2 {
				return nil, false, nil, ret, state
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
				return stack, false, nil, ret, state
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
				return nil, false, nil, ret, state
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
				return nil, false, nil, ret, state
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
				return nil, false, nil, ret, state
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
				return nil, false, nil, ret, state
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
				return nil, false, nil, ret, state
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
				return nil, false, nil, ret, state
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
				return nil, false, nil, ret, state
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
				return nil, false, nil, ret, state
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
				return nil, false, nil, ret, state
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
				return nil, false, nil, ret, state
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
			return nil, false, nil, ret, state

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
				return stack, false, nil, ret, state
			}
		case 0x57:
			value := stack[1]
			if value.Cmp(big.NewInt(0)) != 0 {
				hello := true
				pc, stack, hello = jumpdest(pc, code, stack)
				if !hello {
					return stack, false, nil, ret, state
				}
			} else {
				stack = []*big.Int{}

			}

		case 0x52: // MSTORE
			offset := stack[0]
			stack = stack[1:]
			value := stack[0]
			stack = stack[1:]
			offsetInt := int(offset.Int64())
			valueBytes := value.Bytes()

			if len(valueBytes) < 32 {
				padding := make([]byte, 32-len(valueBytes))
				valueBytes = append(padding, valueBytes...)
			}
			memory.Store(offsetInt, valueBytes)
		case 0x51: // MLOAD
			offset := stack[0]
			stack = stack[1:]
			offsetInt := int(offset.Int64())
			value := new(big.Int).SetBytes(memory.Load(offsetInt))
			stack = append([]*big.Int{value}, stack...)
		case 0x53: // MSTORE8
			offset := stack[0]

			stack = stack[1:]

			value := int8(stack[0].Int64())
			stack = stack[1:]
			offsetInt := int(offset.Uint64())
			memory.Store8(offsetInt, byte(value))
		case 0x59:
			value := memory.GetOffsetMax()
			m := 32
			final_val := ((value + m - 1) / m) * m
			stack = append([]*big.Int{big.NewInt(int64(final_val))}, stack...)
		case 0x20:
			offset := stack[0]
			stack = stack[1:]
			size := stack[0]
			stack = stack[1:]
			data := memory.LoadforSHA3(int(offset.Int64()), int(size.Int64()))
			hash := sha3.NewLegacyKeccak256()
			_, err := hash.Write(data)
			if err != nil {
				panic(err)
			}
			result := hash.Sum(nil)

			stack = append([]*big.Int{new(big.Int).SetBytes(result)}, stack...)
		case 0x30:
			// Check if the 'to' address is not empty
			if len(transaction.To) == 0 {
				return nil, false, nil, ret, state
			}
			toAddress := transaction.To
			// Remove the "0x" prefix if present

			toAddress = toAddress[2:]

			// Convert hex string to big.Int
			bigInt := new(big.Int)
			bigInt.SetString(toAddress, 16)
			stack = append([]*big.Int{new(big.Int).Set(bigInt)}, stack...)
		case 0x33:
			if len(transaction.From) == 0 {
				return nil, false, nil, ret, state
			}
			toAddress := transaction.From
			// Remove the "0x" prefix if present

			toAddress = toAddress[2:]

			// Convert hex string to big.Int
			bigInt := new(big.Int)
			bigInt.SetString(toAddress, 16)
			stack = append([]*big.Int{bigInt}, stack...)
		case 0x32:
			if len(transaction.Origin) == 0 {
				return nil, false, nil, ret, state
			}
			toAddress := transaction.Origin
			// Remove the "0x" prefix if present

			toAddress = toAddress[2:]

			// Convert hex string to big.Int
			bigInt := new(big.Int)
			bigInt.SetString(toAddress, 16)
			stack = append([]*big.Int{bigInt}, stack...)
		case 0x3A:
			if len(transaction.Gasprice) == 0 {
				return nil, false, nil, ret, state
			}
			toAddress := transaction.Gasprice
			toAddress = toAddress[2:]

			// Convert hex string to big.Int
			bigInt := new(big.Int)
			bigInt.SetString(toAddress, 16)
			stack = append([]*big.Int{bigInt}, stack...)
		case 0x48:
			if len(Block.Basefee) == 0 {
				return nil, false, nil, ret, state
			}
			toAddress := Block.Basefee
			toAddress = toAddress[2:]

			bigInt := new(big.Int)
			bigInt.SetString(toAddress, 16)
			stack = append([]*big.Int{bigInt}, stack...)
		case 0x41:
			if len(Block.Coinbase) == 0 {
				return nil, false, nil, ret, state
			}

			toAddress := Block.Coinbase
			toAddress = toAddress[2:]

			bigInt := new(big.Int)
			bigInt.SetString(toAddress, 16)
			stack = append([]*big.Int{bigInt}, stack...)

		case 0x42:
			if len(Block.Timestamp) == 0 {
				return nil, false, nil, ret, state
			}

			toAddress := Block.Timestamp
			toAddress = toAddress[2:]

			bigInt := new(big.Int)
			bigInt.SetString(toAddress, 16)
			stack = append([]*big.Int{bigInt}, stack...)

		case 0x43:
			if len(Block.Number) == 0 {
				return nil, false, nil, ret, state
			}

			toAddress := Block.Number
			toAddress = toAddress[2:]

			bigInt := new(big.Int)
			bigInt.SetString(toAddress, 16)
			stack = append([]*big.Int{bigInt}, stack...)
		case 0x44:
			if len(Block.Difficulty) == 0 {
				return nil, false, nil, ret, state
			}

			toAddress := Block.Difficulty
			toAddress = toAddress[2:]

			bigInt := new(big.Int)
			bigInt.SetString(toAddress, 16)
			stack = append([]*big.Int{bigInt}, stack...)
		case 0x45:
			if len(Block.Gaslimit) == 0 {
				return nil, false, nil, ret, state
			}

			toAddress := Block.Gaslimit
			toAddress = toAddress[2:]

			bigInt := new(big.Int)
			bigInt.SetString(toAddress, 16)
			stack = append([]*big.Int{bigInt}, stack...)
		case 0x46:
			if len(Block.ChainId) == 0 {
				return nil, false, nil, ret, state
			}

			toAddress := Block.ChainId
			toAddress = toAddress[2:]

			bigInt := new(big.Int)
			bigInt.SetString(toAddress, 16)
			stack = append([]*big.Int{bigInt}, stack...)
		case 0x31:
			value := stack[0]
			stack = stack[1:]
			hexValue := fmt.Sprintf("0x%x", value)
			if account, exists := state[hexValue]; exists {

				toAddress := account.Balance
				toAddress = toAddress[2:]

				bigInt := new(big.Int)
				bigInt.SetString(toAddress, 16)
				stack = append([]*big.Int{bigInt}, stack...)
			} else {
				stack = append([]*big.Int{big.NewInt(0)}, stack...)
			}
		case 0x34:
			toAddress := transaction.Value
			toAddress = toAddress[2:]

			bigInt := new(big.Int)
			bigInt.SetString(toAddress, 16)
			stack = append([]*big.Int{bigInt}, stack...)
		case 0x35:
			offset := stack[0].Int64() * 2
			stack = stack[1:]
			CallDataFull := transaction.Data

			CallDataFull = CallDataFull[offset:]
			if len(CallDataFull) < 64 {
				padding := 64 - len(CallDataFull)
				CallDataFull += strings.Repeat("0", padding)
			}

			bigInt := new(big.Int)
			bigInt.SetString(CallDataFull, 16)
			stack = append([]*big.Int{bigInt}, stack...)

		case 0x36:

			bigInt := len(transaction.Data) / 2
			stack = append([]*big.Int{big.NewInt(int64(bigInt))}, stack...)

		case 0x37: // CALLDATACOPY
			destOffset := int(stack[0].Int64())
			offset := int(stack[1].Int64())
			size := int(stack[2].Int64())
			stack = stack[3:]

			// Convert transaction data from hex string to byte slice
			data, err := hex.DecodeString(transaction.Data)
			if err != nil {
				panic("invalid transaction data")
			}

			// Initialize the slice to hold the copied data
			valueBytes := make([]byte, size)

			// Copy the portion of the data from the offset, right-padded with zeros if needed
			if offset < len(data) {
				copyEnd := offset + size
				if copyEnd > len(data) {
					copyEnd = len(data)
				}
				copy(valueBytes, data[offset:copyEnd])
			}

			// Store the result in memory
			memory.Store(destOffset, valueBytes)
		case 0x38:
			value := len(code)

			stack = append([]*big.Int{big.NewInt(int64(value))}, stack...)
		case 0x39:
			destOffset := int(stack[0].Int64())
			offset := int(stack[1].Int64())
			size := int(stack[2].Int64())
			stack = stack[3:]

			// Convert transaction data from hex string to byte slice
			data := code

			// Initialize the slice to hold the copied data
			valueBytes := make([]byte, size)

			// Copy the portion of the data from the offset, right-padded with zeros if needed
			if offset < len(data) {
				copyEnd := offset + size
				if copyEnd > len(data) {
					copyEnd = len(data)
				}
				copy(valueBytes, data[offset:copyEnd])
			}

			// Store the result in memory
			memory.Store(destOffset, valueBytes)
		case 0x3b:
			answer := 0
			value := stack[0]
			stack = stack[1:]
			hexValue := fmt.Sprintf("0x%x", value)
			// Debugging statement
			stateEntry := state[hexValue]

			// Convert hex string to byte slice
			code := stateEntry.UserCode.Bin
			answer = len(code) / 2

			stack = append([]*big.Int{big.NewInt(int64(answer))}, stack...)
		case 0x3c:
			value := stack[0]
			stack = stack[1:]

			destOffset := int(stack[0].Int64())
			offset := int(stack[1].Int64())
			size := int(stack[2].Int64())
			stack = stack[3:]

			hexValue := fmt.Sprintf("0x%x", value)

			stateEntry := state[hexValue]
			// Convert transaction data from hex string to byte slice
			data, err := hex.DecodeString(stateEntry.UserCode.Bin)
			if err != nil {
				return nil, false, nil, ret, state
			}

			// Initialize the slice to hold the copied data
			valueBytes := make([]byte, size)

			// Copy the portion of the data from the offset, right-padded with zeros if needed
			if offset < len(data) {
				copyEnd := offset + size
				if copyEnd > len(data) {
					copyEnd = len(data)
				}
				copy(valueBytes, data[offset:copyEnd])
			}

			// Store the result in memory
			memory.Store(destOffset, valueBytes)
		case 0x3f:

			value := stack[0]
			stack = stack[1:]

			hexValue := fmt.Sprintf("0x%x", value)

			if stateEntry, exists := state[hexValue]; exists {
				// Convert transaction data from hex string to byte slice}
				data, err := hex.DecodeString(stateEntry.UserCode.Bin)
				if err != nil {
					return nil, false, nil, ret, state
				}

				hash := sha3.NewLegacyKeccak256()
				_, error := hash.Write(data)
				if error != nil {
					panic(error)
				}
				result := hash.Sum(nil)

				stack = append([]*big.Int{new(big.Int).SetBytes(result)}, stack...)
			} else {
				stack = append([]*big.Int{big.NewInt(0)}, stack...)
			}
		case 0x47:
			hexValue := transaction.To

			if account, exists := state[hexValue]; exists {

				toAddress := account.Balance
				toAddress = toAddress[2:]

				bigInt := new(big.Int)
				bigInt.SetString(toAddress, 16)
				stack = append([]*big.Int{bigInt}, stack...)
			} else {
				stack = append([]*big.Int{big.NewInt(0)}, stack...)
			}
		case 0x55:
			storage := NewStorage(32)
			key := stack[0].Bytes()
			valueBytes := make([]byte, 32)
			copy(valueBytes, key)
			stack = stack[1:]
			storage.data = append(storage.data, stack[0].Bytes()...)
			stack = stack[1:]
			if sstore == nil {
				return nil, false, nil, ret, state
			}
			sstore[string(valueBytes)] = *storage
		case 0x54:
			key := stack[0].Bytes()
			stack = stack[1:]
			valueBytes := make([]byte, 32)
			copy(valueBytes, key)

			storage := sstore[string(valueBytes)]
			value := storage.data
			stack = append([]*big.Int{new(big.Int).SetBytes(value)}, stack...)
		case 0xA0, 0xA1, 0xA2, 0xA3, 0xA4:
			var topics []string
			op2 := op - 0xA0

			offset := stack[0]
			stack = stack[1:]
			value := int(stack[0].Int64())
			stack = stack[1:]
			offsetInt := int(offset.Int64())
			data := memory.LoadforSHA3(offsetInt, value)
			if int(op2) > 0 {
				for i := 0; i < int(op2); i++ {
					topic := stack[0]
					stack = stack[1:]
					hexValue := fmt.Sprintf("0x%x", topic)
					topics = append(topics, hexValue)
				}
			} else {
				topics = []string{}
			}

			log := Log{
				Address: transaction.To,
				Data:    hex.EncodeToString(data),
				Topics:  topics,
			}
			logs = append(logs, log)
		case 0xf3:
			offset := int(stack[0].Int64())
			stack = stack[1:]
			size := int(stack[0].Int64())
			stack = stack[1:]
			data := memory.LoadforSHA3(offset, size)
			ret = fmt.Sprintf("%x", data)
		case 0xfd:
			offset := int(stack[0].Int64())
			stack = stack[1:]
			size := int(stack[0].Int64())
			stack = stack[1:]
			data := memory.LoadforSHA3(offset, size)
			ret = fmt.Sprintf("%x", data)
			return stack, false, logs, ret, state
		case 0xf1:
			// gas := stack[0]
			stack = stack[1:]
			address := stack[0]
			stack = stack[1:]
			// value := stack[0]
			stack = stack[3:]
			hexValue := fmt.Sprintf("0x%x", address)

			data, err := hex.DecodeString(state[hexValue].UserCode.Bin)
			if err != nil {
				return nil, false, nil, ret, state
			}

			offset := int(stack[0].Int64())
			size := stack[1].Int64()

			stack = stack[2:]

			tx := Transaction{
				From: transaction.To,
				To:   hexValue,
			}

			sstoreCALL := make(map[string]Storage)

			_, callsuccess, _, returnString, st := Evm(data, tx, block{}, state, sstoreCALL)
			state = st

			returnDataSize = returnString

			if len(returnString) < int(size*2) {
				padding := int(size*2) - len(returnString)
				returnString = strings.Repeat("0", padding) + returnString
			}

			if len(returnString) > 0 {
				returnString = returnString[:size*2]
			}

			dataRet, _ := hex.DecodeString(returnString)

			memory.Store(offset, dataRet)

			if callsuccess {
				stack = append([]*big.Int{big.NewInt(1)}, stack...)
			} else {
				stack = append([]*big.Int{big.NewInt(0)}, stack...)
			}

		case 0x3D:
			if len(returnDataSize) > 0 {

				stack = append([]*big.Int{big.NewInt(int64(len(returnDataSize) / 2))}, stack...)
			} else {
				stack = append([]*big.Int{big.NewInt(0)}, stack...)
			}
		case 0x3e:
			destOffset := int(stack[0].Int64())
			offset := int(stack[1].Int64())
			size := int(stack[2].Int64())
			stack = stack[3:]

			// Convert transaction data from hex string to byte slice
			data, _ := hex.DecodeString(returnDataSize)

			// Initialize the slice to hold the copied data
			valueBytes := make([]byte, size)

			// Copy the portion of the data from the offset, right-padded with zeros if needed
			if offset < len(data) {
				copyEnd := offset + size
				if copyEnd > len(data) {
					copyEnd = len(data)
				}
				copy(valueBytes, data[offset:copyEnd])
			}

			// Store the result in memory
			memory.Store(destOffset, valueBytes)
		case 0xF4:
			// gas := stack[0]
			stack = stack[1:]
			address := stack[0]
			stack = stack[1:]
			// value := stack[0]
			stack = stack[2:]
			hexValue := fmt.Sprintf("0x%x", address)

			data, err := hex.DecodeString(state[hexValue].UserCode.Bin)
			if err != nil {
				return nil, false, nil, ret, state
			}

			offset := int(stack[0].Int64())
			size := stack[1].Int64()

			stack = stack[2:]

			tx := Transaction{
				From:     transaction.From,
				To:       transaction.To,
				Origin:   transaction.Origin,
				Gasprice: transaction.Gasprice,
				Value:    transaction.Value,
				Data:     transaction.Data,
			}

			_, callsuccess, _, returnString, _ := Evm(data, tx, block{}, nil, sstore)

			returnDataSize = returnString

			if len(returnString) < int(size*2) {
				padding := int(size*2) - len(returnString)
				returnString = strings.Repeat("0", padding) + returnString
			}

			if len(returnString) > 0 {
				returnString = returnString[:size*2]
			}

			dataRet, _ := hex.DecodeString(returnString)

			memory.Store(offset, dataRet)

			if callsuccess {
				stack = append([]*big.Int{big.NewInt(1)}, stack...)
			} else {
				stack = append([]*big.Int{big.NewInt(0)}, stack...)
			}
		case 0xFA:
			// gas := stack[0]
			stack = stack[1:]
			address := stack[0]
			stack = stack[1:]
			// value := stack[0]
			stack = stack[2:]
			hexValue := fmt.Sprintf("0x%x", address)

			data, err := hex.DecodeString(state[hexValue].UserCode.Bin)
			if err != nil {
				return nil, false, nil, ret, state
			}

			offset := int(stack[0].Int64())
			size := stack[1].Int64()

			stack = stack[2:]

			tx := Transaction{
				From: transaction.To,
				To:   hexValue,
			}

			_, callsuccess, _, returnString, _ := Evm(data, tx, block{}, nil, nil)

			returnDataSize = returnString

			if len(returnString) < int(size*2) {
				padding := int(size*2) - len(returnString)
				returnString = strings.Repeat("0", padding) + returnString
			}

			if len(returnString) > 0 {
				returnString = returnString[:size*2]
			}

			dataRet, _ := hex.DecodeString(returnString)

			memory.Store(offset, dataRet)

			if callsuccess {
				stack = append([]*big.Int{big.NewInt(1)}, stack...)
			} else {
				stack = append([]*big.Int{big.NewInt(0)}, stack...)
			}

		case 0xF0:
			if len(stack) < 3 {
				return nil, false, logs, "Not enough items on stack for CREATE operation", state
			}
			value := stack[0]                 // value to transfer (in Ether)
			inOffset := int(stack[1].Int64()) // offset of input data in memory
			inSize := int(stack[2].Int64())   // size of input data

			if inOffset < 0 || inOffset+inSize > len(memory.data) {
				return nil, false, logs, "Memory access out of bounds", state
			}

			// Load the initialization code from memory
			newContractAddress := generateContractAddress(transaction.To, 0)
			data := memory.data[inOffset : inOffset+inSize]

			tx := Transaction{
				From: transaction.To,
				To:   newContractAddress,
			}

			sstoreCALL := make(map[string]Storage)

			_, callsuccess, _, bytecode, st := Evm(data, tx, block{}, state, sstoreCALL)
			state = st

			if !callsuccess {
				stack = stack[3:]
				stack = append([]*big.Int{big.NewInt(0)}, stack...)
			} else {

				// Generate the new contract address

				// Store the new contract in the state
				state[newContractAddress] = Account{
					Balance: "0x" + value.String(),
					UserCode: usercode{
						Asm: "",
						Bin: bytecode,
					},
				}

				// Push the new contract address onto the stack
				data, _ := hex.DecodeString(strings.TrimPrefix(newContractAddress, "0x"))
				stack = stack[3:]
				stack = append([]*big.Int{new(big.Int).SetBytes(data)}, stack...)
			}
		case 0xFF:
			newbal := state[transaction.To].Balance
			state = nil
			address := stack[0]
			stack = stack[1:]
			hexValue := fmt.Sprintf("0x%x", address)
			state = make(Accounts)

			state[hexValue] = Account{
				Balance: newbal,
			}

		}

	}
	return stack, true, logs, ret, state
}
