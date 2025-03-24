package defaults

import (
	"fmt"

	"github.com/pureapi/pureapi-framework/custom"
	"github.com/pureapi/pureapi-framework/repository"
	repositorytypes "github.com/pureapi/pureapi-framework/repository/types"
)

func DefaultConversionRules() map[string]func(any) any {
	return map[string]func(any) any{
		"uuid": func(a any) any {
			strVal, ok := a.(string)
			if !ok {
				return nil
			}
			uuidVal, err := custom.NewUUIDGen().FromString(strVal)
			if err != nil {
				return nil
			}
			return uuidVal
		},
	}
}

func DefaultCustomRules() map[string]func(any) error {
	return map[string]func(any) error{
		"uuid": func(a any) error {
			_, err := custom.NewUUIDGen().FromString(fmt.Sprintf("%v", a))
			if err != nil {
				return fmt.Errorf("invalid uuid")
			}
			return nil
		},
	}
}

// DefaultMutatorRepo returns a new DefaultMutatorRepo.
func DefaultMutatorRepo() repositorytypes.MutatorRepo {
	return repository.NewMutatorRepo(
		custom.QueryBuilder(), custom.QueryErrorChecker(),
	)
}

// DefaultReaderRepo returns a new DefaultReaderRepo.
func DefaultReaderRepo() repositorytypes.ReaderRepo {
	return repository.NewReaderRepo(
		custom.QueryBuilder(), custom.QueryErrorChecker(),
	)
}

// NewRawQueryer returns a new DefaultRawQueryer.
func NewRawQueryer() repositorytypes.RawQueryer {
	return repository.NewRawQueryer()
}

// DefaultTxManager returns a new DefaultTxManager.
func DefaultTxManager[T any]() repositorytypes.TxManager[T] {
	return repository.NewTxManager[T]()
}
