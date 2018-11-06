package api_test

import (
	"testing"

	"github.com/Toggly/core/internal/domain"

	"github.com/Toggly/core/internal/pkg/api"
	"github.com/Toggly/core/internal/pkg/storage"

	asserts "github.com/stretchr/testify/assert"
)

func TestProject(t *testing.T) {
	assert := asserts.New(t)

	BeforeTest()

	pApi := GetApi()

	pl, err := pApi.List()
	assert.Nil(err)
	assert.Empty(pl)

	pr, err := pApi.Get(ProjectCode)
	assert.Equal(api.ErrProjectNotFound, err)
	assert.Nil(pr)

	pr, err = pApi.Update(ProjectCode, "", domain.ProjectStatusActive)
	assert.Equal(api.ErrProjectNotFound, err)
	assert.Nil(pr)

	err = pApi.Delete(ProjectCode)
	assert.Equal(api.ErrProjectNotFound, err)

	pr, err = pApi.Create(ProjectCode, "Description 1", domain.ProjectStatusActive)
	assert.Nil(err)
	assert.NotNil(pr)
	assert.Equal(ProjectCode, pr.Code)
	assert.Equal("Description 1", pr.Description)
	assert.Equal(ow, pr.OwnerID)
	assert.NotNil(pr.RegDate)
	assert.Equal(domain.ProjectStatusActive, pr.Status)

	pl, err = pApi.List()
	assert.Len(pl, 1)

	_, err = pApi.Create(ProjectCode, "Description 1", domain.ProjectStatusActive)
	assert.NotNil(err)
	assert.IsType(&storage.UniqueIndexError{}, err)

	pr1, err := pApi.Get(ProjectCode)
	assert.Nil(err)
	assert.NotNil(pr1)
	assert.Equal(ProjectCode, pr1.Code)
	assert.Equal("Description 1", pr1.Description)
	assert.Equal(ow, pr1.OwnerID)
	assert.Equal(pr.RegDate, pr1.RegDate)
	assert.Equal(pr.Status, pr1.Status)

	pr1u, err := pApi.Update(ProjectCode, "Description 2", domain.ProjectStatusDisabled)
	assert.Nil(err)
	assert.NotNil(pr1u)
	assert.Equal(ProjectCode, pr1u.Code)
	assert.Equal("Description 2", pr1u.Description)
	assert.Equal(ow, pr1u.OwnerID)
	assert.Equal(pr.RegDate, pr1u.RegDate)
	assert.Equal(domain.ProjectStatusDisabled, pr1u.Status)

	pr2u, err := pApi.Update("p2", "Description 2", domain.ProjectStatusDisabled)
	assert.Nil(pr2u)
	assert.Equal(api.ErrProjectNotFound, err)

	pApi.For(ProjectCode).Environments().Create("env_code", "", false)

	assert.Equal(api.ErrProjectNotEmpty, pApi.Delete("p1"))

	assert.Nil(pApi.For(ProjectCode).Environments().Delete("env_code"))
	assert.Nil(pApi.Delete(ProjectCode))

	assert.Equal(api.ErrProjectNotFound, pApi.Delete(ProjectCode))

	AfterTest()
}
