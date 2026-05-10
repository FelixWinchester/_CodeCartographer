package ast

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	analyzer "github.com/FelixWinchester/CodeCartographer/internal/analyzer"
)

// Parser обходит Go-репозиторий и строит граф зависимостей.
type Parser struct {
	repoPath string
}

func NewParser(repoPath string) *Parser {
	return &Parser{repoPath: repoPath}
}

// Parse обходит все .go файлы в репозитории и возвращает граф.
func (p *Parser) Parse() (*analyzer.Graph, error) {
	graph := &analyzer.Graph{}

	// Маппинг: путь к файлу → ID узла файла в графе
	// Нужен чтобы потом строить рёбра между файлами
	fileNodeIDs := map[string]int64{}

	var nodeID int64 = 1 // временный локальный ID (до записи в Postgres)

	err := filepath.Walk(p.repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Пропускаем папки и не-.go файлы
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Пропускаем тесты (опционально, можно убрать)
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Путь относительно корня репозитория
		relPath, err := filepath.Rel(p.repoPath, path)
		if err != nil {
			return err
		}

		// Парсим файл через go/parser
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if err != nil {
			// Не останавливаемся на ошибке одного файла
			return nil
		}

		// Создаём узел для пакета
		pkgName := f.Name.Name
		pkgNode := &analyzer.Node{
			ID:       nodeID,
			RepoPath: p.repoPath,
			Kind:     analyzer.NodeKindPackage,
			Name:     pkgName,
			FilePath: relPath,
		}
		nodeID++
		graph.Nodes = append(graph.Nodes, pkgNode)

		// Создаём узел для файла
		fileNode := &analyzer.Node{
			ID:       nodeID,
			RepoPath: p.repoPath,
			Kind:     analyzer.NodeKindFile,
			Name:     filepath.Base(path),
			FilePath: relPath,
		}
		fileNodeIDs[relPath] = fileNode.ID
		nodeID++
		graph.Nodes = append(graph.Nodes, fileNode)

		// Связь: пакет → файл
		graph.Edges = append(graph.Edges, &analyzer.Edge{
			RepoPath: p.repoPath,
			FromID:   pkgNode.ID,
			ToID:     fileNode.ID,
			Kind:     analyzer.EdgeKindImport,
		})

		// Извлекаем импорты файла
		for _, imp := range f.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)

			// Создаём узел для импортируемого пакета
			importNode := &analyzer.Node{
				ID:       nodeID,
				RepoPath: p.repoPath,
				Kind:     analyzer.NodeKindPackage,
				Name:     importPath,
				FilePath: relPath,
			}
			nodeID++
			graph.Nodes = append(graph.Nodes, importNode)

			// Связь: файл → импортируемый пакет
			graph.Edges = append(graph.Edges, &analyzer.Edge{
				RepoPath: p.repoPath,
				FromID:   fileNode.ID,
				ToID:     importNode.ID,
				Kind:     analyzer.EdgeKindImport,
			})
		}

		// Извлекаем функции из файла
		for _, decl := range f.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			startLine := fset.Position(funcDecl.Pos()).Line
			endLine := fset.Position(funcDecl.End()).Line

			funcNode := &analyzer.Node{
				ID:        nodeID,
				RepoPath:  p.repoPath,
				Kind:      analyzer.NodeKindFunction,
				Name:      funcDecl.Name.Name,
				FilePath:  relPath,
				StartLine: startLine,
				EndLine:   endLine,
			}
			nodeID++
			graph.Nodes = append(graph.Nodes, funcNode)

			// Связь: файл → функция
			graph.Edges = append(graph.Edges, &analyzer.Edge{
				RepoPath: p.repoPath,
				FromID:   fileNode.ID,
				ToID:     funcNode.ID,
				Kind:     analyzer.EdgeKindCall,
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return graph, nil
}
