package output

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
)

const (
	sessionRecord = "session.json"
	sessionLock   = "session.lock"
	sessionPrefix = "session-"
)

type sessionIdentity struct {
	Version int    `json:"version"`
	Site    string `json:"site"`
	Token   string `json:"token"`
}

func (identity sessionIdentity) matches(site, token string) bool {
	return identity.Version == stateVersion && identity.Site == site &&
		validToken(identity.Token) && identity.Token == token
}

// Session owns one private temporary target and its lifetime advisory lock.
type Session struct {
	directory string
	site      string
	token     string
	lease     *os.File
}

// NewSession reclaims verified inactive sessions without inspecting deployable output.
func NewSession(ctx context.Context, site string) (created *Session, warnings []string, sessionErr error) {
	defer func() {
		if sessionErr != nil && created != nil {
			sessionErr = errors.Join(sessionErr, created.lease.Close())
			created = nil
		}
	}()
	site, err := filepath.Abs(site)
	if err != nil {
		return nil, nil, err
	}
	site, err = filepath.EvalSymlinks(site)
	if err != nil {
		return nil, nil, err
	}
	lock, err := acquireSiteLock(ctx, site)
	if err != nil {
		return nil, nil, err
	}
	defer func() { sessionErr = errors.Join(sessionErr, lock.Close()) }()
	base, err := os.OpenRoot(os.TempDir())
	if err != nil {
		return nil, nil, fmt.Errorf("open temporary storage: %w", err)
	}
	defer func() { sessionErr = errors.Join(sessionErr, base.Close()) }()
	hash := sha256.Sum256([]byte(site))
	registryName := fmt.Sprintf("cress-preview-%d-%x", os.Getuid(), hash)
	if err := ensurePrivate(base, registryName); err != nil {
		return nil, nil, err
	}
	registry, err := base.OpenRoot(registryName)
	if err != nil {
		return nil, nil, err
	}
	defer func() { sessionErr = errors.Join(sessionErr, registry.Close()) }()
	entries, err := fs.ReadDir(registry.FS(), ".")
	if err != nil {
		return nil, nil, err
	}
	for _, entry := range entries {
		candidate, err := openSession(registry, entry.Name(), site)
		if errors.Is(err, ErrBusy) {
			continue
		}
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("preserved preview evidence %s: %v", filepath.Join(registry.Name(), entry.Name()), err))
			continue
		}
		cleanupErr := candidate.remove()
		closeErr := candidate.lease.Close()
		if err := errors.Join(cleanupErr, closeErr); err != nil {
			return nil, warnings, fmt.Errorf("reclaim inactive preview %s: %w", candidate.directory, err)
		}
	}
	token := make([]byte, tokenBytes)
	if _, err := rand.Read(token); err != nil {
		return nil, warnings, err
	}
	name := sessionPrefix + hex.EncodeToString(token)
	if err := registry.Mkdir(name, privateMode); err != nil {
		return nil, warnings, err
	}
	session := &Session{directory: filepath.Join(registry.Name(), name), site: site, token: hex.EncodeToString(token)}
	root, err := registry.OpenRoot(name)
	if err != nil {
		return nil, warnings, fmt.Errorf("open new preview session %s: %w", session.directory, err)
	}
	defer func() { sessionErr = errors.Join(sessionErr, root.Close()) }()
	data, err := json.Marshal(sessionIdentity{Version: stateVersion, Site: site, Token: session.token})
	if err != nil {
		return nil, warnings, err
	}
	if err := root.WriteFile(sessionRecord, data, privateFileMode); err != nil {
		return nil, warnings, fmt.Errorf("record preview session (preserved %s): %w", session.directory, err)
	}
	session.lease, err = root.OpenFile(sessionLock, os.O_CREATE|os.O_EXCL|os.O_RDWR, privateFileMode)
	if err == nil {
		err = syscall.Flock(int(session.lease.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	}
	if err != nil {
		if session.lease != nil {
			err = errors.Join(err, session.lease.Close())
		}
		return nil, warnings, fmt.Errorf("lock new preview session (preserved %s): %w", session.directory, err)
	}
	return session, warnings, nil
}

func openSession(registry *os.Root, name, site string) (opened *Session, openErr error) {
	defer func() {
		if openErr != nil && opened != nil {
			openErr = errors.Join(openErr, opened.lease.Close())
			opened = nil
		}
	}()
	if len(name) != len(sessionPrefix)+tokenBytes*2 || name[:len(sessionPrefix)] != sessionPrefix ||
		!validToken(name[len(sessionPrefix):]) {
		return nil, fmt.Errorf("unrecognized session name")
	}
	info, err := registry.Lstat(name)
	if err != nil || !info.IsDir() || info.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf("session is not a private real directory: %w", errors.Join(err, fs.ErrInvalid))
	}
	root, err := registry.OpenRoot(name)
	if err != nil {
		return nil, err
	}
	defer func() { openErr = errors.Join(openErr, root.Close()) }()
	writer := newWriter(root)
	data, err := writer.readRegular(sessionRecord)
	if err != nil {
		return nil, err
	}
	var identity sessionIdentity
	if err := json.Unmarshal(data, &identity); err != nil || !identity.matches(site, name[len(sessionPrefix):]) {
		return nil, fmt.Errorf("session identity does not match site and directory")
	}
	info, err = root.Lstat(sessionLock)
	if err != nil || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("session lease is not a regular file: %w", errors.Join(err, fs.ErrInvalid))
	}
	lease, err := root.OpenFile(sessionLock, os.O_RDWR|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, privateFileMode)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(lease.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			err = ErrBusy
		}
		return nil, errors.Join(err, lease.Close())
	}
	return &Session{directory: filepath.Join(registry.Name(), name), site: site, token: identity.Token, lease: lease}, nil
}
