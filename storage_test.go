package stream

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	errMockInit = errors.New("errMockInit")
	errMockSet  = errors.New("errMockSet")
)

type mockRemoteDB struct {
	getChecksF func(context.Context) ([]string, error)
	setChecksF func(context.Context, []string) error
}

func (m mockRemoteDB) GetChecks(ctx context.Context) ([]string, error) {
	return m.getChecksF(ctx)
}
func (m mockRemoteDB) SetChecks(ctx context.Context, checks []string) error {
	return m.setChecksF(ctx, checks)
}

func TestGetAbortedChecks(t *testing.T) {
	testCases := []struct {
		name        string
		db          mockRemoteDB
		wantInitErr error
		wantChecks  []string
		wantErr     error
	}{
		{
			name: "Happy path",
			db: mockRemoteDB{
				getChecksF: func(context.Context) ([]string, error) {
					return []string{"check1", "check2"}, nil
				},
			},
			wantChecks: []string{"check1", "check2"},
		},
		{
			name: "Init error",
			db: mockRemoteDB{
				getChecksF: func(context.Context) ([]string, error) {
					return []string{}, errMockInit
				},
			},
			wantInitErr: errMockInit,
			wantChecks:  []string{},
		},
	}

	ctx := context.Background()
	log := log.New()

	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			storage, err := NewStorage(tc.db, log)
			if !errors.Is(err, tc.wantInitErr) {
				t.Fatalf("expected init err to be: %v\nbut got: %v", tc.wantInitErr, err)
			}
			if err != nil {
				return
			}
			checks, err := storage.GetAbortedChecks(ctx)
			if !reflect.DeepEqual(checks, tc.wantChecks) {
				t.Fatalf("expected checks to be:\n%v\nbut got:\n%v", tc.wantChecks, checks)
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected err to be: %v\nbut got: %v", tc.wantErr, err)
			}
		})
	}
}

func TestAddAbortedChecks(t *testing.T) {
	testCases := []struct {
		name        string
		db          mockRemoteDB
		abortChecks []string
		wantCache   cache
		wantErr     error
	}{
		{
			name: "Happy path",
			db: mockRemoteDB{
				setChecksF: func(ctx context.Context, checks []string) error {
					return nil
				},
			},
			abortChecks: []string{"checkID1", "checkID2"},
			wantCache:   []string{"checkID1", "checkID2"},
		},
		{
			name: "Error on set",
			db: mockRemoteDB{
				setChecksF: func(ctx context.Context, checks []string) error {
					return errMockSet
				},
			},
			abortChecks: []string{"checkID1", "checkID2"},
			wantCache:   []string{},
			wantErr:     errMockSet,
		},
	}

	ctx := context.Background()
	log := log.New()

	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			storage := storage{
				db:    tc.db,
				cache: cache{},
				log:   log,
			}
			err := storage.AddAbortedChecks(ctx, tc.abortChecks)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected err to be: %v\nbut got: %v", tc.wantErr, err)
			}
			if !reflect.DeepEqual(tc.wantCache, storage.cache) {
				t.Fatalf("expected local cache to be:\n%v\nbut got:\n%v", tc.wantCache, storage.cache)
			}
		})
	}
}

func TestConcurrentRW(t *testing.T) {
	ctx := context.Background()
	log := log.New()

	db := mockRemoteDB{
		setChecksF: func(ctx context.Context, checks []string) error {
			time.Sleep(1 * time.Second) // mock locking time
			return nil
		},
	}

	abortChecks := []string{"checkID3"}
	initialCache := []string{"checkID1", "checkID2"}
	expectedCache := append(initialCache, abortChecks...)

	storage := storage{
		db:    db,
		cache: initialCache,
		log:   log,
	}

	go storage.AddAbortedChecks(ctx, abortChecks)
	time.Sleep(50 * time.Millisecond) // wait for lock to be acquired

	checks, err := storage.GetAbortedChecks(ctx)
	if err != nil {
		t.Fatalf("expected no error but got: %v", err)
	}
	if !reflect.DeepEqual(checks, expectedCache) {
		t.Fatalf("expected checks to be:\n%v\nbut got:\n%v", expectedCache, checks)
	}

}
