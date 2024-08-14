package storage

type MongoStorage interface {
	Save(data interface{}) error
}
