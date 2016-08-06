package QQLocation

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

const data = "ip-utf8.dat"
const testIP = "59.78.23.18"
const testIPError = "::1"

func TestSearch(t *testing.T) {
	parser, err := Init(data)
	if err != nil {
		t.Fatal("Init failed:", err)
	}

	info, err := parser.Search(testIP)
	if err != nil {
		t.Fatal("Search failed:", err)
	}

	if info.Continent != "亚洲" {
		t.Fatal("country expect = 亚洲, but actual =", info.Continent)
	}

	if info.Country != "中国" {
		t.Fatal("country expect = 中国, but actual =", info.Country)
	}

	if info.Province != "上海" {
		t.Fatal("region expect = 上海, but actual =", info.Province)
	}

	if info.City != "上海" {
		t.Fatal("city expect = 上海, but actual =", info.City)
	}

	if info.Isp != "教育网" {
		t.Fatal("isp expect = 教育网, but actual =", info.Isp)
	}
}

func BenchmarkSearch(b *testing.B) {
	b.StopTimer()
	parser, err := Init(data)
	if err != nil {
		b.Fatal("Init failed:", err)
	}
	rand.Seed(time.Now().Unix())
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		ipStr := fmt.Sprintf("%d.%d.%d.%d", rand.Uint32()%255, rand.Uint32()%255, rand.Uint32()%255, rand.Uint32()%255)
		if info, err := parser.Search(ipStr); err != nil || info == nil {
			b.Fatal("Failed to search IP" + ipStr)
		}
	}
}
