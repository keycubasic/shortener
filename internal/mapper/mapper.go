// Package mapper contains functions for mapping data objects.
package mapper

import (
	"strings"

	"github.com/keycubasic/shortener/internal/dao"
	"github.com/keycubasic/shortener/internal/dto"
)

// LinkDtoToLinkDao converts a data transfer object into a data access object.
func LinkDtoToLinkDao(linkDto dto.LinkDto) *dao.LinkDao {
	linkDao := new(dao.LinkDao)
	linkDao.CorrelationID = linkDto.CorrelationID
	linkDao.OriginalURL = linkDto.OriginalURL
	return linkDao
}

// LinkDaoToLinkDto converts a data access object into a data transfer object.
func LinkDaoToLinkDto(linkDao dao.LinkDao, baseURL string) *dto.LinkDto {
	linkDto := new(dto.LinkDto)
	linkDto.CorrelationID = linkDao.CorrelationID
	linkDto.ShortURL = strings.Join([]string{baseURL, linkDao.ShortURL}, "/")
	return linkDto
}

// LinkDaoToLinkDtoURLs converts a batch data access object into a data transfer object.
func LinkDaoToLinkDtoURLs(linkDao dao.LinkDaoBatch, baseURL string) *dto.LinkDto {
	linkDto := new(dto.LinkDto)
	linkDto.ShortURL = strings.Join([]string{baseURL, linkDao.ShortURL}, "/")
	linkDto.OriginalURL = linkDao.OriginalURL
	return linkDto
}

// LinkToLinkDao converts the user ID and link into a data access object.
func LinkToLinkDao(userID int, link string) *dao.LinkDao {
	linkDao := new(dao.LinkDao)
	linkDao.UserID = userID
	linkDao.ShortURL = link
	linkDao.DeletedFlag = true
	return linkDao
}
