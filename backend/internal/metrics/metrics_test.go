package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetrics(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
	}{
		{
			name:      "default namespace",
			namespace:  "opamp",
		},
		{
			name:      "custom namespace",
			namespace: "test_service",
		},
		{
			name:      "empty namespace",
			namespace: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMetrics(tt.namespace)

			require.NotNil(t, m)
			assert.NotNil(t, m.HTTPRequestsTotal)
			assert.NotNil(t, m.HTTPRequestDuration)
			assert.NotNil(t, m.HTTPRequestSize)
			assert.NotNil(t, m.HTTPResponseSize)

			assert.NotNil(t, m.AgentsTotal)
			assert.NotNil(t, m.AgentsConnected)
			assert.NotNil(t, m.AgentsDisconnected)
			assert.NotNil(t, m.AgentConnectTotal)
			assert.NotNil(t, m.AgentDisconnectTotal)

			assert.NotNil(t, m.ConfigurationsTotal)
			assert.NotNil(t, m.ConfigurationChangesTotal)
			assert.NotNil(t, m.ConfigurationPushTotal)

			assert.NotNil(t, m.DBConnectionsOpen)
			assert.NotNil(t, m.DBConnectionsIdle)
			assert.NotNil(t, m.DBQueriesTotal)
			assert.NotNil(t, m.DBQueryDuration)
		})
	}
}

func TestMetrics_HTTPMetrics(t *testing.T) {
	// 创建自定义 registry 避免全局污染
	registry := prometheus.NewRegistry()

	// 手动创建 metrics (因为 promauto 会使用全局registry)
	m := &Metrics{
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "test",
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "test",
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
			},
			[]string{"method", "path"},
		),
	}

	registry.MustRegister(m.HTTPRequestsTotal)
	registry.MustRegister(m.HTTPRequestDuration)

	t.Run("increment HTTP requests counter", func(t *testing.T) {
		// 记录请求
		m.HTTPRequestsTotal.WithLabelValues("GET", "/api/agents", "200").Inc()
		m.HTTPRequestsTotal.WithLabelValues("GET", "/api/agents", "200").Inc()
		m.HTTPRequestsTotal.WithLabelValues("POST", "/api/agents", "201").Inc()

		// 验证计数器
		metrics, err := registry.Gather()
		require.NoError(t, err)

		var found bool
		for _, mf := range metrics {
			if mf.GetName() == "test_http_requests_total" {
				found = true
				assert.Equal(t, dto.MetricType_COUNTER, mf.GetType())

				// 验证指标数量
				require.Len(t, mf.GetMetric(), 2)
			}
		}
		assert.True(t, found, "http_requests_total metric not found")
	})

	t.Run("observe HTTP request duration", func(t *testing.T) {
		// 记录请求耗时
		m.HTTPRequestDuration.WithLabelValues("GET", "/api/agents").Observe(0.123)
		m.HTTPRequestDuration.WithLabelValues("GET", "/api/agents").Observe(0.456)
		m.HTTPRequestDuration.WithLabelValues("POST", "/api/configurations").Observe(0.789)

		// 验证直方图
		metrics, err := registry.Gather()
		require.NoError(t, err)

		var found bool
		for _, mf := range metrics {
			if mf.GetName() == "test_http_request_duration_seconds" {
				found = true
				assert.Equal(t, dto.MetricType_HISTOGRAM, mf.GetType())

				// 验证有2个不同的标签组合
				require.Len(t, mf.GetMetric(), 2)
			}
		}
		assert.True(t, found, "http_request_duration_seconds metric not found")
	})
}

func TestMetrics_AgentMetrics(t *testing.T) {
	registry := prometheus.NewRegistry()

	m := &Metrics{
		AgentsTotal: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "test",
				Name:      "agents_total",
				Help:      "Total number of agents",
			},
		),
		AgentsConnected: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "test",
				Name:      "agents_connected",
				Help:      "Number of connected agents",
			},
		),
		AgentConnectTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: "test",
				Name:      "agent_connect_total",
				Help:      "Total number of agent connections",
			},
		),
	}

	registry.MustRegister(m.AgentsTotal)
	registry.MustRegister(m.AgentsConnected)
	registry.MustRegister(m.AgentConnectTotal)

	t.Run("set agent gauges", func(t *testing.T) {
		// 设置 gauge 值
		m.AgentsTotal.Set(100)
		m.AgentsConnected.Set(75)

		// 验证
		metrics, err := registry.Gather()
		require.NoError(t, err)

		for _, mf := range metrics {
			if mf.GetName() == "test_agents_total" {
				assert.Equal(t, dto.MetricType_GAUGE, mf.GetType())
				assert.Equal(t, 100.0, mf.GetMetric()[0].GetGauge().GetValue())
			}
			if mf.GetName() == "test_agents_connected" {
				assert.Equal(t, dto.MetricType_GAUGE, mf.GetType())
				assert.Equal(t, 75.0, mf.GetMetric()[0].GetGauge().GetValue())
			}
		}
	})

	t.Run("increment agent connect counter", func(t *testing.T) {
		// 增加连接计数
		m.AgentConnectTotal.Inc()
		m.AgentConnectTotal.Inc()
		m.AgentConnectTotal.Inc()

		// 验证
		metrics, err := registry.Gather()
		require.NoError(t, err)

		for _, mf := range metrics {
			if mf.GetName() == "test_agent_connect_total" {
				assert.Equal(t, dto.MetricType_COUNTER, mf.GetType())
				assert.Equal(t, 3.0, mf.GetMetric()[0].GetCounter().GetValue())
			}
		}
	})

	t.Run("inc and dec gauges", func(t *testing.T) {
		// 重置
		m.AgentsConnected.Set(10)

		// 增加
		m.AgentsConnected.Inc()
		m.AgentsConnected.Inc()

		// 减少
		m.AgentsConnected.Dec()

		// 验证
		metrics, err := registry.Gather()
		require.NoError(t, err)

		for _, mf := range metrics {
			if mf.GetName() == "test_agents_connected" {
				assert.Equal(t, 11.0, mf.GetMetric()[0].GetGauge().GetValue())
			}
		}
	})
}

func TestMetrics_ConfigurationMetrics(t *testing.T) {
	registry := prometheus.NewRegistry()

	m := &Metrics{
		ConfigurationsTotal: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "test",
				Name:      "configurations_total",
				Help:      "Total number of configurations",
			},
		),
		ConfigurationChangesTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: "test",
				Name:      "configuration_changes_total",
				Help:      "Total number of configuration changes",
			},
		),
		ConfigurationPushTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "test",
				Name:      "configuration_push_total",
				Help:      "Total number of configuration pushes to agents",
			},
			[]string{"status"},
		),
	}

	registry.MustRegister(m.ConfigurationsTotal)
	registry.MustRegister(m.ConfigurationChangesTotal)
	registry.MustRegister(m.ConfigurationPushTotal)

	t.Run("configuration metrics", func(t *testing.T) {
		// 设置配置数量
		m.ConfigurationsTotal.Set(50)

		// 记录配置变更
		m.ConfigurationChangesTotal.Inc()
		m.ConfigurationChangesTotal.Inc()

		// 记录配置推送(成功和失败)
		m.ConfigurationPushTotal.WithLabelValues("success").Inc()
		m.ConfigurationPushTotal.WithLabelValues("success").Inc()
		m.ConfigurationPushTotal.WithLabelValues("success").Inc()
		m.ConfigurationPushTotal.WithLabelValues("failure").Inc()

		// 验证
		metrics, err := registry.Gather()
		require.NoError(t, err)

		for _, mf := range metrics {
			switch mf.GetName() {
			case "test_configurations_total":
				assert.Equal(t, 50.0, mf.GetMetric()[0].GetGauge().GetValue())
			case "test_configuration_changes_total":
				assert.Equal(t, 2.0, mf.GetMetric()[0].GetCounter().GetValue())
			case "test_configuration_push_total":
				// 应该有2个标签: success 和 failure
				require.Len(t, mf.GetMetric(), 2)
			}
		}
	})
}

func TestMetrics_DatabaseMetrics(t *testing.T) {
	registry := prometheus.NewRegistry()

	m := &Metrics{
		DBConnectionsOpen: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "test",
				Name:      "db_connections_open",
				Help:      "Number of open database connections",
			},
		),
		DBConnectionsIdle: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "test",
				Name:      "db_connections_idle",
				Help:      "Number of idle database connections",
			},
		),
		DBQueriesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "test",
				Name:      "db_queries_total",
				Help:      "Total number of database queries",
			},
			[]string{"operation"},
		),
		DBQueryDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "test",
				Name:      "db_query_duration_seconds",
				Help:      "Database query duration in seconds",
			},
			[]string{"operation"},
		),
	}

	registry.MustRegister(m.DBConnectionsOpen)
	registry.MustRegister(m.DBConnectionsIdle)
	registry.MustRegister(m.DBQueriesTotal)
	registry.MustRegister(m.DBQueryDuration)

	t.Run("database connection metrics", func(t *testing.T) {
		// 设置连接池状态
		m.DBConnectionsOpen.Set(10)
		m.DBConnectionsIdle.Set(3)

		// 验证
		metrics, err := registry.Gather()
		require.NoError(t, err)

		for _, mf := range metrics {
			if mf.GetName() == "test_db_connections_open" {
				assert.Equal(t, 10.0, mf.GetMetric()[0].GetGauge().GetValue())
			}
			if mf.GetName() == "test_db_connections_idle" {
				assert.Equal(t, 3.0, mf.GetMetric()[0].GetGauge().GetValue())
			}
		}
	})

	t.Run("database query metrics", func(t *testing.T) {
		// 记录查询
		m.DBQueriesTotal.WithLabelValues("SELECT").Inc()
		m.DBQueriesTotal.WithLabelValues("SELECT").Inc()
		m.DBQueriesTotal.WithLabelValues("INSERT").Inc()
		m.DBQueriesTotal.WithLabelValues("UPDATE").Inc()

		// 记录查询耗时
		m.DBQueryDuration.WithLabelValues("SELECT").Observe(0.050)
		m.DBQueryDuration.WithLabelValues("INSERT").Observe(0.100)

		// 验证
		metrics, err := registry.Gather()
		require.NoError(t, err)

		for _, mf := range metrics {
			if mf.GetName() == "test_db_queries_total" {
				// 应该有3种不同的操作类型
				require.Len(t, mf.GetMetric(), 3)
			}
			if mf.GetName() == "test_db_query_duration_seconds" {
				// 应该有2种操作类型
				require.Len(t, mf.GetMetric(), 2)
			}
		}
	})
}

func TestMetrics_AllMetricsRegistered(t *testing.T) {
	t.Run("all metrics are created", func(t *testing.T) {
		m := NewMetrics("opamp_test")

		// HTTP Metrics
		assert.NotNil(t, m.HTTPRequestsTotal, "HTTPRequestsTotal should not be nil")
		assert.NotNil(t, m.HTTPRequestDuration, "HTTPRequestDuration should not be nil")
		assert.NotNil(t, m.HTTPRequestSize, "HTTPRequestSize should not be nil")
		assert.NotNil(t, m.HTTPResponseSize, "HTTPResponseSize should not be nil")

		// Agent Metrics
		assert.NotNil(t, m.AgentsTotal, "AgentsTotal should not be nil")
		assert.NotNil(t, m.AgentsConnected, "AgentsConnected should not be nil")
		assert.NotNil(t, m.AgentsDisconnected, "AgentsDisconnected should not be nil")
		assert.NotNil(t, m.AgentConnectTotal, "AgentConnectTotal should not be nil")
		assert.NotNil(t, m.AgentDisconnectTotal, "AgentDisconnectTotal should not be nil")

		// Configuration Metrics
		assert.NotNil(t, m.ConfigurationsTotal, "ConfigurationsTotal should not be nil")
		assert.NotNil(t, m.ConfigurationChangesTotal, "ConfigurationChangesTotal should not be nil")
		assert.NotNil(t, m.ConfigurationPushTotal, "ConfigurationPushTotal should not be nil")

		// Database Metrics
		assert.NotNil(t, m.DBConnectionsOpen, "DBConnectionsOpen should not be nil")
		assert.NotNil(t, m.DBConnectionsIdle, "DBConnectionsIdle should not be nil")
		assert.NotNil(t, m.DBQueriesTotal, "DBQueriesTotal should not be nil")
		assert.NotNil(t, m.DBQueryDuration, "DBQueryDuration should not be nil")
	})
}
