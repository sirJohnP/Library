package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func compareGrpcVars(t *testing.T, a, b GRPC) {
	t.Helper()
	require.Equal(t, a.Port, b.Port)
	require.Equal(t, a.GatewayPort, b.GatewayPort)
}

func comparePGVars(t *testing.T, a, b PG) {
	t.Helper()
	require.Equal(t, a.Host, b.Host)
	require.Equal(t, a.Port, b.Port)
	require.Equal(t, a.DB, b.DB)
	require.Equal(t, a.User, b.User)
	require.Equal(t, a.Password, b.Password)
	require.Equal(t, a.MaxConn, b.MaxConn)
}

var (
	fields = [][2]string{
		{"GRPC_PORT", "9090"},
		{"GRPC_GATEWAY_PORT", "8080"},
		{"POSTGRES_HOST", "127.0.0.1"},
		{"POSTGRES_PORT", "55432"},
		{"POSTGRES_DB", "library"},
		{"POSTGRES_USER", "library"},
		{"POSTGRES_PASSWORD", "12345678"},
		{"POSTGRES_MAX_CONN", "10"},
	}
)

func TestNewConfig(t *testing.T) {
	config := Config{
		GRPC: GRPC{
			Port:        fields[0][1],
			GatewayPort: fields[1][1],
		},
		PG: PG{
			Host:     fields[2][1],
			Port:     fields[3][1],
			DB:       fields[4][1],
			User:     fields[5][1],
			Password: fields[6][1],
			MaxConn:  fields[7][1],
		},
	}
	for i := range len(fields) {
		t.Setenv(fields[i][0], fields[i][1])
	}
	t.Setenv("OUTBOX_ENABLED", "false")

	result, err := NewConfig()
	require.NoError(t, err)
	compareGrpcVars(t, result.GRPC, config.GRPC)
	comparePGVars(t, result.PG, config.PG)
}
