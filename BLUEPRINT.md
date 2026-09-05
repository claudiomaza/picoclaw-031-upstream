# BLUEPRINT — Subagentes durables y confiables

Fecha: 2026-09-03
Estado: plan aprobado para implementación posterior
Proyecto: PicoClaw / SubagentManager

## 1. Discovery

Estado observado:

- `SubagentManager` mantiene tareas en `map[string]*SubagentTask`, solo en memoria.
- `spawn` devuelve `task_id` y ejecuta subturnos en paralelo.
- `spawn_status` puede consultar tareas y resultados mientras vive el proceso.
- `completed` está demostrado; `collected` y `delivered` necesitan estados explícitos.
- Un fallo del LLM del padre puede impedir el informe final aunque los hijos hayan terminado.
- Los resultados se pierden tras reinicio.
- El proyecto real confirmó que los tests deben resolver rutas desde `__file__`, no desde el `cwd` implícito.
- Restricciones: VM de ~1 GB RAM; evitar dependencias pesadas y builds paralelos.

## 2. Objetivo y no objetivos

Objetivo: que una tarea completada sobreviva a fallos del proveedor, del coordinador y a reinicios, y pueda ser recolectada y entregada sin duplicación.

No objetivos de esta fase:

- sistema distribuido multi-nodo;
- Ruflo, Node, MCP o broker externo;
- reanudar una llamada LLM interrumpida desde tokens parciales;
- ejecutar automáticamente acciones irreversibles.

## 3. Arquitectura propuesta

### 3.1 TaskStore SQLite local

Agregar una capa pequeña `TaskStore` con SQLite WAL, detrás de una interfaz para tests.

Tablas principales:

```sql
subagent_tasks(
  task_id TEXT PRIMARY KEY,
  parent_id TEXT NOT NULL,
  agent_id TEXT NOT NULL,
  label TEXT NOT NULL,
  task TEXT NOT NULL,
  input_hash TEXT NOT NULL,
  status TEXT NOT NULL,
  result TEXT,
  error TEXT,
  attempt INTEGER NOT NULL DEFAULT 0,
  lease_until TEXT,
  created_at TEXT NOT NULL,
  started_at TEXT,
  completed_at TEXT,
  collected_at TEXT,
  delivered_at TEXT,
  updated_at TEXT NOT NULL
);

subagent_events(
  event_id TEXT PRIMARY KEY,
  task_id TEXT NOT NULL,
  kind TEXT NOT NULL,
  payload TEXT NOT NULL,
  created_at TEXT NOT NULL
);
```

Índices: `(parent_id,status)`, `(status,lease_until)`, `input_hash`.

### 3.2 Máquina de estados

Estados válidos:

```text
queued → running → completed → collected → delivered
queued/running → failed
running → expired → queued|failed
completed → orphaned
```

Toda transición se valida y se registra como evento. `completed` se confirma únicamente después de guardar el resultado en SQLite.

### 3.3 Supervisor local

Un supervisor de una sola goroutine, con ticker configurable, debe:

1. reclamar tareas `queued` respetando `max_concurrent`;
2. detectar leases vencidos;
3. persistir completados antes de entregarlos;
4. reintentar entrega al padre;
5. recuperar resultados de padres que volvieron a estar activos;
6. marcar `orphaned` solo cuando no haya padre recuperable;
7. detenerse limpiamente durante shutdown.

No debe llamar al LLM.

### 3.4 API del manager

Conservar `Spawn` y agregar:

```go
GetTask(taskID string) (TaskSnapshot, error)
ListTasks(filter TaskFilter) ([]TaskSnapshot, error)
Collect(ctx context.Context, taskID string) (TaskSnapshot, error)
MarkDelivered(taskID, deliveryID string) error
Recover(ctx context.Context, parentID string) error
Retry(taskID string) error
```

`TaskSnapshot` debe ser una copia inmutable para evitar carreras.

### 3.5 Idempotencia y leases

Crear `input_hash` estable a partir de padre, agente, label y tarea.

- Si existe una tarea terminal con el mismo hash, reutilizarla.
- No duplicar una tarea `running` con lease vigente.
- Renovar lease durante ejecución.
- Al vencer, permitir un nuevo intento limitado.
- `max_attempts` y timeouts deben ser configurables por tarea.

### 3.6 Entrega confirmada

Separar claramente:

```text
completed  = resultado persistido
collected  = resultado retirado por el coordinador/supervisor
 delivered = padre aceptó la entrega
```

Cada entrega tendrá `delivery_id` idempotente. Un retry de entrega no debe duplicar el contenido en el padre.

## 4. Tratamiento de fallos

Clasificar errores antes de aplicar política:

- transporte/502/timeout: retry con backoff;
- 429/cuota: backoff respetando límites y cambio de proveedor;
- fallo de tarea: `failed`, sin retry ciego;
- validación: `failed` o nueva tarea de corrección;
- crash: recovery desde SQLite;
- entrega: conservar `completed` y reintentar.

El circuit breaker del proveedor debe pausar nuevos LLM calls sin borrar ni fallar trabajos ya completados.

## 5. Verificación de tareas críticas

Agregar `VerificationPolicy`:

- archivos esperados y diff permitido;
- comando de tests con `cwd` explícito;
- timeout;
- artefactos y hashes;
- criterios de aceptación.

El estado de negocio solo puede pasar a `verified` después de la verificación local. Para deploy, mensajes, borrados o credenciales: propuesta → validación → aprobación explícita → ejecución → verificación posterior.

## 6. Plan TDD

### Fase Red

Escribir primero tests que fallen para:

1. persistencia de `queued/running/completed`;
2. recuperación después de cerrar/reabrir SQLite;
3. transiciones inválidas;
4. idempotencia de `Spawn`;
5. lease vencido y límite de intentos;
6. `Collect` exactamente una vez;
7. `MarkDelivered` idempotente;
8. entrega cuando el padre desaparece;
9. retry sin duplicar resultado;
10. recuperación después de un 502 simulado;
11. concurrencia con `-race`;
12. shutdown sin perder resultados.

### Fase Green

Implementar el mínimo en este orden:

1. interfaz `TaskStore` y SQLite;
2. migración inicial y WAL;
3. persistencia en `Spawn` y completion;
4. estados y snapshots;
5. `Collect`/`MarkDelivered`;
6. supervisor y leases;
7. idempotencia;
8. recovery al iniciar;
9. políticas de retry/circuit breaker;
10. verificador de artefactos.

### Fase Refactor

- separar manager, store, supervisor y delivery;
- eliminar accesos directos al map salvo cache derivada;
- centralizar transiciones;
- mantener compatibilidad con `spawn_status`;
- documentar métricas y recuperación.

## 7. Verificación empírica

Ejecutar, sin paralelizar builds:

```bash
go test ./pkg/tools ./pkg/agent ./pkg/health
 go test -race ./pkg/tools ./pkg/agent
```

Escenarios obligatorios:

- 3 y 5 tareas simultáneas;
- padre con 502 durante el informe;
- muerte y reinicio del proceso;
- proveedor caído durante spawn;
- padre terminado antes de delivery;
- duplicación del mismo request;
- disco SQLite reabierto;
- tarea que modifica un archivo y falla QA.

Cada escenario debe producir evidencia de logs y consultar SQLite, no depender del texto del LLM.

## 8. Rollout y rollback

1. backup del binario, configuración y SQLite;
2. feature flag `durable_subagents=false` por defecto;
3. migración reversible y backup de DB;
4. activar primero en prueba local;
5. prueba real de 3 tareas;
6. prueba de reinicio;
7. activar en servicio;
8. observar errores, orphan, retries y latencias;
9. rollback al binario anterior sin borrar la DB.

No eliminar la DB durante rollback: conservarla para diagnóstico y reimportación.

## 9. Criterios de aceptación

La implementación se considera lista cuando:

- ningún `completed` se pierde después de reiniciar;
- `Collect` y `MarkDelivered` tienen evidencia independiente;
- un 502 del padre no pierde resultados de hijos;
- reintentar no duplica tareas ni entregas;
- tareas expiradas se recuperan con límite de intentos;
- tests normales y `-race` pasan;
- una tarea crítica requiere verificación externa antes de ejecución irreversible;
- `spawn_status` conserva compatibilidad y expone estados reales.

## 10. Handoff

La implementación debe comenzar en una rama/worktree separado, manteniendo intacto el binario activo hasta completar Red → Green → Refactor y las pruebas de crash. No cambiar configuración de producción durante la fase de diseño.
