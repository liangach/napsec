package audit

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"os"
	"time"
)

// GitLog Git 审计日志
type GitLog struct {
	repo    *git.Repository
	repoDir string
}

// NewGitLog 初始化 Git 仓库
func NewGitLog(dir string) (*GitLog, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	// 尝试打开已有仓库
	repo, err := git.PlainOpen(dir)
	if err == git.ErrRepositoryNotExists {
		// 初始化新仓库
		repo, err = git.PlainInit(dir, false)
		if err != nil {
			return nil, fmt.Errorf("初始化 Git 仓库失败: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("打开 Git 仓库失败: %w", err)
	}

	return &GitLog{
		repo:    repo,
		repoDir: dir,
	}, nil
}

// Commit 提交审计日志
func (g *GitLog) Commit(message string) error {
	wt, err := g.repo.Worktree()
	if err != nil {
		return err
	}

	// 添加所有变更文件
	if _, err := wt.Add("."); err != nil {
		return err
	}

	// 检查是否有变更
	status, err := wt.Status()
	if err != nil {
		return err
	}
	if status.IsClean() {
		return nil // 无变更，跳过 commit
	}

	// 创建 commit
	_, err = wt.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Guardian",
			Email: "guardian@localhost",
			When:  time.Now(),
		},
	})

	return err
}
