// Package loader contains functions for working with file storage.
package loader

import (
	"bufio"
	"context"
	"fmt"
	"os"

	jsoniter "github.com/json-iterator/go"

	"github.com/keycubasic/shortener/internal/dao"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Interface of database
type database interface {
	// GetLink returns the original shortened link from the database.
	GetLink(ctx context.Context, key string) (string, error)
	// SetLink sets a new link-shortening pair to the database.
	SetLink(ctx context.Context, userID int, key, link string) error
}

// Producer stores the file and writer.
type Producer struct {
	file   *os.File
	writer *bufio.Writer
}

func NewProducer(fname string) (*Producer, error) {
	file, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		return nil, fmt.Errorf("loader producer open file: %w", err)
	}

	return &Producer{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

// Save saves the data to a file.
func (p *Producer) Save(_ context.Context, id int, key, link string) error {
	linkDao := dao.NewLinkDao()
	linkDao.UUID = id
	linkDao.ShortURL = key
	linkDao.OriginalURL = link

	data, err := json.Marshal(linkDao)
	if err != nil {
		return fmt.Errorf("loader producer marshal dao: %w", err)
	}

	if _, err := p.writer.Write(data); err != nil {
		return fmt.Errorf("loader producer write data: %w", err)
	}

	if err := p.writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("loader producer write a line break: %w", err)
	}

	return p.writer.Flush()
}

// Close closes the file.
func (p *Producer) Close() error {
	return p.file.Close()
}

// Consumer stores the database interface, file and data scanner.
type Consumer struct {
	db      database
	file    *os.File
	scanner *bufio.Scanner
}

func NewConsumer(db database, fname string) (*Consumer, error) {
	file, err := os.OpenFile(fname, os.O_RDONLY|os.O_CREATE, 0o666)
	if err != nil {
		return nil, fmt.Errorf("loader сonsumer open file: %w", err)
	}

	return &Consumer{
		db:      db,
		file:    file,
		scanner: bufio.NewScanner(file),
	}, nil
}

// Load loads data from a file into the database.
func (c *Consumer) Load(ctx context.Context) error {
	for c.scanner.Scan() {
		linkDao := dao.NewLinkDao()
		err := json.Unmarshal([]byte(c.scanner.Text()), linkDao)
		if err != nil {
			return fmt.Errorf("loader сonsumer unmarshal dao: %w", err)
		}

		if err := c.db.SetLink(ctx, 0, linkDao.ShortURL, linkDao.OriginalURL); err != nil {
			return fmt.Errorf("loader сonsumer load data into storage: %w", err)
		}
	}
	return nil
}

// Close closes the file.
func (c *Consumer) Close() error {
	return c.file.Close()
}
