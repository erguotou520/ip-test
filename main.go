package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/go-ping/ping"
)

func main() {
	// 读取ips.txt文件
	file, err := os.Open("ips.txt")
	if err != nil {
		fmt.Println("无法打开文件:", err)
		return
	}
	defer file.Close()

	// 创建一个等待组，用于等待所有协程完成
	var wg sync.WaitGroup

	// 创建两个通道，用于存储可ping通和不可通的IP
	pingableIPs := make(chan string)
	unpingableIPs := make(chan string)

	// 关闭通道，以便接收协程知道没有更多的数据发送
	defer func() {
		close(pingableIPs)
		close(unpingableIPs)
	}()

	go func() {
		for ip := range pingableIPs {
			fmt.Printf("机器 %s 可ping通\n", ip)
		}
	}()
	go func() {
		for ip := range unpingableIPs {
			fmt.Printf("机器 %s 无法ping通\n", ip)
		}
	}()

	// 逐行读取文件并启动协程进行ping测试
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ip := scanner.Text()
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			if pingTest(ip) {
				pingableIPs <- ip
			} else {
				unpingableIPs <- ip
			}
		}(ip)
	}

	wg.Wait()
	time.Sleep(1 * time.Second)
	fmt.Println("测试完成，按回车键退出")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}

// pingTest函数用于测试IP是否可达
func pingTest(ip string) bool {
	pinger, err := ping.NewPinger(ip)
	if err != nil {
		panic(err)
	}
	if runtime.GOOS == "windows" {
		pinger.SetPrivileged(true)
	}
	pinger.Count = 3
	pinger.SetLogger(nil)
	timeout := false
	go func() {
		<-time.After(5 * time.Second)
		pinger.Stop()
		timeout = true
	}()
	err = pinger.Run()
	// timeout
	if timeout {
		return false
	}
	if err != nil {
		return false
	}
	return true

}
