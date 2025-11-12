# How to run (simple)

## 1) Build once
```bash
go build -o bin/node ./cmd/node
```

## 2) Start three nodes (one terminal per node) BE AWARE OF ERROR IN LOGS UNTIL ALL ARE CREATED!

**Terminal 1**
```bash
./bin/node --id n1 --addr 127.0.0.1:5001 \
  --peers 127.0.0.1:5001,127.0.0.1:5002,127.0.0.1:5003 \
  --logdir logs
```

**Terminal 2**
```bash
./bin/node --id n2 --addr 127.0.0.1:5002 \
  --peers 127.0.0.1:5001,127.0.0.1:5002,127.0.0.1:5003 \
  --logdir logs
```

**Terminal 3**
```bash
./bin/node --id n3 --addr 127.0.0.1:5003 \
  --peers 127.0.0.1:5001,127.0.0.1:5002,127.0.0.1:5003 \
  --logdir logs
```

## Optional: follow logs live
```bash
tail -F logs/n1.log logs/n2.log logs/n3.log
```

---
**Note:** Entrypoint is `cmd/node/main.go` and core logic lives in `pkg/ra/`.
