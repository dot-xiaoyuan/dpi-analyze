package types

type Products struct {
	ProductsID   string `json:"products_id" bson:"_id" redis:"products_id"`
	ProductsName string `json:"products_name" bson:"products_name" redis:"products_name"`
	MgrName      string `json:"mgr_name" bson:"mgr_name" redis:"mgr_name"`
	Controls
	Policy
}

type Controls struct {
	ControlName      string `json:"control_name" bson:"control_name" redis:"control_name"`
	DisableProxy     int    `json:"disable_proxy" bson:"disable_proxy" redis:"disable_proxy"`
	ProxyTimes       int    `json:"proxy_times" bson:"proxy_times" redis:"proxy_times"`
	ProxyDisableTime int    `json:"proxy_disable_time" bson:"proxy_disable_time" redis:"proxy_disable_time"`
}

type Policy struct {
	ALL    int `json:"all" bson:"all"`
	Mobile int `json:"mobile" bson:"mobile"`
	Pc     int `json:"pc" bson:"pc"`
}
