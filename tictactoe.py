import itertools
import functools
import operator
import time

m = 3
n = 3
maxDepth = m*4

tileVals = {
    'x': 1,
    'o': -1,
    'blank': 0
}

def printBoard(board):
    for i in range(n):
        for j in range(n):
            if board[n*i+j] == tileVals['x']:
                print(f'X ', end='')
            elif board[n*i+j] == tileVals['o']:
                print(f'O ', end='')
            else:
                print(f'- ', end='')
        print('')

winningCombs = []

def whoWins(board):
    for comb in winningCombs:
        firstTile = board[comb[0]]
        if firstTile != tileVals['blank']:
            cnt = 1
            for i in range(1, m):
                if board[comb[i]] == firstTile:
                    cnt += 1
                if cnt == m:
                    return firstTile

    return tileVals['blank']

def generateWinningPositions():
    winningCombinations = []
    for combination in itertools.combinations(list(range(n**2)), m):
        # Row
        cnt = 1
        oldJ = combination[0]
        if combination[0] // n == combination[-1] // n:
            for j in range(1, m):
                if combination[j] - oldJ != 1:
                    cnt = 1
                else:
                    cnt += 1
                if cnt >= m:
                    # print('row', combination)
                    winningCombinations.append(combination)
                oldJ = combination[j]

        # Col
        cnt = 1
        oldJ = combination[0]
        for j in range(1, m):
            if combination[j] - oldJ != n:
                cnt = 1
            else:
                cnt += 1
            if cnt >= m:
                winningCombinations.append(combination)
            oldJ = combination[j]

        # Diag Down
        cnt = 1
        oldJ = combination[0]
        if combination[0] % n < m - 1:
            for j in range(1, m):
                if combination[j] - oldJ != n + 1:
                    cnt = 1
                else:
                    cnt += 1
                if cnt >= m:
                    winningCombinations.append(combination)
                oldJ = combination[j]

        # Diag Up
        cnt = 1
        oldJ = combination[0]
        if combination[0] % n >= m - 1:
            for j in range(1, m):
                if combination[j] - oldJ != n - 1:
                    cnt = 1
                else:
                    cnt += 1
                if cnt >= m:
                    winningCombinations.append(combination)
                oldJ = combination[j]

    return winningCombinations

def playerMove():
    idx = int(input("Tile: "))
    while board[idx] != tileVals['blank'] and 0 <= idx < n**2:
        idx = int(input("Tile: "))
    board[idx] = tileVals['x']

def agentMove():
    bestMove = (-1000, -1000)
    for i in range(0, n**2):
        # print('a')
        if board[i] == tileVals['blank']:
            board[i] = tileVals['o']
            currVal = minimax(board, -1000, 1000, False)
            # print(currVal)
            board[i] = tileVals['blank']
            if currVal > bestMove[1]:
                bestMove = (i, currVal)

    # print(bestMove)
    board[bestMove[0]] = tileVals['o']

def isBoardFull(board):
    return functools.reduce(operator.mul, board) != 0

def minimax(board, alpha, beta, maximizing, depth = 0):
    # Al agente (maximizador), nunca le va a ir peor que alpha, ni mejor
    # que beta

    # El valor que devuelva minimax nunca es menor a alpha ni mayor a beta

    # Alpha es el menor valor que puede tomar (minimax es quien toma el valor)
    # y beta es el mayor valor que puede tomar
    winner = whoWins(board)
    if winner == tileVals['o']:
        return 100 - depth
    elif winner == tileVals['x']:
        return -100 + depth
    elif depth >= maxDepth or isBoardFull(board):
        return 0

    # Lo que me imagino que yo puedo hacer
    if maximizing:
        val = -1000
        for i in range(n**2):
            if board[i] == tileVals['blank']:
                board[i] = tileVals['o']
                val = max(val, minimax(board, alpha, beta, False, depth + 1))
                board[i] = tileVals['blank']
                alpha = max(alpha, val)
                if val >= beta:
                    break
        return val
    # Lo que me imagino que me van a hacer
    else:
        val = 10000
        for i in range(n**2):
            if board[i] == tileVals['blank']:
                board[i] = tileVals['x']
                val = min(val, minimax(board, alpha, beta, True, depth + 1))
                board[i] = tileVals['blank']
                beta = min(beta, val)
                if val <= alpha:
                    break
        return val

# -1: x
#  1: o

# board = [tileVals['x']] + [tileVals['blank']] * (n**2 - 1)
board = [tileVals['blank']] * (n**2)
winningCombs =generateWinningPositions()
# print(winningCombs)

turno = tileVals['o']
for i in range(n**2):
    if turno == tileVals['x']:
        playerMove()
        turno = tileVals['o']
    else:
        start = time.perf_counter()
        agentMove()
        end = time.perf_counter()
        print('Time', end-start)
        turno = tileVals['x']

    printBoard(board)
    print()
    winner = whoWins(board)
    if whoWins(board) != tileVals['blank']:
        print(f'Winner: {winner}')
        break
