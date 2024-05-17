//go:build integration

package postgre

import (
	"context"
	"log"
	"testing"

	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/keycubasic/shortener/internal/config"
	"github.com/keycubasic/shortener/internal/dao"
	"github.com/keycubasic/shortener/internal/storage/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type LinksTestSuite struct {
	suite.Suite
	pgContainer *testhelpers.PostgresContainer
	db          *Links
	ctx         context.Context
}

func (suite *LinksTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	pgContainer, err := testhelpers.CreatePostgresContainer(suite.ctx)
	if err != nil {
		log.Fatal(err)
	}
	suite.pgContainer = pgContainer
	cfg := new(config.Cfg)
	cfg.DatabaseDSN = pgContainer.ConnectionString
	db, err := NewLinks(suite.ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Migrate(suite.ctx)
	if err != nil {
		log.Fatal(err)
	}

	suite.db = db
}

func (suite *LinksTestSuite) TearDownSuite() {
	if err := suite.pgContainer.Terminate(suite.ctx); err != nil {
		log.Fatalf("error terminating postgres container: %s", err)
	}
}

func (suite *LinksTestSuite) Test_1_PingContext() {
	t := suite.T()
	err := suite.db.PingContext(suite.ctx)
	assert.NoError(t, err)
}

func (suite *LinksTestSuite) Test_2_AddUser() {
	t := suite.T()
	userID, err := suite.db.AddUser(suite.ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, userID)
}

func (suite *LinksTestSuite) Test_31_SetLinkOK() {
	t := suite.T()
	err := suite.db.SetLink(suite.ctx, 1, "6YGS4ZUF", "https://ya.ru")
	assert.NoError(t, err)
}

func (suite *LinksTestSuite) Test_32_SetLinkKO() {
	t := suite.T()
	err := suite.db.SetLink(suite.ctx, 1, "6YGS4ZUF", "https://ya.ru")
	require.Error(t, err)
}

func (suite *LinksTestSuite) Test_41_GetLinkOK() {
	t := suite.T()
	link, err := suite.db.GetLink(suite.ctx, "6YGS4ZUF")
	assert.NoError(t, err)
	assert.Equal(t, "https://ya.ru", link)
}

func (suite *LinksTestSuite) Test_42_GetLinkKO() {
	t := suite.T()
	_, err := suite.db.GetLink(suite.ctx, "6YGS4ZUF1")
	require.Error(t, err)
}

func (suite *LinksTestSuite) Test_51_GetShortLinkOK() {
	t := suite.T()
	link, err := suite.db.GetShortLink(suite.ctx, "https://ya.ru")
	assert.NoError(t, err)
	assert.Equal(t, "6YGS4ZUF", link)
}

func (suite *LinksTestSuite) Test_52_GetShortLinkKO() {
	t := suite.T()
	_, err := suite.db.GetShortLink(suite.ctx, "https://ya1.ru")
	require.Error(t, err)
}

func (suite *LinksTestSuite) Test_61_CreateShortLinksBatchOK() {
	t := suite.T()
	reqLinksDao := make(dao.LinksDao, 0, 1)
	reqLinksDao = append(reqLinksDao, &dao.LinkDao{CorrelationID: "1", OriginalURL: "https://google.com"})
	resLinksDao, err := suite.db.CreateShortLinksBatch(suite.ctx, 1, reqLinksDao)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(resLinksDao))
}

func (suite *LinksTestSuite) Test_62_CreateShortLinksBatchKO() {
	t := suite.T()
	reqLinksDao := make(dao.LinksDao, 0, 1)
	reqLinksDao = append(reqLinksDao, &dao.LinkDao{CorrelationID: "1", OriginalURL: "https://google.com"})
	_, err := suite.db.CreateShortLinksBatch(suite.ctx, 1, reqLinksDao)
	require.Error(t, err)
}

func (suite *LinksTestSuite) Test_7_GetUserURLs() {
	t := suite.T()
	resLinksDaoBatch, err := suite.db.GetUserURLs(suite.ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(resLinksDaoBatch))
}

func (suite *LinksTestSuite) Test_8_DeleteShortLinksBatch() {
	t := suite.T()
	reqLinksDao := make(dao.LinksDao, 0, 1)
	reqLinksDao = append(reqLinksDao, &dao.LinkDao{CorrelationID: "1", OriginalURL: "https://google.com"})
	err := suite.db.DeleteShortLinksBatch(suite.ctx, reqLinksDao)
	assert.NoError(t, err)
}

func (suite *LinksTestSuite) Test_9_Close() {
	t := suite.T()
	err := suite.db.Close()
	assert.NoError(t, err)
}

func TestLinksTestSuite(t *testing.T) {
	suite.Run(t, new(LinksTestSuite))
}
