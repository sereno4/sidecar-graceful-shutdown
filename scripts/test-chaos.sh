#!/bin/bash
set -e

NAMESPACE="sidecar-lab"
JOB="dev-batch-job"

echo "═════════════════════════════════════════════════════════════"
echo "  TESTE DE CAOS: SIGTERM durante processamento"
echo "═════════════════════════════════════════════════════════════"

# 1. Deletar job antigo
echo "[1/6] Limpando job anterior..."
kubectl delete job -n $NAMESPACE $JOB 2>/dev/null || true
sleep 3

# 2. Aplicar config com retry
echo "[2/6] Aplicando config com backoffLimit=3..."
kubectl apply -k k8s/overlays/dev/ 2>/dev/null || true
sleep 5

# 3. Pegar nome do pod
POD=$(kubectl get pods -n $NAMESPACE -l job-name=$JOB -o jsonpath='{.items[0].metadata.name}')
echo "[3/6] Pod criado: $POD"

# 4. Aguardar processamento iniciar
echo "[4/6] Aguardando 8s para processamento iniciar..."
sleep 8

# 5. Capturar logs antes do caos
echo "[5/6] Capturando logs antes do SIGTERM..."
kubectl logs -n $NAMESPACE $POD -c batch-processor > /tmp/batch-before.log 2>&1 || true
kubectl logs -n $NAMESPACE $POD -c telemetry-sidecar > /tmp/sidecar-before.log 2>&1 || true

# 6. Deletar pod (simula SIGTERM do K8s)
echo "[6/6] Deletando pod (simula SIGTERM)..."
kubectl delete pod -n $NAMESPACE $POD --grace-period=60

# Aguardar resultado
echo "Aguardando 60s para Job recriar pod..."
sleep 60

# Verificar
echo ""
echo "═════════════════════════════════════════════════════════════"
echo "  RESULTADOS"
echo "═════════════════════════════════════════════════════════════"

echo ""
echo "Status do job:"
kubectl get job -n $NAMESPACE $JOB -o wide

echo ""
echo "Pods:"
kubectl get pods -n $NAMESPACE -l job-name=$JOB -o wide

echo ""
echo "Logs do batch processor (antes):"
cat /tmp/batch-before.log 2>/dev/null | tail -20 || echo "  (nao disponivel)"

echo ""
echo "Logs do sidecar (antes):"
cat /tmp/sidecar-before.log 2>/dev/null | tail -20 || echo "  (nao disponivel)"

echo ""
echo "═════════════════════════════════════════════════════════════"
