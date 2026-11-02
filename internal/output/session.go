package output

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

// Site reports the canonical source whose writer lock protects this session.
func (session *Session) Site() string { return session.site }

// Directory reports the private target, not the deployable site directory.
func (session *Session) Directory() string { return session.directory }

// Attempt holds the shared site writer lock while generating into this session only.
func (session *Session) Attempt(ctx context.Context, generate func(*Writer) error) (token string, attemptErr error) {
	if session.lease == nil {
		return "", fmt.Errorf("preview session is closed")
	}
	lock, err := acquireSiteLock(ctx, session.site)
	if err != nil {
		token, verifyErr := session.Verify()
		return token, errors.Join(err, verifyErr)
	}
	defer func() { attemptErr = errors.Join(attemptErr, lock.Close()) }()
	root, err := session.openRoot()
	if err != nil {
		return "", err
	}
	writer := newWriter(root)
	defer func() { attemptErr = errors.Join(attemptErr, writer.Close()) }()
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
	writer := newWriter(root)
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
	lock, err := acquireSiteLock(context.Background(), session.site)
	if err != nil {
		return fmt.Errorf("preview cleanup blocked; preserved %s: %w", session.directory, err)
	}
	defer func() { closeErr = errors.Join(closeErr, lock.Close()) }()
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
	writer := newWriter(root)
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
	writer := newWriter(root)
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
