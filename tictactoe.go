package main

import (
	"flag"
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

var N uint16 = 3

const (
	COL_1     uint16 = 0b0000000100100100
	COL_2     uint16 = 0b0000000010010010
	COL_3     uint16 = 0b0000000001001001
	ROW_1     uint16 = 0b0000000111000000
	ROW_2     uint16 = 0b0000000000111000
	ROW_3     uint16 = 0b0000000000000111
	DIAG_UP   uint16 = 0b0000000001010100
	DIAG_DOWN uint16 = 0b0000000100010001
)

const (
	OUT_OF_BOUNDS_3 uint16 = 0b1111111000000000
	FULL_BOARD_3    uint16 = 0b0000000111111111
	OUT_OF_BOUNDS_4 uint16 = 0b0000000000000000
	FULL_BOARD_4    uint16 = 0b1111111111111111
)

var WINNING_PATTERNS_3 = []uint16{
	COL_1, COL_2, COL_3, ROW_1, ROW_2, ROW_3, DIAG_DOWN, DIAG_UP,
}

var WINNING_PATTERNS_4 = []uint16{
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

var WINNING_PATTERNS []uint16
var OUT_OF_BOUNDS uint16
var FULL_BOARD uint16

var zobrist_table = make([]uint64, 0)
var ttHits int64 = 0
var tranTable = make(map[uint64]*TtEntry)
var isGameSmart bool = true
var isGameDesperate bool
var isGameRandom bool

type TtEntry struct {
	zobrist uint64
	flag    uint8
	depth   uint16
	value   int16
}

func init_zobrist() {
	zobrist_table = make([]uint64, 2*N*N)
	for i := range 2 * N * N {
		zobrist_table[i] = rand.Uint64()
	}
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
			newHash ^= zobrist_table[N*N+move]
		} else {
			newHash ^= zobrist_table[move]
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
	board := player | adversary
	for idx >= N*N || (board|(0b1<<idx) == board) {
		fmt.Printf("Your move: ")
		fmt.Scanf("%d", &idx)
	}
	return (player | 0b1<<idx)
}

func agentMove(agent, adversary uint16) uint16 {
	fmt.Println("Agent's move...")
	if !isGameDesperate && !isGameRandom {
		return getSmartMove(agent, adversary)
	} else if isGameDesperate && !isGameRandom {
		return getDesperateMove(agent, adversary)
	}

	return getRandomMove(agent, adversary)
}

func getSmartMove(agent, adversary uint16) uint16 {
	var bestScore int16 = -10000
	var bestMove uint16 = 1000

	for i := range N * N {
		bestScore = -10000
		bestMove = 1000
		for _, move := range getPossibleMoves(agent | adversary) {
			agent |= 0b1 << move // Haz el movimiento
			boardHash := zobrist_hash(adversary, agent)
			newScore := -negamax(boardHash, agent, adversary, i, -10000, 10000, false)

			if newScore > 0 {
				return agent | (0b1 << move)
			}

			agent ^= 0b1 << move // Deshaz el movimiento

			if newScore > bestScore {
				bestScore = newScore
				bestMove = move
			}
		}
	}

	if bestScore < 0 {
		return getDesperateMove(agent, adversary)
	}

	return agent | (0b1 << bestMove)
}

func getDesperateMove(agent, adversary uint16) uint16 {
	for _, move := range getPossibleMoves(agent | adversary) {
		adversary |= 0b1 << move
		if score := evalBoard(agent, adversary); score != 0 {
			return agent | (0b1 << move)
		}
		adversary ^= 0b1 << move
	}
	return getRandomMove(agent, adversary)
}

func getRandomMove(agent, adversary uint16) uint16 {
	move := uint16(10000)
	board := agent | adversary

	for move >= N*N || (board|(0b1<<move)) == board {
		move = uint16(rand.Intn(int(N * N)))
	}

	return (agent | 0b1<<move)
}

func printBoard(xMoves, oMoves uint16) {
	for i := range N {
		for j := range N {
			idx := uint16(0b1 << (i*N + j))
			if xMoves&idx != 0 {
				fmt.Print("X ")
			} else if oMoves&idx != 0 {
				fmt.Print("O ")
			} else {
				fmt.Print("- ")
			}
		}
		fmt.Println("")
	}
	fmt.Println("")
}

func main() {
	var xTurn bool
	var n int

	flag.BoolVar(&xTurn, "a", false, "Does the agent start?")
	flag.BoolVar(&isGameDesperate, "l", false, "Is the agent trying not to lose? (mutually exclusive with -l)")
	flag.BoolVar(&isGameRandom, "r", false, "Is the agent random? (mutually exclusive with -r)")
	flag.IntVar(&n, "n", 3, "Board size")
	flag.Parse()

	if isGameDesperate && isGameRandom {
		panic("Options -l and -r are mutually exclusive!!")
	}

	if n == 3 {
		OUT_OF_BOUNDS = OUT_OF_BOUNDS_3
		FULL_BOARD = FULL_BOARD_3
		WINNING_PATTERNS = WINNING_PATTERNS_3
		N = 3
	} else if n == 4 {
		OUT_OF_BOUNDS = OUT_OF_BOUNDS_4
		FULL_BOARD = FULL_BOARD_4
		WINNING_PATTERNS = WINNING_PATTERNS_4
		N = 4
	} else {
		panic("Only n=3,4 are supported!")
	}

	xTurn = !xTurn

	var player1, player2 uint16
	printBoard(player1, player2)
	init_zobrist()

	for !isFull(player1 | player2) {
		if xTurn {
			player1 = playerMove(player1, player2)
		} else {
			start := time.Now()
			player2 = agentMove(player2, player1)
			elapsed := time.Since(start)
			fmt.Printf("Time to think: %v\n", elapsed)
		}
		printBoard(player1, player2)

		if score := evalBoard(player1, player2); score != 0 {
			if score > 0 {
				fmt.Print("\nX Wins!\n")
			} else {
				fmt.Printf("\nO Wins!\n")
			}
			break
		}

		xTurn = !xTurn
	}
}
