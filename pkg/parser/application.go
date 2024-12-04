package parser

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
)

type Application struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	SrcPort  string `json:"src_port"`
	DstPort  string `json:"dst_port"`
	Hostname string `json:"hostname"`
	Request  string `json:"request"`
	Dict     string `json:"dict"`
	Category string `json:"category"`
}

// ParseApplications 解析特征数据，返回域名特征列表
func ParseApplications(data []byte) ([]Application, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	var results []Application
	lineNumber := 0
	var category string
	var parseErr error
	var mutex sync.Mutex

	var wg sync.WaitGroup
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineNumber++

		// 跳过空行
		if line == "" {
			continue
		}
		// 判断分类注释
		if strings.HasPrefix(line, "#") {
			if strings.Contains(line, "class") {
				temp := strings.Split(line, " ")
				category = temp[len(temp)-1]
			}
			continue
		}

		wg.Add(1)
		go func(line string, lineNum int, category string) {
			defer wg.Done()

			applications, err := parseLine(line, category)
			if err != nil {
				parseErr = fmt.Errorf("line %d: %w", lineNum, err)
				return
			}

			mutex.Lock()
			results = append(results, applications...)
			mutex.Unlock()
		}(line, lineNumber, category)
	}

	wg.Wait()
	if scanner.Err() != nil {
		return nil, fmt.Errorf("error reading feature file: %w", scanner.Err())
	}
	return results, parseErr
}

func parseLine(line, category string) ([]Application, error) {
	re := regexp.MustCompile(`(\d+) (.+):\[(.+)]`)
	match := re.FindStringSubmatch(line)
	if len(match) < 4 {
		return nil, errors.New("invalid feature format")
	}

	application := Application{
		ID:       match[1],
		Name:     match[2],
		Category: category,
	}

	var applications []Application
	featureList := strings.Split(match[3], ",")
	for _, item := range featureList {
		feature := strings.Split(item, ";")
		if len(feature) != 6 {
			continue
		}

		application.Protocol, application.SrcPort, application.DstPort = feature[0], feature[1], feature[2]
		application.Hostname, application.Request, application.Dict = feature[3], feature[4], feature[5]
		applications = append(applications, application)
	}
	return applications, nil
}
