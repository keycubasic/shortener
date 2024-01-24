package loader

import (
	"bufio"
	"os"

	"github.com/keycubasic/shortener/internal/dao"
)

type database interface {
	GetLink(key string) (string, bool)
	SetLink(key string, link string)
}

type Producer struct {
	file   *os.File
	writer *bufio.Writer
}

func NewProducer(fname string) (*Producer, error) {
	file, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		return nil, err
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

	data, err := linkDao.Marshal()

	if err != nil {
		return err
	}

	if _, err := p.writer.Write(data); err != nil {
		return err
	}

	if err := p.writer.WriteByte('\n'); err != nil {
		return err
	}

	return p.writer.Flush()
}

func (p *Producer) Close() error {
	return p.file.Close()
}

type Consumer struct {
	file    *os.File
	scanner *bufio.Scanner
}

func NewConsumer(fname string) (*Consumer, error) {
	file, err := os.OpenFile(fname, os.O_RDONLY|os.O_CREATE, 0666)

	if err != nil {
		return nil, err
	}

	return &Consumer{
		file:    file,
		scanner: bufio.NewScanner(file),
	}, nil
}

func (c *Consumer) Load(db database) error {
	for c.scanner.Scan() {
		linkDao := dao.NewLinkDao()
		err := linkDao.Unmarshal([]byte(c.scanner.Text()))

		if err != nil {
			return err
		}

		db.SetLink(linkDao.ShortURL, linkDao.OriginalURL)
	}
	return nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}
