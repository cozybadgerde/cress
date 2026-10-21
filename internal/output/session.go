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
	"slices"
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
	writer, err := acquireSite(ctx, site)
	if err != nil {
		return nil, nil, err
	}
	defer func() { sessionErr = errors.Join(sessionErr, writer.Close()) }()
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
	writer := &Writer{root: root}
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

// Site reports the canonical source whose writer lock protects this session.
func (session *Session) Site() string { return session.site }

// Directory reports the private target, not the deployable site directory.
func (session *Session) Directory() string { return session.directory }

// Attempt holds the shared site writer lock while generating into this session only.
func (session *Session) Attempt(ctx context.Context, generate func(*Writer) error) (token string, attemptErr error) {
	if session.lease == nil {
		return "", fmt.Errorf("preview session is closed")
	}
	writer, err := acquireSite(ctx, session.site)
	if err != nil {
		token, verifyErr := session.Verify()
		return token, errors.Join(err, verifyErr)
	}
	defer func() { attemptErr = errors.Join(attemptErr, writer.Close()) }()
	root, err := session.openRoot()
	if err != nil {
		return "", err
	}
	if err := writer.root.Close(); err != nil {
		return "", errors.Join(err, root.Close())
	}
	writer.root = root
	writer.rename, writer.removeAll = root.Rename, root.RemoveAll
	if err := ensurePrivate(root, stateDirectory); err != nil {
		return "", err
	}
	if err := writer.loadState(); err != nil {
		return "", fmt.Errorf("preview target %s: %w", session.directory, writer.guidance(err))
	}
	attemptErr = writer.recover()
	if attemptErr == nil {
		attemptErr = generate(writer)
	}
	token, verifyErr := writer.verifiedToken()
	if err := errors.Join(attemptErr, verifyErr); err != nil {
		return token, fmt.Errorf("preview target %s: %w", session.directory, err)
	}
	return token, nil
}

func (writer *Writer) verifiedToken() (string, error) {
	if err := writer.state.validate(); err != nil {
		return "", err
	}

	token, err := writer.inspectOutput("public")
	if err != nil {
		return "", err
	}
	if validToken(token) && token == writer.state.Current {
		return token, nil
	}
	return "", fmt.Errorf("no verified usable preview output; inspect %s", writer.root.Name())
}

// Verify checks usable session output without mutation or deployable-output access.
func (session *Session) Verify() (token string, verifyErr error) {
	root, err := session.openRoot()
	if err != nil {
		return "", err
	}
	defer func() { verifyErr = errors.Join(verifyErr, root.Close()) }()
	info, err := root.Lstat(stateDirectory)
	if err != nil || !info.IsDir() || info.Mode().Perm()&0o077 != 0 {
		return "", fmt.Errorf("preview state is not a private real directory: %w", errors.Join(err, fs.ErrInvalid))
	}
	writer := &Writer{root: root}
	if err := writer.loadState(); err != nil {
		return "", err
	}
	return writer.verifiedToken()
}

// Close removes only verified session output and releases the lifetime lock.
func (session *Session) Close() (closeErr error) {
	if session.lease == nil {
		return nil
	}
	defer func() {
		closeErr = errors.Join(closeErr, session.lease.Close())
		session.lease = nil
	}()
	writer, err := acquireSite(context.Background(), session.site)
	if err != nil {
		return fmt.Errorf("preview cleanup blocked; preserved %s: %w", session.directory, err)
	}
	defer func() { closeErr = errors.Join(closeErr, writer.Close()) }()
	return session.remove()
}

func (session *Session) remove() (removeErr error) {
	root, err := session.openRoot()
	if err != nil {
		return err
	}
	defer func() { removeErr = errors.Join(removeErr, root.Close()) }()
	entries, err := fs.ReadDir(root.FS(), ".")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !slices.Contains([]string{sessionRecord, sessionLock, stateDirectory, "public"}, entry.Name()) {
			return fmt.Errorf("preserved %s: unrecognized file %s", session.directory, entry.Name())
		}
	}
	writer := &Writer{root: root, rename: root.Rename, removeAll: root.RemoveAll}
	if _, err := root.Lstat(stateDirectory); err == nil {
		if err := ensurePrivate(root, stateDirectory); err != nil {
			return err
		}
		if err := writer.prepare(); err != nil {
			return fmt.Errorf("preserved ambiguous preview %s: %w", session.directory, err)
		}
		if _, err := root.Lstat("public"); err == nil {
			if _, err := writer.verifiedToken(); err != nil {
				return err
			}
		} else if !errors.Is(err, fs.ErrNotExist) {
			return err
		} else if writer.state.Current != "" {
			return fmt.Errorf("preserved %s: previously owned output is missing", session.directory)
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	} else if _, err := root.Lstat("public"); !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("preserved unowned preview output %s", session.directory)
	}
	parent, err := os.OpenRoot(filepath.Dir(session.directory))
	if err != nil {
		return err
	}
	defer func() { removeErr = errors.Join(removeErr, parent.Close()) }()
	if err := parent.RemoveAll(filepath.Base(session.directory)); err != nil {
		return fmt.Errorf("cleanup preview %s: %w", session.directory, err)
	}
	return nil
}

func (session *Session) openRoot() (_ *os.Root, openErr error) {
	info, err := os.Lstat(session.directory)
	if err != nil || !info.IsDir() || info.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf("preview target must remain a private real directory: %w", errors.Join(err, fs.ErrInvalid))
	}
	root, err := os.OpenRoot(session.directory)
	if err != nil {
		return nil, err
	}
	defer func() {
		if openErr != nil {
			openErr = errors.Join(openErr, root.Close())
		}
	}()
	writer := &Writer{root: root}
	data, err := writer.readRegular(sessionRecord)
	if err != nil {
		return nil, err
	}
	var identity sessionIdentity
	if err := json.Unmarshal(data, &identity); err != nil || !identity.matches(session.site, session.token) {
		return nil, fmt.Errorf("preview session identity changed; preserved %s", session.directory)
	}
	if session.lease == nil {
		return nil, fmt.Errorf("preview session is closed")
	}
	leaseInfo, err := session.lease.Stat()
	if err != nil {
		return nil, err
	}
	info, err = root.Lstat(sessionLock)
	if err != nil || !info.Mode().IsRegular() || !os.SameFile(info, leaseInfo) {
		return nil, fmt.Errorf("preview lifetime lock changed; preserved %s", session.directory)
	}
	return root, nil
}
