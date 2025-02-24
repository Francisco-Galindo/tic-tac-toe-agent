package main

import (
	"fmt"
	"math/bits"
	"time"
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

var WINNING_PATTERNS = []uint16{
	COL_1, COL_2, COL_3, ROW_1, ROW_2, ROW_3, DIAG_DOWN, DIAG_UP,
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

func negamax(player, agent uint16, depth, alpha, beta int16) int16 {
	if depth == 0 || isFull(player|agent) {
		return 0
	}
	score := evalBoard(agent, player)
	if score != 0 {
		return score
	}
	val := int16(-10000)
	for _, move := range getPossibleMoves(player | agent) {
		agent |= 0b1 << move
		val = max(val, -negamax(agent, player, depth-1, -beta, -alpha))
		agent ^= 0b1 << move
		alpha = max(alpha, val)
		if alpha >= beta {
			break
		}
	}
	return val
}

func playerMove(player, adversary uint16) uint16 {
	var idx uint16 = 10
	for idx >= 3*3 {
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
		newScore := -negamax(agent, adversary, 123, -10000, 10000)
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

func printBoard(x_moves, o_moves uint16) {
	for i := range 3 {
		for j := range 3 {
			idx := uint16(0b1 << (i*3 + j))
			if x_moves&idx != 0 {
				fmt.Printf("X ")
			} else if o_moves&idx != 0 {
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
	// for _,v := range WINNING_PATTERNS {
	// 	fmt.Printf("%b, %v\n", v, v)
	// }
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
}
