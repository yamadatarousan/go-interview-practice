package main

import (
	"fmt"
	"sync"
)

// ConcurrentBFSQueries concurrently processes BFS queries on the provided graph.
// - graph: adjacency list, e.g., graph[u] = []int{v1, v2, ...}
// - queries: a list of starting nodes for BFS.
// - numWorkers: how many goroutines can process BFS queries simultaneously.
//
// Return a map from the query (starting node) to the BFS order as a slice of nodes.
// YOU MUST use concurrency (goroutines + channels) to pass the performance tests.
func ConcurrentBFSQueries(graph map[int][]int, queries []int, numWorkers int) map[int][]int {
	// TODO: Implement concurrency-based BFS for multiple queries.
	// Return an empty map so the code compiles but fails tests if unchanged.
	results := make(map[int][]int)
	var mu sync.Mutex // mapへの同時書き込み防止

	jobs := make(chan int) //開始ノードを流す
	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()
		for start := range jobs {
			order := bfs(graph, start) // BFSの実装は別途行う
			mu.Lock()
			results[start] = order
			mu.Unlock()
		}
	}

	if numWorkers < 1 {
		return results
	}
	// ワーカー起動（クエリより多くても意味ない）
	if numWorkers > len(queries) {
		numWorkers = len(queries)
	}
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker()
	}

	// ジョブ投入
	for _, q := range queries {
		jobs <- q
	}
	close(jobs) // ジョブの投入完了

	wg.Wait() // 全ワーカーの終了を待つ
	return results
}

// bfs はスタート頂点から到達可能な全頂点を
// 「近い順」に訪問する幅優先探索
func bfs(graph map[int][]int, start int) []int {
	// visited: すでにキューに入れた頂点を記録
	// 同じ頂点を二度入れると無限ループになるので防止する
	visited := make(map[int]bool)

	// queue: 次に調べる頂点の待ち行列
	// 先入れ先出し FIFO で「近い順」を保つ
	queue := []int{start}
	visited[start] = true // スタートは最初から訪問済み扱い

	// order: 実際に取り出した順番を記録する答え
	order := []int{}

	// キューが空になるまで繰り返す
	// 空になったら到達可能な頂点は全て処理済み
	for len(queue) > 0 {
		// --- 1. キューの先頭を取り出す (dequeue) ---
		v := queue[0]            // 一番古く入れた頂点
		queue = queue[1:]        // 先頭を切り捨てて詰める
		order = append(order, v) // 訪問順に追加

		// --- 2. v から1手で行ける隣接頂点を調べる ---
		for _, nb := range graph[v] {
			// まだ訪れていない頂点だけを対象にする
			if !visited[nb] {
				visited[nb] = true        // 二重登録防止のため先にマーク
				queue = append(queue, nb) // キューの末尾に追加 (enqueue)
				// これで nb は「vより1段階遠い」グループとして後で処理される
			}
		}
		// ループを回るたびに「距離0の層」「距離1の層」…の順で
		// 頂点が order に溜まっていく
	}

	return order
}

func main() {
	// You can insert optional local tests here if desired.
	graph := map[int][]int{
		0: {1, 2},
		1: {2, 3},
		2: {3},
		3: {4},
		4: {},
	}
	queries := []int{0, 1, 2}
	numWorkers := 2
	results := ConcurrentBFSQueries(graph, queries, numWorkers)

	for _, q := range queries {
		fmt.Printf("results[%d] = %v\n", q, results[q])
	}

	/*
	   Possible output:
	   results[0] = [0 1 2 3 4]
	   results[1] = [1 2 3 4]
	   results[2] = [2 3 4]
	*/
}
