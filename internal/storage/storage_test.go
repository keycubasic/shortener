package storage

import (
	"context"
	"log"
	"testing"

	"github.com/keycubasic/shortener/internal/config"
	"github.com/keycubasic/shortener/internal/dao"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type LinksTestSuite struct {
	suite.Suite
	db  *Links
	ctx context.Context
}

func (suite *LinksTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	cfg := new(config.Cfg)
	cfg.DatabaseDSN = ""
	db, err := NewLinks(suite.ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Migrate(suite.ctx)
	if err != nil {
		log.Fatal(err)
	}
	suite.db = db
	suite.db.SetLink(suite.ctx, 1, "xKh7DnOW", "https://mail.ru")
}

func (suite *LinksTestSuite) TearDownSuite() {
	if err := suite.db.Close(); err != nil {
		log.Fatalf("error close database (map): %s", err)
	}
}

func (suite *LinksTestSuite) Test_PingContext() {
	t := suite.T()
	err := suite.db.PingContext(suite.ctx)
	assert.NoError(t, err)
}

func (suite *LinksTestSuite) Test_AddUser() {
	t := suite.T()
	userID, err := suite.db.AddUser(suite.ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, userID)
}

func (suite *LinksTestSuite) Test_SetLinkOK() {
	t := suite.T()
	err := suite.db.SetLink(suite.ctx, 1, "mZmevP23", "https://google.com")
	assert.NoError(t, err)
}

func (suite *LinksTestSuite) Test_SetLinkKO() {
	t := suite.T()
	err := suite.db.SetLink(suite.ctx, 1, "xKh7DnOW", "https://mail.ru")
	require.Error(t, err)
}

func (suite *LinksTestSuite) Test_GetLinkOK() {
	t := suite.T()
	link, err := suite.db.GetLink(suite.ctx, "xKh7DnOW")
	assert.NoError(t, err)
	assert.Equal(t, "https://mail.ru", link)
}

func (suite *LinksTestSuite) Test_GetLinkKO() {
	t := suite.T()
	_, err := suite.db.GetLink(suite.ctx, "xKh7DnOW1")
	require.Error(t, err)
}

func (suite *LinksTestSuite) Test_GetShortLinkOK() {
	t := suite.T()
	link, err := suite.db.GetShortLink(suite.ctx, "https://mail.ru")
	assert.NoError(t, err)
	assert.Equal(t, "xKh7DnOW", link)
}

func (suite *LinksTestSuite) Test_GetShortLinkKO() {
	t := suite.T()
	_, err := suite.db.GetShortLink(suite.ctx, "https://ya1.ru")
	require.Error(t, err)
}

func (suite *LinksTestSuite) Test_CreateShortLinksBatchOK() {
	t := suite.T()
	reqLinksDao := make(dao.LinksDao, 0, 1)
	reqLinksDao = append(reqLinksDao, &dao.LinkDao{CorrelationID: "1", OriginalURL: "https://yahoo.com"})
	resLinksDao, err := suite.db.CreateShortLinksBatch(suite.ctx, 1, reqLinksDao)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(resLinksDao))
}

func (suite *LinksTestSuite) Test_CreateShortLinksBatchKO() {
	t := suite.T()
	reqLinksDao := make(dao.LinksDao, 0, 1)
	reqLinksDao = append(reqLinksDao, &dao.LinkDao{CorrelationID: "1", OriginalURL: "https://mail.ru"})
	_, err := suite.db.CreateShortLinksBatch(suite.ctx, 1, reqLinksDao)
	require.Error(t, err)
}

func (suite *LinksTestSuite) Test_GetUserURLs() {
	t := suite.T()
	_, err := suite.db.GetUserURLs(suite.ctx, 1)
	assert.NoError(t, err)
}

func (suite *LinksTestSuite) Test_DeleteShortLinksBatch() {
	t := suite.T()
	reqLinksDao := make(dao.LinksDao, 0, 1)
	reqLinksDao = append(reqLinksDao, &dao.LinkDao{CorrelationID: "1", OriginalURL: "https://google.com"})
	err := suite.db.DeleteShortLinksBatch(suite.ctx, reqLinksDao)
	assert.NoError(t, err)
}

func (suite *LinksTestSuite) Test_GetStats() {
	t := suite.T()
	_, err := suite.db.GetStats(suite.ctx)
	require.NoError(t, err)
}

func TestLinksTestSuite(t *testing.T) {
	suite.Run(t, new(LinksTestSuite))
}
