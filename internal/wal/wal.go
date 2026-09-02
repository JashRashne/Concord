package wal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/jashrashne/concord/internal/command"
)

const CurrentVersion = 1

type Record struct {
	Version int             `json:"version"`
	Command command.Command `json:"command"`
}

type Log struct {
	mu   sync.Mutex
	path string
	file *os.File
}

func NewRecord(cmd command.Command) Record {
	return Record{
		Version: CurrentVersion,
		Command: cmd,
	}
}

func Open(path string) (*Log, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("create WAL directory: %w", err)
	}

	file, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_RDWR|os.O_APPEND,
		0644,
	)
	if err != nil {
		return nil, fmt.Errorf("open WAL: %w", err)
	}

	return &Log{
		path: path,
		file: file,
	}, nil
}

func (l *Log) Append(record Record) error {
	if record.Version != CurrentVersion {
		return fmt.Errorf(
			"unsupported WAL version %d",
			record.Version,
		)
	}

	if err := record.Command.Validate(); err != nil {
		return err
	}

	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("encode WAL record: %w", err)
	}

	data = append(data, '\n')

	l.mu.Lock()
	defer l.mu.Unlock()

	n, err := l.file.Write(data)
	if err != nil {
		return fmt.Errorf("write WAL record: %w", err)
	}

	if n != len(data) {
		return io.ErrShortWrite
	}

	if err := l.file.Sync(); err != nil {
		return fmt.Errorf("sync WAL: %w", err)
	}

	return nil
}

func (l *Log) ReadAll() ([]Record, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	file, err := os.Open(l.path)
	if err != nil {
		return nil, fmt.Errorf("open WAL for reading: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	var records []Record

	for {
		var record Record

		err := decoder.Decode(&record)

		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("decode WAL record: %w", err)
		}

		if record.Version != CurrentVersion {
			return nil, fmt.Errorf(
				"unsupported WAL version %d",
				record.Version,
			)
		}

		if err := record.Command.Validate(); err != nil {
			return nil, fmt.Errorf("invalid WAL record: %w", err)
		}

		records = append(records, record)
	}

	return records, nil
}

func (l *Log) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.file.Close()
}
