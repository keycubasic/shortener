package loader

import (
	"bufio"
	"fmt"
	"os"

	"github.com/keycubasic/shortener/internal/dao"
	json "github.com/mailru/easyjson"
)

type database interface {
	GetLink(key string) (string, bool)
	SetLink(key string, link string) error
}

type Producer struct {
	file   *os.File
	writer *bufio.Writer
}

func NewProducer(fname string) (*Producer, error) {
	file, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		return nil, fmt.Errorf("producer failed to open file: %w", err)
	}

	return &Producer{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

func (p *Producer) Save(id int, key, link string) error {
	linkDao := dao.NewLinkDao()
	linkDao.UUID = id
	linkDao.ShortURL = key
	linkDao.OriginalURL = link

	data, err := json.Marshal(linkDao)

	if err != nil {
		return fmt.Errorf("producer failed to marshal dao: %w", err)
	}

	if _, err := p.writer.Write(data); err != nil {
		return fmt.Errorf("producer failed to write data: %w", err)
	}

	if err := p.writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("producer failed to write a line break: %w", err)
	}

	return p.writer.Flush()
}

func (p *Producer) Close() error {
	return p.file.Close()
}

type Consumer struct {
	db      database
	file    *os.File
	scanner *bufio.Scanner
}

func NewConsumer(db database, fname string) (*Consumer, error) {
	file, err := os.OpenFile(fname, os.O_RDONLY|os.O_CREATE, 0666)

	if err != nil {
		return nil, fmt.Errorf("сonsumer failed to open file: %w", err)
	}

	return &Consumer{
		db:      db,
		file:    file,
		scanner: bufio.NewScanner(file),
	}, nil
}

func (c *Consumer) Load() error {
	for c.scanner.Scan() {
		linkDao := dao.NewLinkDao()
		err := json.Unmarshal([]byte(c.scanner.Text()), linkDao)

		if err != nil {
			return fmt.Errorf("сonsumer failed to unmarshal dao: %w", err)
		}

		if err := c.db.SetLink(linkDao.ShortURL, linkDao.OriginalURL); err != nil {
			return fmt.Errorf("сonsumer failed to load data into storage: %w", err)
		}
	}
	return nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}
