package cmd

import (
	"fmt"
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
				if ipNet.IP.IsLinkLocalUnicast() {
					continue
				}

				_, ipNet, err = net.ParseCIDR(addr.String())
				if err != nil || ipNet.IP == nil {
					continue
				}
				// 获取网络地址
				network := ipNet.IP

				// 获取子网掩码的位数
				mask := ipNet.Mask
				maskSize, _ := mask.Size()

				// 获取广播地址
				broadcast := make(net.IP, len(network))
				for i := range network {
					broadcast[i] = network[i] | ^mask[i]
				}

				// 有效主机范围
				firstHost := make(net.IP, len(network))
				lastHost := make(net.IP, len(network))
				copy(firstHost, network)
				copy(lastHost, broadcast)
				firstHost[len(firstHost)-1]++ // 第一有效地址
				lastHost[len(lastHost)-1]--   // 最后一有效地址

				table.Append([]string{
					iface.Name,
					network.String(),
					mask.String(),
					fmt.Sprintf("%s/%d", network, maskSize),
					broadcast.String(),
					fmt.Sprintf("%s - %s", firstHost, lastHost),
				})
				//recorded = true
				break // 只记录第一条
			}
		}
		// 渲染表格
		table.Render()
	},
}
