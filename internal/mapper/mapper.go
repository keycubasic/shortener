package mapper

import (
	"github.com/keycubasic/shortener/internal/dao"
	"github.com/keycubasic/shortener/internal/dto"
)

func LinkDtoToLinkDao(linkDto dto.LinkDto) *dao.LinkDao {
	linkDao := dao.NewLinkDao()
	linkDao.CorrelationID = linkDto.CorrelationID
	linkDao.OriginalURL = linkDto.OriginalURL
	return linkDao
}

func LinkDaoToLinkDto(linkDao dao.LinkDao) *dto.LinkDto {
	linkDto := dto.NewLinkDto()
	linkDto.CorrelationID = linkDao.CorrelationID
	linkDto.ShortURL = linkDao.ShortURL
	return linkDto
}
