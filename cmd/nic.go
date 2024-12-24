package cmd

import (
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"net"
	"os"
)

var NicCmd = &cobra.Command{
	Use:   "nic",
	Short: "Show NICs",
	Run: func(cmd *cobra.Command, args []string) {
		// 创建表格
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Interface Name", "IP Address", "Subnet Mask", "Network Address", "Broadcast Address", "Valid Host Range"})

		// 获取所有网卡接口
		interfaces, err := net.Interfaces()
		if err != nil {
			fmt.Println("Error fetching interfaces:", err)
			return
		}

		for _, iface := range interfaces {
			// 忽略没有启用或没有地址的接口
			if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
				continue
			}

			// 获取该接口的所有地址
			addrs, err := iface.Addrs()
			if err != nil {
				fmt.Printf("Error fetching addresses for %s: %v\n", iface.Name, err)
				continue
			}

			// 只记录第一条有效地址
			//var recorded bool
			for _, addr := range addrs {
				// 检查地址类型（IPv4 或 IPv6）
				ipNet, ok := addr.(*net.IPNet)
				if !ok {
					continue
				}

				// 跳过链路本地地址（例如 fe80::）
				//if ipNet.IP.IsLinkLocalUnicast() {
				//	continue
				//}

				// 获取子网信息
				ip, ipNet, err := utils.GetSubnetInfo(addr.String())
				if err != nil {
					fmt.Println("Error:", err)
					return
				}
				broadcast := utils.GetBroadcast(ipNet)
				firstHost, lastHost := utils.GetIPRange(ipNet, broadcast)
				table.Append([]string{
					iface.Name,
					ip.String(),
					ipNet.Mask.String(),
					ipNet.String(),
					broadcast.String(),
					fmt.Sprintf("%s - %s", firstHost, lastHost),
				})
				//recorded = true
				continue // 只记录第一条
			}
		}
		// 渲染表格
		table.Render()
	},
}
