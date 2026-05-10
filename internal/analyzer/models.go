package analyzer

// NodeKind — тип узла в графе
type NodeKind string

const (
	NodeKindPackage  NodeKind = "package"
	NodeKindFile     NodeKind = "file"
	NodeKindFunction NodeKind = "function"
)

// EdgeKind — тип связи между узлами
type EdgeKind string

const (
	EdgeKindImport EdgeKind = "import"
	EdgeKindCall   EdgeKind = "call"
)

// Node — узел графа (пакет, файл или функция)
type Node struct {
	ID        int64
	RepoPath  string
	Kind      NodeKind
	Name      string
	FilePath  string
	StartLine int
	EndLine   int
}

// Edge — связь между двумя узлами
type Edge struct {
	ID       int64
	RepoPath string
	FromID   int64
	ToID     int64
	Kind     EdgeKind
}

// Graph — результат анализа репозитория
type Graph struct {
	Nodes []*Node
	Edges []*Edge
}