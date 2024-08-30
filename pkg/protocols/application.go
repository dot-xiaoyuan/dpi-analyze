package protocols

type ApplicationData struct {
	Protocol string      `bson:"protocols"`
	Data     interface{} `bson:"data"`
}
