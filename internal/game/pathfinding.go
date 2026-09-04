package game

import (
	"container/heap"
	"math"
)

const (
	PathCellSize  = 16.0
	PathGridCols  = 125 // 2000 / 16
	PathGridRows  = 125 // 2000 / 16
	MaxAStarNodes = 3000
)

// Point represents a 2D position in world coordinates
type Point struct {
	X float64
	Y float64
}

// PathGrid holds pre-computed impassable cells for the entire map
type PathGrid struct {
	cols    int
	rows    int
	blocked []bool
}

// InvalidatePathGrid marks the cached grid as dirty so it will be rebuilt on the next query
func (g *Game) InvalidatePathGrid() {
	g.basePathGrid = nil
}

// EnsureBasePathGrid builds or returns the cached obstacle grid
func (g *Game) EnsureBasePathGrid() *PathGrid {
	if g.basePathGrid != nil {
		return g.basePathGrid
	}

	grid := &PathGrid{
		cols:    PathGridCols,
		rows:    PathGridRows,
		blocked: make([]bool, PathGridCols*PathGridRows),
	}

	// 1. Mark borders where a 32x32 unit's top-left would be out of map bounds
	maxCol := int(math.Floor((MapWidth - unitSize) / PathCellSize))
	maxRow := int(math.Floor((MapHeight - unitSize) / PathCellSize))
	for r := 0; r < grid.rows; r++ {
		for c := 0; c < grid.cols; c++ {
			if c > maxCol || r > maxRow {
				grid.blocked[r*grid.cols+c] = true
			}
		}
	}

	// 2. Mark Buildings with C-space inflation for unitSize
	for _, b := range g.buildings {
		startCol := int(math.Floor((b.x - unitSize + 1.0) / PathCellSize))
		endCol := int(math.Ceil((b.x + b.width - 1.0) / PathCellSize))
		startRow := int(math.Floor((b.y - unitSize + 1.0) / PathCellSize))
		endRow := int(math.Ceil((b.y + b.height - 1.0) / PathCellSize))

		if startCol < 0 {
			startCol = 0
		}
		if startRow < 0 {
			startRow = 0
		}
		if endCol >= grid.cols {
			endCol = grid.cols - 1
		}
		if endRow >= grid.rows {
			endRow = grid.rows - 1
		}

		for r := startRow; r <= endRow; r++ {
			rowOffset := r * grid.cols
			for c := startCol; c <= endCol; c++ {
				grid.blocked[rowOffset+c] = true
			}
		}
	}

	// 3. Mark Resource Nodes with C-space inflation
	for _, res := range g.resourceNodes {
		if res.amount <= 0 {
			continue
		}
		startCol := int(math.Floor((res.x - unitSize + 1.0) / PathCellSize))
		endCol := int(math.Ceil((res.x + res.width - 1.0) / PathCellSize))
		startRow := int(math.Floor((res.y - unitSize + 1.0) / PathCellSize))
		endRow := int(math.Ceil((res.y + res.height - 1.0) / PathCellSize))

		if startCol < 0 {
			startCol = 0
		}
		if startRow < 0 {
			startRow = 0
		}
		if endCol >= grid.cols {
			endCol = grid.cols - 1
		}
		if endRow >= grid.rows {
			endRow = grid.rows - 1
		}

		for r := startRow; r <= endRow; r++ {
			rowOffset := r * grid.cols
			for c := startCol; c <= endCol; c++ {
				grid.blocked[rowOffset+c] = true
			}
		}
	}

	g.basePathGrid = grid
	return grid
}

// isWalkable checks if a cell is clear for a unit, optionally ignoring a target building or resource
func (g *Game) isWalkable(col, row int, ignoreBuilding *Building, ignoreNode *ResourceNode) bool {
	if col < 0 || col >= PathGridCols || row < 0 || row >= PathGridRows {
		return false
	}

	grid := g.EnsureBasePathGrid()
	idx := row*PathGridCols + col
	if !grid.blocked[idx] {
		return true
	}

	// Check if cell is only blocked due to the ignored target building
	if ignoreBuilding != nil {
		bx1 := ignoreBuilding.x - unitSize + 1.0
		bx2 := ignoreBuilding.x + ignoreBuilding.width - 1.0
		by1 := ignoreBuilding.y - unitSize + 1.0
		by2 := ignoreBuilding.y + ignoreBuilding.height - 1.0

		cellX := float64(col) * PathCellSize
		cellY := float64(row) * PathCellSize

		if cellX >= bx1 && cellX <= bx2 && cellY >= by1 && cellY <= by2 {
			// Ensure it doesn't overlap any OTHER building or resource
			if !g.isCollidingWithOthers(cellX, cellY, ignoreBuilding, ignoreNode) {
				return true
			}
		}
	}

	// Check if cell is only blocked due to the ignored target resource node
	if ignoreNode != nil {
		nx1 := ignoreNode.x - unitSize + 1.0
		nx2 := ignoreNode.x + ignoreNode.width - 1.0
		ny1 := ignoreNode.y - unitSize + 1.0
		ny2 := ignoreNode.y + ignoreNode.height - 1.0

		cellX := float64(col) * PathCellSize
		cellY := float64(row) * PathCellSize

		if cellX >= nx1 && cellX <= nx2 && cellY >= ny1 && cellY <= ny2 {
			if !g.isCollidingWithOthers(cellX, cellY, ignoreBuilding, ignoreNode) {
				return true
			}
		}
	}

	return false
}

// isCollidingWithOthers verifies if a position collides with entities other than the ignored ones
func (g *Game) isCollidingWithOthers(ux, uy float64, ignoreBuilding *Building, ignoreNode *ResourceNode) bool {
	for _, b := range g.buildings {
		if b == ignoreBuilding {
			continue
		}
		if ux+unitSize > b.x && ux < b.x+b.width && uy+unitSize > b.y && uy < b.y+b.height {
			return true
		}
	}
	for _, res := range g.resourceNodes {
		if res == ignoreNode || res.amount <= 0 {
			continue
		}
		if ux+unitSize > res.x && ux < res.x+res.width && uy+unitSize > res.y && uy < res.y+res.height {
			return true
		}
	}
	return false
}

// findNearestWalkableCell searches around a cell if it is blocked
func (g *Game) findNearestWalkableCell(targetCol, targetRow int, ignoreBuilding *Building, ignoreNode *ResourceNode) (int, int) {
	if g.isWalkable(targetCol, targetRow, ignoreBuilding, ignoreNode) {
		return targetCol, targetRow
	}

	for radius := 1; radius <= 12; radius++ {
		for dc := -radius; dc <= radius; dc++ {
			for dr := -radius; dr <= radius; dr++ {
				if math.Abs(float64(dc)) != float64(radius) && math.Abs(float64(dr)) != float64(radius) {
					continue
				}
				c := targetCol + dc
				r := targetRow + dr
				if g.isWalkable(c, r, ignoreBuilding, ignoreNode) {
					return c, r
				}
			}
		}
	}
	return targetCol, targetRow
}

// A* Node priority queue
type pathNode struct {
	col, row int
	g, f     float64
	parent   *pathNode
	index    int
}

type priorityQueue []*pathNode

func (pq priorityQueue) Len() int           { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool { return pq[i].f < pq[j].f }
func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}
func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*pathNode)
	item.index = n
	*pq = append(*pq, item)
}
func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

// octileDistance heuristic for 8-directional movement
func octileDistance(c1, r1, c2, r2 int) float64 {
	dx := math.Abs(float64(c1 - c2))
	dy := math.Abs(float64(r1 - r2))
	return (dx + dy) + (1.41421356-2.0)*math.Min(dx, dy)
}

// FindPath executes an A* search on the grid and applies line-of-sight smoothing
func (g *Game) FindPath(startX, startY, targetX, targetY float64, ignoreBuilding *Building, ignoreNode *ResourceNode) []Point {
	// 1. Direct Line of Sight check
	startPt := Point{X: startX, Y: startY}
	targetPt := Point{X: targetX, Y: targetY}
	if g.hasLineOfSight(startPt, targetPt, ignoreBuilding, ignoreNode) {
		return []Point{targetPt}
	}

	startCol := int(math.Round(startX / PathCellSize))
	startRow := int(math.Round(startY / PathCellSize))
	goalCol := int(math.Round(targetX / PathCellSize))
	goalRow := int(math.Round(targetY / PathCellSize))

	startCol, startRow = g.findNearestWalkableCell(startCol, startRow, ignoreBuilding, ignoreNode)
	goalCol, goalRow = g.findNearestWalkableCell(goalCol, goalRow, ignoreBuilding, ignoreNode)

	if startCol == goalCol && startRow == goalRow {
		return []Point{targetPt}
	}

	// 8 directions: orthogonal (1.0) and diagonal (sqrt(2))
	dCols := [8]int{0, 0, -1, 1, -1, 1, -1, 1}
	dRows := [8]int{-1, 1, 0, 0, -1, -1, 1, 1}
	dCosts := [8]float64{1.0, 1.0, 1.0, 1.0, 1.41421356, 1.41421356, 1.41421356, 1.41421356}

	pq := make(priorityQueue, 0, 128)
	heap.Init(&pq)

	startNode := &pathNode{
		col:    startCol,
		row:    startRow,
		g:      0,
		f:      octileDistance(startCol, startRow, goalCol, goalRow),
		parent: nil,
	}
	heap.Push(&pq, startNode)

	nodeMap := make(map[int]*pathNode, 256)
	closed := make(map[int]bool, 256)
	startIndex := startRow*PathGridCols + startCol
	nodeMap[startIndex] = startNode

	var closestNode *pathNode = startNode
	minHeuristic := startNode.f

	nodesEvaluated := 0
	var finalNode *pathNode

	for pq.Len() > 0 {
		curr := heap.Pop(&pq).(*pathNode)
		currIdx := curr.row*PathGridCols + curr.col
		closed[currIdx] = true

		if curr.col == goalCol && curr.row == goalRow {
			finalNode = curr
			break
		}

		nodesEvaluated++
		if nodesEvaluated >= MaxAStarNodes {
			// Fallback: reach closest node evaluated
			finalNode = closestNode
			break
		}

		for i := 0; i < 8; i++ {
			nc := curr.col + dCols[i]
			nr := curr.row + dRows[i]

			if nc < 0 || nc >= PathGridCols || nr < 0 || nr >= PathGridRows {
				continue
			}

			nIdx := nr*PathGridCols + nc
			if closed[nIdx] {
				continue
			}

			// Diagonal movement corner check: prevent clipping through walls
			if i >= 4 {
				c1Blocked := !g.isWalkable(curr.col+dCols[i], curr.row, ignoreBuilding, ignoreNode)
				c2Blocked := !g.isWalkable(curr.col, curr.row+dRows[i], ignoreBuilding, ignoreNode)
				if c1Blocked || c2Blocked {
					continue
				}
			}

			if !g.isWalkable(nc, nr, ignoreBuilding, ignoreNode) {
				continue
			}

			tentativeG := curr.g + dCosts[i]
			neighbor, exists := nodeMap[nIdx]

			if !exists {
				h := octileDistance(nc, nr, goalCol, goalRow)
				neighbor = &pathNode{
					col:    nc,
					row:    nr,
					g:      tentativeG,
					f:      tentativeG + h,
					parent: curr,
				}
				nodeMap[nIdx] = neighbor
				heap.Push(&pq, neighbor)

				if h < minHeuristic {
					minHeuristic = h
					closestNode = neighbor
				}
			} else if tentativeG < neighbor.g {
				neighbor.g = tentativeG
				neighbor.f = tentativeG + octileDistance(nc, nr, goalCol, goalRow)
				neighbor.parent = curr
				heap.Fix(&pq, neighbor.index)
			}
		}
	}

	if finalNode == nil {
		finalNode = closestNode
	}

	// Reconstruct grid path
	var gridPath []Point
	curr := finalNode
	for curr != nil {
		gridPath = append(gridPath, Point{
			X: float64(curr.col) * PathCellSize,
			Y: float64(curr.row) * PathCellSize,
		})
		curr = curr.parent
	}

	// Reverse path from start to goal
	for i, j := 0, len(gridPath)-1; i < j; i, j = i+1, j-1 {
		gridPath[i], gridPath[j] = gridPath[j], gridPath[i]
	}

	// Prepend starting point and append target point
	fullPath := make([]Point, 0, len(gridPath)+2)
	fullPath = append(fullPath, startPt)
	fullPath = append(fullPath, gridPath...)
	fullPath = append(fullPath, targetPt)

	// Smooth path using line-of-sight raycasting
	smoothed := g.smoothPath(fullPath, ignoreBuilding, ignoreNode)

	// Remove start point so waypoints only include upcoming milestones
	if len(smoothed) > 1 && distance(smoothed[0].X, smoothed[0].Y, startX, startY) < 4.0 {
		smoothed = smoothed[1:]
	}

	return smoothed
}

// hasLineOfSight performs raycasting to check for unobstructed line between two points
func (g *Game) hasLineOfSight(p1, p2 Point, ignoreBuilding *Building, ignoreNode *ResourceNode) bool {
	dx := p2.X - p1.X
	dy := p2.Y - p1.Y
	dist := math.Hypot(dx, dy)
	if dist < PathCellSize {
		return true
	}

	// Sample every half-cell along the vector
	steps := int(math.Ceil(dist / (PathCellSize * 0.5)))
	for s := 1; s < steps; s++ {
		t := float64(s) / float64(steps)
		sx := p1.X + dx*t
		sy := p1.Y + dy*t
		col := int(math.Round(sx / PathCellSize))
		row := int(math.Round(sy / PathCellSize))
		if !g.isWalkable(col, row, ignoreBuilding, ignoreNode) {
			return false
		}
	}
	return true
}

// smoothPath collapses intermediate waypoints when direct line of sight exists
func (g *Game) smoothPath(raw []Point, ignoreBuilding *Building, ignoreNode *ResourceNode) []Point {
	if len(raw) <= 2 {
		return raw
	}

	smoothed := make([]Point, 0, len(raw))
	smoothed = append(smoothed, raw[0])

	curr := 0
	for curr < len(raw)-1 {
		next := len(raw) - 1
		for next > curr+1 {
			if g.hasLineOfSight(raw[curr], raw[next], ignoreBuilding, ignoreNode) {
				break
			}
			next--
		}
		smoothed = append(smoothed, raw[next])
		curr = next
	}

	return smoothed
}
