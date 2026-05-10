package git

import (
	"fmt"
	"sort"
	"strings"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// FileMetrics — git метрики по одному файлу
type FileMetrics struct {
	FilePath      string
	ChurnRate     int            // сколько раз файл менялся
	Owner         string         // кто менял больше всего
	AuthorCommits map[string]int // сколько коммитов у каждого автора
}

// CouplingPair — два файла которые часто меняются вместе
type CouplingPair struct {
	FileA         string
	FileB         string
	CouplingScore float64 // 0..1, чем выше тем чаще меняются вместе
}

// Analyzer анализирует git историю репозитория
type Analyzer struct {
	repoPath string
}

func NewAnalyzer(repoPath string) *Analyzer {
	return &Analyzer{repoPath: repoPath}
}

// AnalyzeMetrics вычисляет churn rate и ownership для каждого файла
func (a *Analyzer) AnalyzeMetrics() (map[string]*FileMetrics, error) {
	repo, err := gogit.PlainOpen(a.repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repo: %w", err)
	}

	// Получаем все коммиты
	commits, err := a.getAllCommits(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}

	metrics := map[string]*FileMetrics{}

	for _, commit := range commits {
		// Получаем список файлов изменённых в этом коммите
		files, err := a.getChangedFiles(commit)
		if err != nil {
			continue
		}

		for _, filePath := range files {
			// Считаем только .go файлы
			if !strings.HasSuffix(filePath, ".go") {
				continue
			}

			if _, ok := metrics[filePath]; !ok {
				metrics[filePath] = &FileMetrics{
					FilePath:      filePath,
					AuthorCommits: map[string]int{},
				}
			}

			// Увеличиваем churn rate
			metrics[filePath].ChurnRate++

			// Считаем коммиты по авторам
			author := commit.Author.Name
			metrics[filePath].AuthorCommits[author]++
		}
	}

	// Определяем owner — тот кто сделал больше всего коммитов
	for _, m := range metrics {
		m.Owner = findOwner(m.AuthorCommits)
	}

	return metrics, nil
}

// AnalyzeCoupling определяет файлы которые всегда меняются вместе
func (a *Analyzer) AnalyzeCoupling() ([]CouplingPair, error) {
	repo, err := gogit.PlainOpen(a.repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repo: %w", err)
	}

	commits, err := a.getAllCommits(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}

	// Считаем сколько раз каждая пара файлов менялась вместе
	pairCount := map[string]int{}
	fileCount := map[string]int{}

	for _, commit := range commits {
		files, err := a.getChangedFiles(commit)
		if err != nil {
			continue
		}

		// Фильтруем только .go файлы
		goFiles := []string{}
		for _, f := range files {
			if strings.HasSuffix(f, ".go") {
				goFiles = append(goFiles, f)
				fileCount[f]++
			}
		}

		// Перебираем все пары файлов в этом коммите
		for i := 0; i < len(goFiles); i++ {
			for j := i + 1; j < len(goFiles); j++ {
				// Сортируем чтобы пара (A,B) и (B,A) были одинаковыми
				pair := makePairKey(goFiles[i], goFiles[j])
				pairCount[pair]++
			}
		}
	}

	// Вычисляем coupling score для каждой пары
	// Score = количество совместных изменений / max(изменений файла A, изменений файла B)
	var pairs []CouplingPair
	for pairKey, count := range pairCount {
		if count < 2 {
			// Игнорируем пары которые менялись вместе меньше 2 раз
			continue
		}

		parts := strings.SplitN(pairKey, "|", 2)
		if len(parts) != 2 {
			continue
		}
		fileA, fileB := parts[0], parts[1]

		maxCount := max(fileCount[fileA], fileCount[fileB])
		if maxCount == 0 {
			continue
		}

		score := float64(count) / float64(maxCount)
		pairs = append(pairs, CouplingPair{
			FileA:         fileA,
			FileB:         fileB,
			CouplingScore: score,
		})
	}

	// Сортируем по score — самые связанные файлы первыми
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].CouplingScore > pairs[j].CouplingScore
	})

	return pairs, nil
}

// getAllCommits возвращает все коммиты репозитория за последний год
func (a *Analyzer) getAllCommits(repo *gogit.Repository) ([]*object.Commit, error) {
	ref, err := repo.Head()
	if err != nil {
		return nil, err
	}

	iter, err := repo.Log(&gogit.LogOptions{
		From:  ref.Hash(),
		Since: timePtr(time.Now().AddDate(-1, 0, 0)), // последний год
	})
	if err != nil {
		return nil, err
	}

	var commits []*object.Commit
	err = iter.ForEach(func(c *object.Commit) error {
		commits = append(commits, c)
		return nil
	})

	return commits, err
}

// getChangedFiles возвращает список файлов изменённых в коммите
func (a *Analyzer) getChangedFiles(commit *object.Commit) ([]string, error) {
	// Для первого коммита нет родителя — возвращаем пустой список
	if commit.NumParents() == 0 {
		return nil, nil
	}

	parent, err := commit.Parent(0)
	if err != nil {
		return nil, err
	}

	patch, err := parent.Patch(commit)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, fp := range patch.FilePatches() {
		from, to := fp.Files()
		if to != nil {
			files = append(files, to.Path())
		} else if from != nil {
			files = append(files, from.Path())
		}
	}

	return files, nil
}

// findOwner возвращает автора с наибольшим количеством коммитов
func findOwner(authorCommits map[string]int) string {
	maxCount := 0
	owner := ""
	for author, count := range authorCommits {
		if count > maxCount {
			maxCount = count
			owner = author
		}
	}
	return owner
}

// makePairKey создаёт уникальный ключ для пары файлов
func makePairKey(a, b string) string {
	if a > b {
		a, b = b, a
	}
	return a + "|" + b
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
