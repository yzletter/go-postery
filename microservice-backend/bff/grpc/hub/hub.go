package hub

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/yzletter/go-postery/microservice-backend/bff/config"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

type ETCDServiceHub struct {
	heartbeatFrequency int64
	prefix             string
	client             *etcdv3.Client
	loadBalancer       LoadBalancer
}

func NewEtcdServiceHub(config config.ServiceHubConfig, client *etcdv3.Client, loadBalancer LoadBalancer) *ETCDServiceHub {
	return &ETCDServiceHub{
		heartbeatFrequency: int64(config.HeartbeatFrequency),
		prefix:             config.ServiceRegisterPrefix,
		client:             client,
		loadBalancer:       loadBalancer,
	}
}

func (hub *ETCDServiceHub) Register(ctx context.Context, service string, endpoint string, leaseID int64) (int64, error) {
	if leaseID > 0 {
		if _, err := hub.client.KeepAliveOnce(ctx, etcdv3.LeaseID(leaseID)); err != nil {
			if errors.Is(err, rpctypes.ErrLeaseNotFound) {
				return hub.Register(ctx, service, endpoint, 0)
			}
			slog.Error("Register Service Failed", "error", err, "leaseID", leaseID)
			return 0, err
		}
		return leaseID, nil
	}

	resp, err := hub.client.Grant(ctx, hub.heartbeatFrequency)
	if err != nil {
		slog.Error("Register Service Failed", "error", err)
		return 0, err
	}

	key := hub.prefix + "/" + service + "/" + endpoint
	if _, err := hub.client.Put(ctx, key, "", etcdv3.WithLease(resp.ID)); err != nil {
		slog.Error("etcd Put Service Key Failed", "error", err, "Key", key)
		return 0, err
	}

	return int64(resp.ID), nil
}

func (hub *ETCDServiceHub) Unregister(ctx context.Context, service string, endpoint string) error {
	key := hub.prefix + "/" + service + "/" + endpoint
	if _, err := hub.client.Delete(ctx, key); err != nil {
		slog.Error("Unregister Failed", "error", err)
		return err
	}
	return nil
}

func (hub *ETCDServiceHub) GetServiceEndpoints(ctx context.Context, service string) []string {
	keyPrefix := hub.prefix + "/" + service + "/"
	resp, err := hub.client.Get(ctx, keyPrefix, etcdv3.WithPrefix())
	if err != nil {
		slog.Error("Get Service Endpoints Failed", "error", err)
		return nil
	}

	endpoints := make([]string, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		segments := strings.Split(string(kv.Key), "/")
		endpoints[i] = segments[len(segments)-1]
	}

	return endpoints
}

func (hub *ETCDServiceHub) GetServiceEndpoint(ctx context.Context, service string) string {
	return hub.loadBalancer.Take(hub.GetServiceEndpoints(ctx, service))
}
