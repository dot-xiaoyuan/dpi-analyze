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
		table.SetHeader([]string{"Interface Name", "IP Address", "Subnet Mask", "Network Range"})

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
			var recorded bool
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

				// 提取 IP、子网掩码和网段
				ip := ipNet.IP.String()
				mask := net.IP(ipNet.Mask).String()
				network := ipNet.String()

				// 添加到表格
				table.Append([]string{iface.Name, ip, mask, network})
				recorded = true
				break // 只记录第一条
			}

			// 如果没有记录任何地址，添加一个空记录以表明该接口无有效地址
			if !recorded {
				table.Append([]string{iface.Name, "N/A", "N/A", "N/A"})
			}
		}

		// 渲染表格
		table.Render()
	},
}
