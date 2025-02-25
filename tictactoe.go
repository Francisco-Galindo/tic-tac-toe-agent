package main

import (
	"fmt"
	"math/bits"
	"math/rand"
	"time"
)

const (
	EXACT      uint8 = 1
	LOWERBOUND uint8 = 2
	UPPERBOUND uint8 = 3
)

const N uint16 = 4

const (
	// OUT_OF_BOUNDS uint16 = 0b1111111000000000
	// FULL_BOARD    uint16 = 0b0000000111111111
	// COL_1         uint16 = 0b0000000100100100
	// COL_2         uint16 = 0b0000000010010010
	// COL_3         uint16 = 0b0000000001001001
	// ROW_1         uint16 = 0b0000000111000000
	// ROW_2         uint16 = 0b0000000000111000
	// ROW_3         uint16 = 0b0000000000000111
	// DIAG_UP       uint16 = 0b0000000001010100
	// DIAG_DOWN     uint16 = 0b0000000100010001

	OUT_OF_BOUNDS uint16 = 0b0000000000000000
	FULL_BOARD    uint16 = 0b1111111111111111
)

var WINNING_PATTERNS = []uint16{
	0b0000000000000111,
	0b0000000000001110,
	0b0000000001110000,
	0b0000000011100000,
	0b0000011100000000,
	0b0000111000000000,
	0b0111000000000000,
	0b1110000000000000,

	0b0000000100010001,
	0b0001000100010000,
	0b0000001000100010,
	0b0010001000100000,
	0b0000010001000100,
	0b0100010001000000,
	0b0000100010001000,
	0b1000100010000000,

	0b0000010000100001,
	0b0000100001000010,
	0b0100001000010000,
	0b1000010000100000,

	0b0000000100100100,
	0b0000001001001000,
	0b0001001001000000,
	0b0010010010000000,
}

var zobrist_table = make([]uint64, 2*N*N)
var ttHits int64 = 0
var tranTable = make(map[uint64]*TtEntry)

func init_zobrist() {
	for i := range 2 * N * N {
		zobrist_table[i] = rand.Uint64()
	}
}

type TtEntry struct {
	zobrist uint64
	flag    uint8
	depth   uint16
	value   int16
}

func zobrist_hash(xMoves, oMoves uint16) uint64 {
	var h uint64 = 0
	for i := range N * N {
		if xMoves|(0b1<<i) == xMoves {
			h ^= zobrist_table[i]
		}
	}
	for i := range N * N {
		if oMoves|(0b1<<i) == oMoves {
			h ^= zobrist_table[N*N+i]
		}
	}
	return h
}

func getPossibleMoves(board uint16) []uint16 {
	board = (^(board)) ^ OUT_OF_BOUNDS
	possibleMoves := make([]uint16, 0)
	for board != 0 {
		idx := uint16(bits.TrailingZeros16(board))
		possibleMoves = append(possibleMoves, idx)
		board ^= 0b1 << idx
	}

	return possibleMoves
}

func isFull(board uint16) bool {
	return (board & FULL_BOARD) == FULL_BOARD
}

func evalBoard(player, adversary uint16) int16 {
	for _, pattern := range WINNING_PATTERNS {
		if pattern&player == pattern {
			return 100
		} else if pattern&adversary == pattern {
			return -100
		}
	}
	return 0
}

func negamax(boardHash uint64, agent, adversary uint16, depth uint16, alpha, beta int16, marker bool) int16 {
	alphaOrig := alpha

	ttEntry := tranTable[boardHash]

	if ttEntry != nil && ttEntry.depth >= depth {
		ttHits++
		if ttEntry.flag == EXACT {
			return ttEntry.value
		} else if ttEntry.flag == LOWERBOUND {
			alpha = max(alpha, ttEntry.value)
		} else if ttEntry.flag == UPPERBOUND {
			beta = min(beta, ttEntry.value)
		}

		if alpha >= beta {
			return ttEntry.value
		}
	} else if ttEntry == nil {
		ttEntry = &TtEntry{zobrist: boardHash}
	}

	if depth == 0 || isFull(agent|adversary) {
		return 0
	}
	score := evalBoard(adversary, agent)
	if score != 0 {
		if score > 0 {
			return score + int16(depth)
		} else {
			return score - int16(depth)
		}
	}
	value := int16(-10000)
	for _, move := range getPossibleMoves(agent | adversary) {
		adversary |= 0b1 << move

		newHash := boardHash
		if marker {
			newHash ^= zobrist_table[move]
		} else {
			newHash ^= zobrist_table[N*N+move]
		}

		value = max(value, -negamax(newHash, adversary, agent, depth-1, -beta, -alpha, !marker))

		adversary ^= 0b1 << move
		alpha = max(alpha, value)
		if alpha >= beta {
			break
		}
	}

	ttEntry.value = value
	if value <= alphaOrig {
		ttEntry.flag = UPPERBOUND
	} else if value >= beta {
		ttEntry.flag = LOWERBOUND
	} else {
		ttEntry.flag = EXACT
	}
	ttEntry.depth = depth
	tranTable[boardHash] = ttEntry

	return value
}

func playerMove(player, adversary uint16) uint16 {
	var idx uint16 = 123
	for idx >= N*N {
		fmt.Printf("Your move: ")
		fmt.Scanf("%d", &idx)
	}
	return (player | 0b1<<idx)
}

func agentMove(agent, adversary uint16) uint16 {
	var score int16 = -1000
	var bestMove uint16 = 100

	fmt.Println("Agent's move:")
	for _, move := range getPossibleMoves(agent | adversary) {
		agent |= 0b1 << move
		boardHash := zobrist_hash(adversary, agent)
		// fmt.Println(boardHash)
		newScore := -negamax(boardHash, agent, adversary, 16, -10000, 10000, false)
		// fmt.Println("\t",newScore)
		agent ^= 0b1 << move
		if newScore > score {
			score = newScore
			bestMove = move
		}
	}
	// fmt.Println("Move:", bestMove)
	return agent | (0b1 << bestMove)
}

func printBoard(xMoves, oMoves uint16) {
	for i := range N {
		for j := range N {
			idx := uint16(0b1 << (i*N + j))
			if xMoves&idx != 0 {
				fmt.Printf("X ")
			} else if oMoves&idx != 0 {
				fmt.Printf("O ")
			} else {
				fmt.Printf("- ")
			}
		}
		fmt.Println("")
	}
	fmt.Println("")
}

func main() {
	var player1, player2 uint16
	printBoard(player1, player2)
	init_zobrist()

	xTurn := false
	for !isFull(player1 | player2) {
		if xTurn {
			player1 = playerMove(player1, player2)
		} else {
			start := time.Now()
			player2 = agentMove(player2, player1)
			elapsed := time.Since(start).Seconds()
			fmt.Println("Time: ", elapsed)
		}
		printBoard(player1, player2)

		if score := evalBoard(player1, player2); score != 0 {
			fmt.Println("GOL", score)
			break
		}

		xTurn = !xTurn
	}
	fmt.Println(ttHits)
}
