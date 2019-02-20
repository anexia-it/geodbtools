package mmdatformat

import (
	"testing"

	"github.com/anexia-it/geodbtools"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRegisterType(t *testing.T) {
	t.Run("AlreadyRegistered", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dbType := NewMockType(ctrl)

		typeRegistryMu.Lock()
		originalTypeRegistry := typeRegistry
		typeRegistry = map[DatabaseTypeID]Type{
			DatabaseTypeIDBase: dbType,
		}
		typeRegistryMu.Unlock()

		defer func() {
			typeRegistryMu.Lock()
			defer typeRegistryMu.Unlock()
			typeRegistry = originalTypeRegistry
		}()

		assert.EqualError(t, RegisterType(DatabaseTypeIDBase, dbType), ErrTypeRegistered.Error())
		assert.Len(t, typeRegistry, 1)
	})

	t.Run("OK", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dbType := NewMockType(ctrl)

		typeRegistryMu.Lock()
		originalTypeRegistry := typeRegistry
		typeRegistry = map[DatabaseTypeID]Type{}
		typeRegistryMu.Unlock()

		defer func() {
			typeRegistryMu.Lock()
			defer typeRegistryMu.Unlock()
			typeRegistry = originalTypeRegistry
		}()

		assert.NoError(t, RegisterType(DatabaseTypeIDBase, dbType))
		assert.Len(t, typeRegistry, 1)
		assert.EqualValues(t, dbType, typeRegistry[DatabaseTypeIDBase])
	})
}

func TestMustRegisterType(t *testing.T) {
	t.Run("AlreadyRegistered", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dbType := NewMockType(ctrl)

		typeRegistryMu.Lock()
		originalTypeRegistry := typeRegistry
		typeRegistry = map[DatabaseTypeID]Type{
			DatabaseTypeIDBase: dbType,
		}
		typeRegistryMu.Unlock()

		defer func() {
			typeRegistryMu.Lock()
			defer typeRegistryMu.Unlock()
			typeRegistry = originalTypeRegistry
		}()

		assert.PanicsWithValue(t, ErrTypeRegistered, func() {
			MustRegisterType(DatabaseTypeIDBase, dbType)
		})
		assert.Len(t, typeRegistry, 1)
	})

	t.Run("OK", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dbType := NewMockType(ctrl)

		typeRegistryMu.Lock()
		originalTypeRegistry := typeRegistry
		typeRegistry = map[DatabaseTypeID]Type{}
		typeRegistryMu.Unlock()

		defer func() {
			typeRegistryMu.Lock()
			defer typeRegistryMu.Unlock()
			typeRegistry = originalTypeRegistry
		}()

		assert.NotPanics(t, func() {
			MustRegisterType(DatabaseTypeIDBase, dbType)
		})

		assert.Len(t, typeRegistry, 1)
		assert.EqualValues(t, dbType, typeRegistry[DatabaseTypeIDBase])
	})
}

func TestLookupType(t *testing.T) {
	t.Run("NotFound", func(t *testing.T) {
		typeRegistryMu.Lock()
		originalTypeRegistry := typeRegistry
		typeRegistry = map[DatabaseTypeID]Type{}
		typeRegistryMu.Unlock()

		defer func() {
			typeRegistryMu.Lock()
			defer typeRegistryMu.Unlock()
			typeRegistry = originalTypeRegistry
		}()

		dbType, typeID, err := LookupType(geodbtools.DatabaseTypeCountry)
		assert.Nil(t, dbType)
		assert.EqualValues(t, 0, typeID)
		assert.EqualError(t, err, ErrTypeNotFound.Error())
	})

	t.Run("OK", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedDBType := NewMockType(ctrl)
		expectedDBType.EXPECT().DatabaseType().Return(geodbtools.DatabaseType("test type"))

		typeRegistryMu.Lock()
		originalTypeRegistry := typeRegistry
		typeRegistry = map[DatabaseTypeID]Type{
			DatabaseTypeIDBase: expectedDBType,
		}
		typeRegistryMu.Unlock()

		defer func() {
			typeRegistryMu.Lock()
			defer typeRegistryMu.Unlock()
			typeRegistry = originalTypeRegistry
		}()

		dbType, typeID, err := LookupType(geodbtools.DatabaseType("test type"))
		assert.NoError(t, err)
		assert.EqualValues(t, expectedDBType, dbType)

		assert.EqualValues(t, DatabaseTypeIDBase, typeID)
	})
}

func TestLookupTypeByDatabaseType(t *testing.T) {
	t.Run("NotFound", func(t *testing.T) {
		typeRegistryMu.Lock()
		originalTypeRegistry := typeRegistry
		typeRegistry = map[DatabaseTypeID]Type{}
		typeRegistryMu.Unlock()

		defer func() {
			typeRegistryMu.Lock()
			defer typeRegistryMu.Unlock()
			typeRegistry = originalTypeRegistry
		}()

		dbType, err := LookupTypeByDatabaseType(DatabaseTypeIDBase)
		assert.Nil(t, dbType)
		assert.EqualError(t, err, ErrTypeNotFound.Error())
	})

	t.Run("OK", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedDBType := NewMockType(ctrl)

		typeRegistryMu.Lock()
		originalTypeRegistry := typeRegistry
		typeRegistry = map[DatabaseTypeID]Type{
			DatabaseTypeIDBase: expectedDBType,
		}
		typeRegistryMu.Unlock()

		defer func() {
			typeRegistryMu.Lock()
			defer typeRegistryMu.Unlock()
			typeRegistry = originalTypeRegistry
		}()

		dbType, err := LookupTypeByDatabaseType(DatabaseTypeIDBase)
		assert.NoError(t, err)
		assert.EqualValues(t, expectedDBType, dbType)
	})
}
