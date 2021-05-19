package geodata_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/platform/filesystem"
	"github.com/v2fly/v2ray-core/v4/infra/conf/geodata"
	_ "github.com/v2fly/v2ray-core/v4/infra/conf/geodata/memconservative"
	_ "github.com/v2fly/v2ray-core/v4/infra/conf/geodata/standard"
)

func init() {
	const (
		geoipURL   = "https://raw.githubusercontent.com/v2fly/geoip/release/geoip.dat"
		geositeURL = "https://raw.githubusercontent.com/v2fly/domain-list-community/release/dlc.dat"
	)

	wd, err := os.Getwd()
	common.Must(err)

	tempPath := filepath.Join(wd, "..", "..", "..", "testing", "temp")
	geoipPath := filepath.Join(tempPath, "geoip.dat")
	geositePath := filepath.Join(tempPath, "geosite.dat")

	os.Setenv("v2ray.location.asset", tempPath)

	if _, err := os.Stat(geoipPath); err != nil && errors.Is(err, fs.ErrNotExist) {
		common.Must(os.MkdirAll(tempPath, 0o755))
		geoipBytes, err := common.FetchHTTPContent(geoipURL)
		common.Must(err)
		common.Must(filesystem.WriteFile(geoipPath, geoipBytes))
	}
	if _, err := os.Stat(geositePath); err != nil && errors.Is(err, fs.ErrNotExist) {
		common.Must(os.MkdirAll(tempPath, 0o755))
		geositeBytes, err := common.FetchHTTPContent(geositeURL)
		common.Must(err)
		common.Must(filesystem.WriteFile(geositePath, geositeBytes))
	}
}

func BenchmarkStandardLoaderGeoIP(b *testing.B) {
	standardLoader, err := geodata.GetGeoDataLoader("standard")
	common.Must(err)

	m1 := runtime.MemStats{}
	m2 := runtime.MemStats{}
	runtime.ReadMemStats(&m1)
	standardLoader.LoadGeoIP("cn")
	standardLoader.LoadGeoIP("us")
	standardLoader.LoadGeoIP("private")
	runtime.ReadMemStats(&m2)

	b.ReportMetric(float64(m2.Alloc-m1.Alloc)/1024/1024, "MiB(GeoIP-Alloc)")
	b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/1024/1024, "MiB(GeoIP-TotalAlloc)")
}

func BenchmarkStandardLoaderGeoSite(b *testing.B) {
	standardLoader, err := geodata.GetGeoDataLoader("standard")
	common.Must(err)

	m3 := runtime.MemStats{}
	m4 := runtime.MemStats{}
	runtime.ReadMemStats(&m3)
	standardLoader.LoadGeoSite("cn")
	standardLoader.LoadGeoSite("geolocation-!cn")
	standardLoader.LoadGeoSite("private")
	runtime.ReadMemStats(&m4)

	b.ReportMetric(float64(m4.Alloc-m3.Alloc)/1024/1024, "MiB(GeoSite-Alloc)")
	b.ReportMetric(float64(m4.TotalAlloc-m3.TotalAlloc)/1024/1024, "MiB(GeoSite-TotalAlloc)")
}

func BenchmarkMemconservativeLoaderGeoIP(b *testing.B) {
	standardLoader, err := geodata.GetGeoDataLoader("memconservative")
	common.Must(err)

	m1 := runtime.MemStats{}
	m2 := runtime.MemStats{}
	runtime.ReadMemStats(&m1)
	standardLoader.LoadGeoIP("cn")
	standardLoader.LoadGeoIP("us")
	standardLoader.LoadGeoIP("private")
	runtime.ReadMemStats(&m2)

	b.ReportMetric(float64(m2.Alloc-m1.Alloc)/1024, "KiB(GeoIP-Alloc)")
	b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/1024/1024, "MiB(GeoIP-TotalAlloc)")
}

func BenchmarkMemconservativeLoaderGeoSite(b *testing.B) {
	standardLoader, err := geodata.GetGeoDataLoader("memconservative")
	common.Must(err)

	m3 := runtime.MemStats{}
	m4 := runtime.MemStats{}
	runtime.ReadMemStats(&m3)
	standardLoader.LoadGeoSite("cn")
	standardLoader.LoadGeoSite("geolocation-!cn")
	standardLoader.LoadGeoSite("private")
	runtime.ReadMemStats(&m4)

	b.ReportMetric(float64(m4.Alloc-m3.Alloc)/1024, "KiB(GeoSite-Alloc)")
	b.ReportMetric(float64(m4.TotalAlloc-m3.TotalAlloc)/1024/1024, "MiB(GeoSite-TotalAlloc)")
}

func BenchmarkAllLoader(b *testing.B) {
	type testingProfileForLoader struct {
		name string
	}
	testCase := []testingProfileForLoader{
		{"standard"},
		{"memconservative"},
	}
	for _, v := range testCase {
		b.Run(v.name, func(b *testing.B) {
			b.Run("Geosite", func(b *testing.B) {
				loader, err := geodata.GetGeoDataLoader(v.name)
				if err != nil {
					b.Fatal(err)
				}

				m3 := runtime.MemStats{}
				m4 := runtime.MemStats{}
				runtime.ReadMemStats(&m3)
				loader.LoadGeoSite("cn")
				loader.LoadGeoSite("geolocation-!cn")
				loader.LoadGeoSite("private")
				runtime.ReadMemStats(&m4)

				b.ReportMetric(float64(m4.Alloc-m3.Alloc)/1024, "KiB(GeoSite-Alloc)")
				b.ReportMetric(float64(m4.TotalAlloc-m3.TotalAlloc)/1024/1024, "MiB(GeoSite-TotalAlloc)")
			})

			b.Run("GeoIP", func(b *testing.B) {
				loader, err := geodata.GetGeoDataLoader(v.name)
				if err != nil {
					b.Fatal(err)
				}

				m1 := runtime.MemStats{}
				m2 := runtime.MemStats{}
				runtime.ReadMemStats(&m1)
				loader.LoadGeoIP("cn")
				loader.LoadGeoIP("us")
				loader.LoadGeoIP("private")
				runtime.ReadMemStats(&m2)

				b.ReportMetric(float64(m2.Alloc-m1.Alloc)/1024/1024, "MiB(GeoIP-Alloc)")
				b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/1024/1024, "MiB(GeoIP-TotalAlloc)")
			})
		})
	}
}
