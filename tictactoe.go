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

const (
	OUT_OF_BOUNDS uint16 = 0b1111111000000000
	FULL_BOARD    uint16 = 0b0000000111111111
	COL_1         uint16 = 0b0000000100100100
	COL_2         uint16 = 0b0000000010010010
	COL_3         uint16 = 0b0000000001001001
	ROW_1         uint16 = 0b0000000111000000
	ROW_2         uint16 = 0b0000000000111000
	ROW_3         uint16 = 0b0000000000000111
	DIAG_UP       uint16 = 0b0000000001010100
	DIAG_DOWN     uint16 = 0b0000000100010001
)

const N uint16 = 3

var WINNING_PATTERNS = []uint16{
	COL_1, COL_2, COL_3, ROW_1, ROW_2, ROW_3, DIAG_DOWN, DIAG_UP,
}

var zobrist_table = make([]uint64, 2*N*N)
var tTable = NewTranspositionTable(8192)

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

type TtNode struct {
	entry TtEntry
	next  *TtNode
}

type TranspositionTable struct {
	buckets []*TtNode
	size    uint64
}

func NewTranspositionTable(size uint64) *TranspositionTable {
	return &TranspositionTable{
		buckets: make([]*TtNode, size),
		size:    size,
	}
}

func (tt *TranspositionTable) store(entry TtEntry) {
	index := entry.zobrist % tt.size
	node := &TtNode{entry: entry}

	if tt.buckets[index] == nil {
		tt.buckets[index] = node
	} else {
		curr := tt.buckets[index]
		for curr.next != nil {
			if curr.entry.zobrist == entry.zobrist {
				curr.entry = entry
				return
			}
			curr = curr.next
		}
		curr.next = node
	}
}

func (tt *TranspositionTable) lookup(zobrist uint64) *TtEntry {
	index := zobrist % tt.size
	if tt.buckets[index] != nil {
		curr := tt.buckets[index]
		for curr.next != nil {
			if curr.entry.zobrist == zobrist {
				return &curr.entry
			}
			curr = curr.next
		}
	}

	return nil
}

func (tt *TranspositionTable) zobrist_hash(xMoves, oMoves uint16) uint64 {
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

func negamax(agent, adversary uint16, depth uint16, alpha, beta int16, marker bool) int16 {
	alphaOrig := alpha
	var ttEntry *TtEntry
	if marker {
		// ttEntry = tTable.lookup(agent, adversary)
		ttEntry = tTable.lookup(tTable.zobrist_hash(agent, adversary))
	} else {
		// ttEntry = tTable.lookup(adversary, agent)
		ttEntry = tTable.lookup(tTable.zobrist_hash(adversary, agent))
	}

	if ttEntry != nil && ttEntry.depth >= depth {
		// fmt.Println("HOLA")
		if ttEntry.flag == EXACT {
			return ttEntry.value
		} else if ttEntry.flag == LOWERBOUND {
			alpha = max(alpha, ttEntry.value)
		} else if ttEntry.flag == UPPERBOUND {
			beta = min(beta, ttEntry.value)
		}

		if alpha >= beta {
			// fmt.Println("cutoff")
			return ttEntry.value
		}
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
		value = max(value, -negamax(adversary, agent, depth-1, -beta, -alpha, !marker))
		adversary ^= 0b1 << move
		alpha = max(alpha, value)
		if alpha >= beta {
			break
		}
	}

	if ttEntry == nil {
		if marker {
			// ttEntry = &TtEntry{oMoves: adversary, xMoves: agent}
			ttEntry = &TtEntry{zobrist: tTable.zobrist_hash(agent, adversary)}
		} else {
			// ttEntry = &TtEntry{oMoves: agent, xMoves: adversary}
			ttEntry = &TtEntry{zobrist: tTable.zobrist_hash(adversary, agent)}
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
	tTable.store(*ttEntry)

	return value
}

func playerMove(player, adversary uint16) uint16 {
	var idx uint16 = 10
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
		newScore := -negamax(agent, adversary, 16, -10000, 10000, false)
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
	var color int16 = -1
	printBoard(player1, player2)
	init_zobrist()
	// for _,v := range WINNING_PATTERNS {
	// 	fmt.Printf("%b, %v\n", v, v)
	// }
	fmt.Println(zobrist_table)
	for !isFull(player1 | player2) {
		if color == 1 {
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

		color *= -1
	}
	// fmt.Println(tTable)
	// for _, list := range tTable.buckets {
	// 	if list != nil {
	// 		curr := list
	// 		for curr.next != nil {
	// 			fmt.Printf("%v\n", curr.entry)
	// 			curr = curr.next
	// 		}
	// 	}
	// }
	myEntry := tTable.lookup(tTable.zobrist_hash(0, 1))
	fmt.Printf("%v\n", myEntry)
	myEntry = tTable.lookup(tTable.zobrist_hash(1, 0))
	fmt.Printf("%v\n", myEntry)
}
