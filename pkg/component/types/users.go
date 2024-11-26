package types

import "go.mongodb.org/mongo-driver/bson/primitive"

const (
	ListRadOnline       = "list:rad_online"
	ListRadOnlineUpdate = "list:rad_online:update"
	HashRadOnline       = "hash:rad_online:%s"
	ListAntiProxy       = "list:antiproxy:%s"
)

type User struct {
	RadOnlineID int    `redis:"rad_online_id" json:"rad_online_id"`
	UserName    string `redis:"user_name" json:"user_name"`
	IP          string `redis:"ip" json:"ip"`
	UserMac     string `redis:"user_mac" json:"user_mac"`
	LineType    int    `redis:"line_type" json:"line_type"`
	AddTime     int    `redis:"add_time" json:"add_time"`
	ProductsID  int    `redis:"products_id" json:"products_id"`
	BillingID   int    `redis:"billing_id" json:"billing_id"`
	ContractID  int    `redis:"contract_id" json:"contract_id"`
}

type UserEvent struct {
	ID              primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Action          int                `json:"action" bson:"action,omitempty"`
	SessionId       string             `json:"session_id" bson:"session_id,omitempty"`
	NasIp           string             `json:"nas_ip" bson:"nas_ip,omitempty"`
	NasIp1          string             `json:"nas_ip1" bson:"nas_ip1,omitempty"`
	UserName        string             `json:"user_name" bson:"user_name,omitempty"`
	Ip              string             `json:"ip" bson:"ip,omitempty"`
	Ip6             string             `json:"ip6" bson:"ip6,omitempty"`
	Ip61            string             `json:"ip6_1" bson:"ip6_1,omitempty"`
	Ip62            string             `json:"ip6_2" bson:"ip6_2,omitempty"`
	Ip63            string             `json:"ip6_3" bson:"ip6_3,omitempty"`
	UserMac         string             `json:"user_mac" bson:"user_mac,omitempty"`
	NasPort         int                `json:"nas_port" bson:"nas_port,omitempty"`
	NasPortId       string             `json:"nas_port_id" bson:"nas_port_id,omitempty"`
	CalledStationId string             `json:"called_station_id" bson:"called_station_id,omitempty"`
	NasIdentifier   string             `json:"nas_identifier" bson:"nas_identifier,omitempty"`
	NasPortType     int                `json:"nas_port_type" bson:"nas_port_type,omitempty"`
	VlanId          string             `json:"vlan_id" bson:"vlan_id,omitempty"`
	VlanId1         string             `json:"vlan_id1" bson:"vlan_id1,omitempty"`
	VlanId2         string             `json:"vlan_id2" bson:"vlan_id2,omitempty"`
	DeviceId        string             `json:"device_id" bson:"device_id,omitempty"`
	BandwidthUp     int                `json:"bandwidth_up" bson:"bandwidth_up,omitempty"`
	BandwidthDown   int                `json:"bandwidth_down" bson:"bandwidth_down,omitempty"`
	ProductsId      int                `json:"products_id" bson:"products_id,omitempty"`
	BillingId       int                `json:"billing_id" bson:"billing_id,omitempty"`
	ControlId       int                `json:"control_id" bson:"control_id,omitempty"`
	GroupId         int                `json:"group_id" bson:"group_id,omitempty"`
	RadOnlineId     int                `json:"rad_online_id" bson:"rad_online_id,omitempty"`
	DisableProxy    int                `json:"disable_proxy" bson:"disable_proxy,omitempty"`
	Domain          string             `json:"domain" bson:"domain,omitempty"`
	OsName          string             `json:"os_name" bson:"os_name,omitempty"`
	ClassName       string             `json:"class_name" bson:"class_name,omitempty"`
	MobilePhone     string             `json:"mobile_phone" bson:"mobile_phone,omitempty"`
	MobilePassword  string             `json:"mobile_password" bson:"mobile_password,omitempty"`
	IsArrears       int                `json:"is_arrears" bson:"is_arrears,omitempty"`
	BytesIn         int                `json:"bytes_in" bson:"bytes_in,omitempty"`
	BytesOut        int                `json:"bytes_out" bson:"bytes_out,omitempty"`
	AddTime         int                `json:"add_time" bson:"add_time,omitempty"`
	MyIp            string             `json:"my_ip" bson:"my_ip,omitempty"`
	DropCause       int                `json:"drop_cause" bson:"drop_cause,omitempty"`
	UserDebug       int                `json:"user_debug" bson:"user_debug,omitempty"`
	LineType        int                `json:"line_type" bson:"line_type,omitempty"`
	AcType          string             `json:"ac_type" bson:"ac_type,omitempty"`
	Daa             int                `json:"daa" bson:"daa,omitempty"`
	PoolId          int                `json:"pool_id" bson:"pool_id,omitempty"`
	Drop            int                `json:"drop" bson:"drop,omitempty"`
	CurBytesIn      int                `json:"cur_bytes_in" bson:"cur_bytes_in,omitempty"`
	CurBytesOut     int                `json:"cur_bytes_out" bson:"cur_bytes_out,omitempty"`
	CurBytesIn6     int                `json:"cur_bytes_in6" bson:"cur_bytes_in6,omitempty"`
	CurBytesOut6    int                `json:"cur_bytes_out6" bson:"cur_bytes_out6,omitempty"`
	CheckoutDate    int                `json:"checkout_date" bson:"checkout_date,omitempty"`
	RemainDay       int                `json:"remain_day" bson:"remain_day,omitempty"`
	RemainBytes     int                `json:"remain_bytes" bson:"remain_bytes,omitempty"`
	SumTimes        int                `json:"sum_times" bson:"sum_times,omitempty"`
	SumBytes        int                `json:"sum_bytes" bson:"sum_bytes,omitempty"`
	SumSeconds      int                `json:"sum_seconds" bson:"sum_seconds,omitempty"`
	AllBytes        int                `json:"all_bytes" bson:"all_bytes,omitempty"`
	AllSeconds      int                `json:"all_seconds" bson:"all_seconds,omitempty"`
	UserBalance     float64            `json:"user_balance" bson:"user_balance,omitempty"`
	UserCharge      float64            `json:"user_charge" bson:"user_charge,omitempty"`
	CurCharge       float64            `json:"cur_charge" bson:"cur_charge,omitempty"`
	DropReason      int                `json:"drop_reason" bson:"drop_reason,omitempty"`
	DropTime        int                `json:"drop_time" bson:"drop_time,omitempty"`
	DestControl     int                `json:"dest_control" bson:"dest_control,omitempty"`
	Proc            string             `json:"proc" bson:"proc,omitempty"`
}

type Events interface {
	LoadEvent()
	DropEvent()
	Save2Mongo()
}
