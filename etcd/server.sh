docker run -p 2379:2379 -p 2380:2380 \
  --name my-etcd \
  quay.io/coreos/etcd:v3.4.14 \
  /usr/local/bin/etcd \
  --advertise-client-urls http://0.0.0.0:2379 \
  --listen-client-urls http://0.0.0.0:2379 \
  --initial-advertise-peer-urls http://0.0.0.0:2380 \
  --listen-peer-urls http://0.0.0.0:2380 \
  --initial-cluster default=http://0.0.0.0:2380 \
  --initial-cluster-token my-cluster-token \
  --initial-cluster-state new
