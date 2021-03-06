package generatorbeat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/metricbeat/mb"
)

type metricSet struct {
	mb.BaseMetricSet
	url    string
	client *http.Client // HTTP client that is reused across requests.

	statsLast time.Time
	stats     beatStats
}

type beatStats struct {
	LogstashSuccess int64 `json:"libbeatLogstashPublishedAndAckedEvents"`
	LogstashFail    int64 `json:"libbeatLogstashPublishedButNotAckedEvents"`
	KafkaSuccess    int64 `json:"libbeatKafkaPublishedAndAckedEvents"`
	KafkaFail       int64 `json:"libbeatKafkaPublishedButNotAckedEvents"`
	ESSuccess       int64 `json:"libbeatEsPublishedAndAckedEvents"`
	ESFail          int64 `json:"libbeatEsPublishedButNotAckedEvents"`
	TotalSuccess    int64
	TotalFail       int64
}

func init() {
	if err := mb.Registry.AddMetricSet("beats", "generatorbeat", New); err != nil {
		panic(err)
	}
}

func New(base mb.BaseMetricSet) (mb.MetricSet, error) {
	config := struct {
		Port uint16
	}{
		Port: 6060,
	}
	if err := base.Module().UnpackConfig(&config); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("http://%v:%v/debug/vars", base.Host(), config.Port)

	return &metricSet{
		BaseMetricSet: base,
		url:           url,
		client:        &http.Client{Timeout: base.Module().Config().Timeout},
	}, nil
}

func (m *metricSet) reset() {
	m.statsLast = time.Time{}
}

func (m *metricSet) Fetch() (common.MapStr, error) {
	resp, err := m.client.Get(m.url)
	if err != nil {
		m.reset()
		return nil, err
	}
	now := time.Now()
	defer resp.Body.Close()

	stats := beatStats{}
	err = json.NewDecoder(resp.Body).Decode(&stats)
	if err != nil {
		m.reset()
		return nil, err
	}
	stats.TotalSuccess = stats.LogstashSuccess + stats.KafkaSuccess + stats.ESSuccess
	stats.TotalFail = stats.LogstashFail + stats.KafkaFail + stats.ESFail

	var event common.MapStr
	if !m.statsLast.IsZero() {
		dt := now.Sub(m.statsLast).Seconds()
		old := &m.stats
		delta := func(v, old int64) float64 {
			return float64(v-old) / dt
		}

		event = common.MapStr{
			"hostname":   m.Host(),
			"@timestamp": common.Time(now),
			"logstash": common.MapStr{
				"success": delta(stats.LogstashSuccess, old.LogstashSuccess),
				"fail":    delta(stats.LogstashFail, old.LogstashFail),
			},
			"kafka": common.MapStr{
				"success": delta(stats.KafkaSuccess, old.KafkaSuccess),
				"fail":    delta(stats.KafkaFail, old.KafkaFail),
			},
			"es": common.MapStr{
				"success": delta(stats.ESSuccess, old.ESSuccess),
				"fail":    delta(stats.ESFail, old.ESFail),
			},
			"total": common.MapStr{
				"success": delta(stats.TotalSuccess, old.TotalSuccess),
				"fail": delta(stats.TotalFail, old.TotalFail),
			},
		}
	}

	m.statsLast = now
	m.stats = stats

	return event, nil
}
