package protocol

type ApplicationData struct {
	Protocol string      `bson:"protocol"`
	Data     interface{} `bson:"data"`
}
