package output

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
)

const lockPath = ".cress/writer.lock"

// ErrBusy identifies contention with another cooperating writer.
var ErrBusy = errors.New("site writer is busy")

type siteLock struct {
	root *os.Root
	file *os.File
}

func acquireSiteLock(ctx context.Context, site string) (_ *siteLock, acquireErr error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	canonical, err := filepath.EvalSymlinks(site)
	if err != nil {
		return nil, fmt.Errorf("resolve writer site: %w", err)
	}
	root, err := os.OpenRoot(canonical)
	if err != nil {
		return nil, fmt.Errorf("open writer site: %w", err)
	}
	lock := &siteLock{root: root}
	defer func() {
		if acquireErr != nil {
			acquireErr = errors.Join(acquireErr, lock.Close())
		}
	}()
	if err := ensurePrivate(root, stateDirectory); err != nil {
		return nil, err
	}
	if info, err := root.Lstat(lockPath); err == nil && !info.Mode().IsRegular() {
		return nil, fmt.Errorf("writer lock must be a regular file, not a symlink or special file")
	} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("inspect writer lock: %w", err)
	}
	lock.file, err = root.OpenFile(lockPath, os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, privateFileMode)
	if err != nil {
		return nil, fmt.Errorf("open writer lock: %w", err)
	}
	info, err := lock.file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("writer lock must be a regular file: %w", errors.Join(err, fs.ErrInvalid))
	}
	if err := syscall.Flock(int(lock.file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return nil, fmt.Errorf("%w; another Cress build holds %s", ErrBusy, lockPath)
		}
		return nil, fmt.Errorf("acquire site writer lock: %w", err)
	}
	return lock, nil
}

func (lock *siteLock) Close() error {
	var err error
	if lock.root != nil {
		err = lock.root.Close()
		lock.root = nil
	}
	if lock.file != nil {
		err = errors.Join(err, lock.file.Close())
		lock.file = nil
	}
	if err != nil {
		return fmt.Errorf("close site writer lock: %w", err)
	}
	return nil
}
