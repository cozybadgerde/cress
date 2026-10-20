package output

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"syscall"
)

const (
	stateDirectory  = ".cress"
	statePath       = ".cress/output.json"
	nextStatePath   = ".cress/output.next"
	lockPath        = ".cress/writer.lock"
	markerPath      = ".cress-output.json"
	privateMode     = 0o700
	privateFileMode = 0o600
	tokenBytes      = 16
	stateVersion    = 1
)

// ErrBusy identifies contention with another cooperating writer.
var ErrBusy = errors.New("site writer is busy")

type ownership struct {
	Version int    `json:"version"`
	Token   string `json:"token"`
}

type transaction struct {
	Token     string `json:"token"`
	Previous  string `json:"previous"`
	HadPublic bool   `json:"had_public"`
	Phase     string `json:"phase"`
}

type state struct {
	Version int          `json:"version"`
	Current string       `json:"current"`
	Pending *transaction `json:"pending,omitempty"`
}

// Writer holds site-wide OS ownership through preparation, publication and recovery.
type Writer struct {
	root      *os.Root
	lock      *os.File
	state     state
	rename    func(string, string) error
	removeAll func(string) error
}

// Acquire fails fast on contention and recovers only verified transactions.
func Acquire(ctx context.Context, site string) (_ *Writer, acquireErr error) {
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
	writer := &Writer{root: root, rename: root.Rename, removeAll: root.RemoveAll}
	defer func() {
		if acquireErr != nil {
			acquireErr = errors.Join(acquireErr, writer.Close())
		}
	}()
	info, err := root.Lstat(stateDirectory)
	if errors.Is(err, fs.ErrNotExist) {
		if err := root.Mkdir(stateDirectory, privateMode); err != nil && !errors.Is(err, fs.ErrExist) {
			return nil, fmt.Errorf("create private writer state: %w", err)
		}
		info, err = root.Lstat(stateDirectory)
	}
	if err != nil {
		return nil, fmt.Errorf("inspect private writer state: %w", err)
	}
	if !info.IsDir() || info.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf(".cress must be a real private directory (mode 0700), not a symlink or file")
	}
	if info, err := root.Lstat(lockPath); err == nil && !info.Mode().IsRegular() {
		return nil, fmt.Errorf("writer lock must be a regular file, not a symlink or special file")
	} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("inspect writer lock: %w", err)
	}
	writer.lock, err = root.OpenFile(lockPath, os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, privateFileMode)
	if err != nil {
		return nil, fmt.Errorf("open writer lock: %w", err)
	}
	info, err = writer.lock.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("writer lock must be a regular file: %w", errors.Join(err, fs.ErrInvalid))
	}
	if err := syscall.Flock(int(writer.lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return nil, fmt.Errorf("%w; another Cress build holds %s", ErrBusy, lockPath)
		}
		return nil, fmt.Errorf("acquire site writer lock: %w", err)
	}
	if err := writer.loadState(); err != nil {
		return nil, writer.guidance(err)
	}
	if err := writer.recover(); err != nil {
		return nil, err
	}
	return writer, nil
}

// Close releases OS ownership without deleting the persistent lock file.
func (writer *Writer) Close() error {
	var err error
	if writer.lock != nil {
		err = writer.lock.Close()
		writer.lock = nil
	}
	if writer.root != nil {
		err = errors.Join(err, writer.root.Close())
		writer.root = nil
	}
	if err != nil {
		return fmt.Errorf("close site writer: %w", err)
	}
	return nil
}

func validToken(token string) bool {
	data, err := hex.DecodeString(token)
	return err == nil && len(data) == tokenBytes && hex.EncodeToString(data) == token
}

func (writer *Writer) loadState() error {
	writer.state = state{Version: stateVersion}
	data, err := writer.readRegular(statePath)
	if err == nil {
		writer.state = state{}
		if err := json.Unmarshal(data, &writer.state); err != nil {
			return fmt.Errorf("invalid output ownership record: %w", err)
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	pending := writer.state.Pending
	if err := writer.state.validate(); err != nil {
		return err
	}
	entries, err := fs.ReadDir(writer.root.FS(), stateDirectory)
	if err != nil {
		return fmt.Errorf("inspect private output state: %w", err)
	}
	for _, entry := range entries {
		name := path.Join(stateDirectory, entry.Name())
		if name == statePath || name == lockPath {
			continue
		}
		if pending != nil && (name == pending.stage() || name == pending.backup()) {
			continue
		}
		return fmt.Errorf("unrecognized recovery evidence at %s", name)
	}
	return nil
}

func (record state) validate() error {
	if record.Version != stateVersion || (record.Current != "" && !validToken(record.Current)) {
		return fmt.Errorf("invalid output ownership version or token")
	}
	pending := record.Pending
	if pending != nil && (!validToken(pending.Token) || (pending.Previous != "" && !validToken(pending.Previous)) ||
		pending.Token == pending.Previous || (!pending.HadPublic && pending.Previous != "") ||
		(pending.Phase != "staging" && pending.Phase != "ready" && pending.Phase != "published") ||
		(pending.Phase == "published" && record.Current != pending.Token) ||
		(pending.Phase != "published" && pending.Previous != record.Current)) {
		return fmt.Errorf("invalid output transaction record")
	}
	return nil
}

func (writer *Writer) readRegular(name string) ([]byte, error) {
	info, err := writer.root.Lstat(name)
	if err != nil {
		return nil, fmt.Errorf("inspect %s: %w", name, err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s must be a regular file, not a symlink or special file", name)
	}
	data, err := writer.root.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", name, err)
	}
	return data, nil
}

func (writer *Writer) saveState() error {
	data, err := json.Marshal(writer.state)
	if err != nil {
		return fmt.Errorf("encode output state: %w", err)
	}
	file, err := writer.root.OpenFile(nextStatePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, privateFileMode)
	if err != nil {
		return fmt.Errorf("create next output state: %w", err)
	}
	_, writeErr := file.Write(append(data, '\n'))
	syncErr := file.Sync()
	if err := errors.Join(writeErr, syncErr, file.Close()); err != nil {
		return errors.Join(fmt.Errorf("write output state: %w", err), writer.root.Remove(nextStatePath))
	}
	if err := writer.rename(nextStatePath, statePath); err != nil {
		return fmt.Errorf("install output state (next state preserved): %w", err)
	}
	return nil
}

func (pending transaction) stage() string  { return stateDirectory + "/stage-" + pending.Token }
func (pending transaction) backup() string { return stateDirectory + "/backup-" + pending.Token }

func (writer *Writer) guidance(err error) error {
	paths := "public/ and .cress/"
	if pending := writer.state.Pending; pending != nil {
		paths = "public/, " + pending.stage() + "/, " + pending.backup() + "/ and " + statePath
	}
	return fmt.Errorf("%w; preserved recovery evidence: inspect %s before retrying; do not delete state or adopt unverified output", err, paths)
}

// inspectOutput returns absent, empty, or an ownership token, refusing other states.
func (writer *Writer) inspectOutput(name string) (string, error) {
	info, err := writer.root.Lstat(name)
	if errors.Is(err, fs.ErrNotExist) {
		return "absent", nil
	}
	if err != nil {
		return "", fmt.Errorf("inspect %s: %w", name, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("refuse %s: output must be a real directory, not a symlink or file", name)
	}
	entries, err := fs.ReadDir(writer.root.FS(), name)
	if err != nil {
		return "", fmt.Errorf("inspect %s contents: %w", name, err)
	}
	if len(entries) == 0 {
		return "empty", nil
	}
	data, err := writer.readRegular(path.Join(name, markerPath))
	if err != nil {
		return "", fmt.Errorf("refuse nonempty %s without verifiable Cress ownership: %w", name, err)
	}
	var marker ownership
	if err := json.Unmarshal(data, &marker); err != nil || marker.Version != stateVersion || !validToken(marker.Token) {
		return "", fmt.Errorf("refuse %s: invalid ownership marker", name)
	}
	return marker.Token, nil
}

// Write stages all files before touching public, then replaces verified output.
func (writer *Writer) Write(ctx context.Context, files []File) error {
	if err := Validate(files); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := writer.loadState(); err != nil {
		return writer.guidance(err)
	}
	if err := writer.recover(); err != nil {
		return err
	}
	current, err := writer.inspectOutput("public")
	if err != nil {
		return err
	}
	if current != "absent" && current != "empty" && current != writer.state.Current {
		return fmt.Errorf("refuse public: ownership marker does not match private site record")
	}
	if current == "absent" || current == "empty" {
		writer.state.Current = ""
	}
	token := make([]byte, tokenBytes)
	if _, err := rand.Read(token); err != nil {
		return fmt.Errorf("generate ownership token: %w", err)
	}
	pending := &transaction{
		Token: hex.EncodeToString(token), Previous: writer.state.Current,
		HadPublic: current != "absent", Phase: "staging",
	}
	writer.state.Pending = pending
	if err := writer.saveState(); err != nil {
		return writer.guidance(err)
	}
	if err := writer.stage(ctx, files); err != nil {
		return errors.Join(fmt.Errorf("stage output; previous public unchanged: %w", err), writer.recover())
	}
	if err := ctx.Err(); err != nil {
		return errors.Join(err, writer.recover())
	}
	pending.Phase = "ready"
	if err := writer.saveState(); err != nil {
		return writer.guidance(err)
	}
	if pending.HadPublic {
		if err := writer.rename("public", pending.backup()); err != nil {
			return errors.Join(fmt.Errorf("preserve previous output: %w", err), writer.recover())
		}
	}
	if err := writer.rename(pending.stage(), "public"); err != nil {
		return errors.Join(fmt.Errorf("publish staged output: %w", err), writer.recover())
	}
	return writer.recover()
}

func (writer *Writer) stage(ctx context.Context, files []File) (stageErr error) {
	pending := writer.state.Pending
	if err := writer.root.Mkdir(pending.stage(), privateMode); err != nil {
		return fmt.Errorf("create stage: %w", err)
	}
	target, err := writer.root.OpenRoot(pending.stage())
	if err != nil {
		return fmt.Errorf("open stage: %w", err)
	}
	defer func() { stageErr = errors.Join(stageErr, target.Close()) }()
	data, err := json.Marshal(ownership{Version: stateVersion, Token: pending.Token})
	if err != nil {
		return err
	}
	if err := writeFile(target, File{Path: markerPath, Source: "Cress ownership marker", Data: data}); err != nil {
		return err
	}
	for _, file := range files {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := writeFile(target, file); err != nil {
			return err
		}
	}
	// The stage is private; the published directory must remain readable by hosting tools.
	if err := target.Chmod(".", directoryMode); err != nil {
		return fmt.Errorf("set output permissions: %w", err)
	}
	return nil
}

func (writer *Writer) recover() error {
	pending := writer.state.Pending
	if pending == nil {
		return nil
	}
	public, err := writer.inspectOutput("public")
	if err != nil {
		return writer.guidance(err)
	}
	backup, err := writer.inspectOutput(pending.backup())
	if err != nil {
		return writer.guidance(err)
	}
	stage, err := writer.inspectOutput(pending.stage())
	if err != nil {
		return writer.guidance(err)
	}
	previous := pending.Previous
	if previous == "" {
		previous = "empty"
	}
	if stage != "absent" && stage != pending.Token && (pending.Phase != "staging" || stage != "empty") {
		return writer.guidance(fmt.Errorf("stage ownership mismatch"))
	}
	if backup != "absent" && (!pending.HadPublic || backup != previous) {
		return writer.guidance(fmt.Errorf("backup ownership mismatch"))
	}
	if public == pending.Token {
		if stage != "absent" || pending.Phase == "staging" ||
			(pending.Phase == "ready" && pending.HadPublic && backup == "absent") {
			return writer.guidance(fmt.Errorf("ambiguous published transaction"))
		}
		if pending.Phase != "published" {
			writer.state.Current = pending.Token
			pending.Phase = "published"
			if err := writer.saveState(); err != nil {
				return writer.guidance(fmt.Errorf("new output is published; record publication state: %w", err))
			}
		}
		if backup != "absent" {
			if err := writer.removeAll(pending.backup()); err != nil {
				return writer.guidance(fmt.Errorf("new output is published; backup cleanup failed: %w", err))
			}
		}
	} else {
		if pending.Phase == "published" {
			return writer.guidance(fmt.Errorf("confirmed public output cannot be verified"))
		}
		switch {
		case public == "absent" && pending.HadPublic && backup == previous && pending.Phase == "ready":
			if err := writer.rename(pending.backup(), "public"); err != nil {
				return writer.guidance(fmt.Errorf("restore previous output failed: %w", err))
			}
		case public == previous && pending.HadPublic && backup == "absent":
		case public == "absent" && !pending.HadPublic && backup == "absent":
		default:
			return writer.guidance(fmt.Errorf("ambiguous previous-output transaction"))
		}
		if stage != "absent" {
			if err := writer.removeAll(pending.stage()); err != nil {
				return writer.guidance(fmt.Errorf("previous output restored or unchanged; stage cleanup failed: %w", err))
			}
		}
	}
	writer.state.Pending = nil
	if err := writer.saveState(); err != nil {
		writer.state.Pending = pending
		return writer.guidance(err)
	}
	return nil
}
