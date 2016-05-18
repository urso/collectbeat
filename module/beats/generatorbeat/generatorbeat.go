package generatorbeat

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/metricbeat/mb"
)

type metricSet struct {
	mb.BaseMetricSet
	url    string
	client *http.Client // HTTP client that is reused across requests.

	stats map[string]float64
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
		stats:         map[string]float64{},
	}, nil
}

func (m *metricSet) Fetch() (common.MapStr, error) {
	resp, err := m.client.Get(m.url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	stats := struct {
		LogstashSuccess int64 `json:"libbeatLogstashPublishedAndAckedEvents"`
		LogstashFail    int64 `json:"libbeatLogstashPublishedButNotAckedEvents"`
		KafkaSuccess    int64 `json:"libbeatKafkaPublishedAndAckedEvents"`
		KafkaFail       int64 `json:"libbeatKafkaPublishedButNotAckedEvents"`
		ESSuccess       int64 `json:"libbeatEsPublishedAndAckedEvents"`
		ESFail          int64 `json:"libbeatEsPublishedButNotAckedEvents"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&stats)
	if err != nil {
		return nil, err
	}

	event := common.MapStr{
		"hostname": m.Host(),
		"logstash": common.MapStr{
			"success": stats.LogstashSuccess,
			"fail":    stats.LogstashFail,
		},
		"kafka": common.MapStr{
			"success": stats.KafkaSuccess,
			"fail":    stats.KafkaFail,
		},
		"es": common.MapStr{
			"success": stats.ESSuccess,
			"fail":    stats.ESFail,
		},
	}
	return event, nil
}
