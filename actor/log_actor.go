package actor

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"

	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

const (
	logCounterKey = "__counter"
	logPrefix     = "row-"
)

type LogActor struct {
	logPath       string
	disabledWrite bool
	db            *leveldb.DB
}

func (l *LogActor) OpenDB() error {
	var err error
	l.db, err = leveldb.OpenFile(l.logPath, &opt.Options{
		Filter: filter.NewBloomFilter(16),
		//OpenFilesCacheCapacity: 1000,
		BlockCacheCapacity:     100 * opt.MiB,
		WriteBuffer:            100 * opt.MiB, // Two of these are used internally
		DisableSeeksCompaction: true,
	})

	if _, corrupted := err.(*errors.ErrCorrupted); corrupted {
		l.db, err = leveldb.RecoverFile(l.logPath, nil)
	}
	if err != nil {
		return err
	}

	return nil
}

func (l *LogActor) SetDisabledWrite(disabledWrite bool) {
	l.disabledWrite = disabledWrite
}

// in: nothing out: interface{} from createStruct
func (l *LogActor) LogRestoreDaemon(createStruct func() interface{}) Daemon {
	return NewDaemon(func(ctx context.Context, in chan interface{}, out chan interface{}, errChan chan error) error {
		iter := l.db.NewIterator(util.BytesPrefix([]byte(logPrefix)), nil)
		defer iter.Release()
		for iter.Next() {
			logJson := iter.Value()
			row := createStruct()

			if err := json.Unmarshal(logJson, row); err != nil {
				select {
				case <-ctx.Done():
					return nil
				case errChan <- fmt.Errorf("error unmarshal data: %w", err):
					fmt.Println("ERR marshal:", err)
				}
			}

			select {
			case <-ctx.Done():
				return nil
			case out <- row:
			}
		}

		if err := iter.Error(); err != nil {
			return fmt.Errorf("iterator error: %w", err)
		}

		return nil
	})
}

func (l *LogActor) LogActor() (Actor, error) {
	lastRowIdBytes, err := l.db.Get([]byte(logCounterKey), nil)
	if err != nil && err != leveldb.ErrNotFound {
		return nil, err
	}

	var lastRowId uint64
	if len(lastRowIdBytes) > 0 {
		lastRowId, err = strconv.ParseUint(string(lastRowIdBytes), 10, 64)
		if err != nil {
			return nil, err
		}
	}

	return NewActor(func(ctx context.Context, in interface{}) (out interface{}, err error) {
		if l.disabledWrite {
			return in, nil
		}

		currentRowId := atomic.AddUint64(&lastRowId, 1)

		inJson, err := json.Marshal(in)
		if err != nil {
			return in, err
		}

		err = l.db.Put([]byte(logPrefix+strconv.FormatUint(currentRowId, 10)), inJson, nil)
		if err != nil {
			return in, err
		}

		err = l.db.Put([]byte(logCounterKey), []byte(strconv.FormatUint(currentRowId, 10)), nil)
		if err != nil {
			return in, err
		}

		return in, nil
	}), nil
}

func (l *LogActor) Close() error {
	return l.db.Close()
}

func NewLogActor(logPath string) *LogActor {
	return &LogActor{
		logPath: logPath,
	}
}
