#!/bin/bash
set -e

NAMESPACE="sidecar-lab"
JOB="dev-batch-job"

echo "═════════════════════════════════════════════════════════════"
echo "  TESTE: SIGTERM no Container (simula Kubelet)"
echo "═════════════════════════════════════════════════════════════"

# 1. Limpar
echo "[1/5] Limpando..."
kubectl delete job -n $NAMESPACE $JOB 2>/dev/null || true
sleep 3

# 2. Criar job
echo "[2/5] Criando job..."
kubectl apply -k k8s/overlays/dev/ 2>/dev/null || true
sleep 5

# 3. Pegar pod
POD=$(kubectl get pods -n $NAMESPACE -l job-name=$JOB -o jsonpath='{.items[0].metadata.name}')
echo "[3/5] Pod: $POD"

# 4. Aguardar
echo "[4/5] Aguardando 8s..."
sleep 8

# 5. Enviar SIGTERM para o PID 1 do container batch-processor
echo "[5/5] Enviando SIGTERM para batch-processor (PID 1)..."
kubectl exec -n $NAMESPACE $POD -c batch-processor -- kill -TERM 1

# Aguardar
echo "Aguardando 30s..."
sleep 30

# Resultados
echo ""
echo "═════════════════════════════════════════════════════════════"
echo "  RESULTADOS"
echo "═════════════════════════════════════════════════════════════"

echo ""
echo "Logs do batch processor:"
kubectl logs -n $NAMESPACE $POD -c batch-processor 2>/dev/null || echo "  (nao disponivel)"

echo ""
echo "Logs do sidecar:"
kubectl logs -n $NAMESPACE $POD -c telemetry-sidecar 2>/dev/null || echo "  (nao disponivel)"

echo ""
echo "Status do pod:"
kubectl get pod -n $NAMESPACE $POD -o jsonpath='{range .status.containerStatuses[*]}{.name}: {.state}{"\n"}{end}' 2>/dev/null || echo "  (pod nao existe)"

echo ""
echo "═════════════════════════════════════════════════════════════"
