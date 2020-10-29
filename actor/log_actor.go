package actor

//const (
//	logCounterKey = "__counter"
//)
//
//func NewLogActor(logPath string) (Actor, error) {
//	db, err := leveldb.OpenFile(logPath, &opt.Options{
//		Filter: filter.NewBloomFilter(10),
//		//OpenFilesCacheCapacity: handles,
//		//BlockCacheCapacity:     cache / 2 * opt.MiB,
//		//WriteBuffer:            cache / 4 * opt.MiB, // Two of these are used internally
//		DisableSeeksCompaction: true,
//	})
//
//	if _, corrupted := err.(*errors.ErrCorrupted); corrupted {
//		db, err = leveldb.RecoverFile(logPath, nil)
//	}
//	if err != nil {
//		return nil, err
//	}
//
//	return NewActor(func(ctx context.Context, in interface{}) (out interface{}, err error) {
//
//	}), nil
//}
