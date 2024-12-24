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

		// 遍历所有网卡接口
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

			// 跟踪网卡的子网信息
			var recorded bool

			// 遍历该接口的所有地址
			for _, addr := range addrs {
				// 检查地址类型（IPv4 或 IPv6）
				ipNet, ok := addr.(*net.IPNet)
				if !ok {
					continue
				}

				// 如果是 IPv6 地址，则跳过
				if ipNet.IP.To4() == nil && len(addrs) == 1 {
					continue
				}

				// 获取子网信息
				ip, ipNet, err := utils.GetSubnetInfo(addr.String())
				if err != nil {
					fmt.Println("Error:", err)
					continue
				}
				broadcast := utils.GetBroadcast(ipNet)
				firstHost, lastHost := utils.GetIPRange(ipNet, broadcast)

				// 如果是第一次记录该网卡信息，则记录网卡名称
				if !recorded {
					table.Append([]string{
						iface.Name, // 网卡名称
						ip.String(),
						ipNet.Mask.String(),
						ipNet.String(),
						broadcast.String(),
						fmt.Sprintf("%s - %s", firstHost, lastHost),
					})
					recorded = true
				} else {
					// 如果网卡已经有记录，只需要填写子网相关信息
					table.Append([]string{
						"", // 合并单元格（网卡名称列空白）
						ip.String(),
						ipNet.Mask.String(),
						ipNet.String(),
						broadcast.String(),
						fmt.Sprintf("%s - %s", firstHost, lastHost),
					})
				}
			}

		}

		// 渲染表格
		table.Render()
	},
}
